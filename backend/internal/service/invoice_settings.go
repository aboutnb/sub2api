package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	settingKeyInvoiceIntegrationConfig = "invoice_integration_config"
	InvoiceFeePayerCustomer            = "customer"
	InvoiceFeePayerPlatform            = "platform"
	InvoiceFeePayerUserChoice          = "user_choice"
)

var ErrInvoiceSecretEncryptionKeyNotConfigured = infraerrors.BadRequest(
	"INVOICE_SECRET_ENCRYPTION_KEY_NOT_CONFIGURED",
	"cannot store the invoice client secret without a fixed TOTP_ENCRYPTION_KEY",
)

type InvoiceAdminSettings struct {
	Enabled                bool
	BaseURL                string
	ClientID               string
	ClientSecret           string
	ClientSecretConfigured bool
	TimeoutSeconds         int
	FeePayer               string
}

type invoiceStoredSettings struct {
	Enabled        bool   `json:"enabled"`
	BaseURL        string `json:"base_url"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	FeePayer       string `json:"fee_payer"`
}

// InvoiceSettingsService stores the XZNOAuth credentials outside the public
// system-settings document and resolves config-file values only as a fallback.
type InvoiceSettingsService struct {
	settingRepo             SettingRepository
	encryptor               SecretEncryptor
	fallback                config.InvoiceIntegrationConfig
	encryptionKeyConfigured bool
}

func NewInvoiceSettingsService(
	settingRepo SettingRepository,
	encryptor SecretEncryptor,
	fallback config.InvoiceIntegrationConfig,
	encryptionKeyConfigured bool,
) *InvoiceSettingsService {
	fallback = normalizeInvoiceIntegrationConfig(fallback)
	return &InvoiceSettingsService{
		settingRepo:             settingRepo,
		encryptor:               encryptor,
		fallback:                fallback,
		encryptionKeyConfigured: encryptionKeyConfigured,
	}
}

func (s *InvoiceSettingsService) GetAdminSettings(ctx context.Context) (*InvoiceAdminSettings, error) {
	stored, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return invoiceAdminSettingsFromConfig(s.fallback), nil
	}
	return &InvoiceAdminSettings{
		Enabled:                stored.Enabled,
		BaseURL:                stored.BaseURL,
		ClientID:               stored.ClientID,
		ClientSecretConfigured: stored.ClientSecret != "" || s.fallback.ClientSecret != "",
		TimeoutSeconds:         stored.TimeoutSeconds,
		FeePayer:               stored.FeePayer,
	}, nil
}

// Update replaces the editable fields. An empty ClientSecret preserves the
// stored secret, or the config-file fallback when the UI is being used for the
// first time.
func (s *InvoiceSettingsService) Update(ctx context.Context, in InvoiceAdminSettings) (*InvoiceAdminSettings, error) {
	stored, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		stored = &invoiceStoredSettings{}
	}

	secret := strings.TrimSpace(in.ClientSecret)
	if secret == "" {
		secret = stored.ClientSecret
	} else {
		if !s.encryptionKeyConfigured {
			return nil, ErrInvoiceSecretEncryptionKeyNotConfigured
		}
		secret, err = s.encryptor.Encrypt(secret)
		if err != nil {
			return nil, fmt.Errorf("encrypt invoice client secret: %w", err)
		}
	}

	next := invoiceStoredSettings{
		Enabled:        in.Enabled,
		BaseURL:        strings.TrimSpace(strings.TrimRight(in.BaseURL, "/")),
		ClientID:       strings.TrimSpace(in.ClientID),
		ClientSecret:   secret,
		TimeoutSeconds: in.TimeoutSeconds,
		FeePayer:       strings.ToLower(strings.TrimSpace(in.FeePayer)),
	}
	if next.TimeoutSeconds == 0 {
		next.TimeoutSeconds = 15
	}
	if next.FeePayer == "" {
		next.FeePayer = InvoiceFeePayerCustomer
	}
	if err := validateInvoiceStoredSettings(next, s.fallback.ClientSecret != ""); err != nil {
		return nil, err
	}

	raw, err := json.Marshal(next)
	if err != nil {
		return nil, fmt.Errorf("marshal invoice settings: %w", err)
	}
	if err := s.settingRepo.Set(ctx, settingKeyInvoiceIntegrationConfig, string(raw)); err != nil {
		return nil, fmt.Errorf("save invoice settings: %w", err)
	}
	return s.GetAdminSettings(ctx)
}

// EffectiveConfig returns decrypted runtime credentials. Stored non-secret
// fields override the deployment config; an empty stored secret keeps the
// deployment secret so the first admin save does not silently disable invoicing.
func (s *InvoiceSettingsService) EffectiveConfig(ctx context.Context) (config.InvoiceIntegrationConfig, error) {
	stored, err := s.load(ctx)
	if err != nil {
		return config.InvoiceIntegrationConfig{}, err
	}
	if stored == nil {
		return s.fallback, nil
	}

	secret := stored.ClientSecret
	if secret == "" {
		secret = s.fallback.ClientSecret
	} else {
		secret, err = s.encryptor.Decrypt(secret)
		if err != nil {
			return config.InvoiceIntegrationConfig{}, fmt.Errorf("decrypt invoice client secret: %w", err)
		}
	}
	return normalizeInvoiceIntegrationConfig(config.InvoiceIntegrationConfig{
		Enabled:        stored.Enabled,
		BaseURL:        stored.BaseURL,
		ClientID:       stored.ClientID,
		ClientSecret:   secret,
		TimeoutSeconds: stored.TimeoutSeconds,
	}), nil
}

func (s *InvoiceSettingsService) FeePayer(ctx context.Context) (string, error) {
	stored, err := s.load(ctx)
	if err != nil {
		return "", err
	}
	if stored == nil {
		return InvoiceFeePayerCustomer, nil
	}
	return normalizeInvoiceFeePayer(stored.FeePayer), nil
}

func (s *InvoiceSettingsService) load(ctx context.Context) (*invoiceStoredSettings, error) {
	if s == nil || s.settingRepo == nil {
		return nil, nil //nolint:nilnil // no repository means no stored override
	}
	raw, err := s.settingRepo.GetValue(ctx, settingKeyInvoiceIntegrationConfig)
	if errors.Is(err, ErrSettingNotFound) || strings.TrimSpace(raw) == "" {
		return nil, nil //nolint:nilnil // an absent override is a valid state
	}
	if err != nil {
		return nil, fmt.Errorf("load invoice settings: %w", err)
	}
	var stored invoiceStoredSettings
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return nil, fmt.Errorf("parse invoice settings: %w", err)
	}
	stored.BaseURL = strings.TrimSpace(strings.TrimRight(stored.BaseURL, "/"))
	stored.ClientID = strings.TrimSpace(stored.ClientID)
	if stored.TimeoutSeconds == 0 {
		stored.TimeoutSeconds = 15
	}
	stored.FeePayer = normalizeInvoiceFeePayer(stored.FeePayer)
	return &stored, nil
}

func invoiceAdminSettingsFromConfig(cfg config.InvoiceIntegrationConfig) *InvoiceAdminSettings {
	return &InvoiceAdminSettings{
		Enabled:                cfg.Enabled,
		BaseURL:                cfg.BaseURL,
		ClientID:               cfg.ClientID,
		ClientSecretConfigured: cfg.ClientSecret != "",
		TimeoutSeconds:         cfg.TimeoutSeconds,
		FeePayer:               InvoiceFeePayerCustomer,
	}
}

func normalizeInvoiceIntegrationConfig(cfg config.InvoiceIntegrationConfig) config.InvoiceIntegrationConfig {
	cfg.BaseURL = strings.TrimSpace(strings.TrimRight(cfg.BaseURL, "/"))
	cfg.ClientID = strings.TrimSpace(cfg.ClientID)
	cfg.ClientSecret = strings.TrimSpace(cfg.ClientSecret)
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = 15
	}
	return cfg
}

func validateInvoiceStoredSettings(settings invoiceStoredSettings, fallbackSecretConfigured bool) error {
	switch settings.FeePayer {
	case InvoiceFeePayerCustomer, InvoiceFeePayerPlatform, InvoiceFeePayerUserChoice:
	default:
		return infraerrors.BadRequest("INVOICE_INVALID_FEE_PAYER", "invoice fee payer must be customer, platform, or user_choice")
	}
	if settings.TimeoutSeconds < 1 || settings.TimeoutSeconds > 120 {
		return infraerrors.BadRequest("INVOICE_INVALID_TIMEOUT", "invoice timeout must be between 1 and 120 seconds")
	}
	if settings.BaseURL != "" {
		parsed, err := url.Parse(settings.BaseURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
			return infraerrors.BadRequest("INVOICE_INVALID_BASE_URL", "invoice base URL must be an absolute HTTP or HTTPS URL without credentials, query, or fragment")
		}
	}
	if !settings.Enabled {
		return nil
	}
	if settings.BaseURL == "" || settings.ClientID == "" || (settings.ClientSecret == "" && !fallbackSecretConfigured) {
		return infraerrors.BadRequest("INVOICE_CONFIG_INCOMPLETE", "base URL, client ID, and client secret are required when invoicing is enabled")
	}
	return nil
}

func normalizeInvoiceFeePayer(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case InvoiceFeePayerPlatform:
		return InvoiceFeePayerPlatform
	case InvoiceFeePayerUserChoice:
		return InvoiceFeePayerUserChoice
	default:
		return InvoiceFeePayerCustomer
	}
}
