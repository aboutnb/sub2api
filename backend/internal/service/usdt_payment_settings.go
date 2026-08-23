package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/usdtpayment"
)

const settingKeyUSDTPaymentConfig = "usdt_payment_config"

var ErrUSDTPaymentSecretEncryptionKeyNotConfigured = infraerrors.BadRequest(
	"USDT_PAYMENT_SECRET_ENCRYPTION_KEY_NOT_CONFIGURED",
	"cannot store the USDT API secret without a fixed TOTP_ENCRYPTION_KEY",
)

// USDTPaymentAdminSettings is the safe, editable representation returned to
// the admin UI. The API secret itself is never returned.
type USDTPaymentAdminSettings struct {
	Enabled                  bool
	APIBase                  string
	PublicBaseURL            string
	PublicCallbackBaseURL    string
	KeyID                    string
	APISecret                string
	APISecretConfigured      bool
	Fiat                     string
	EnabledNetworks          []string
	OrderTimeoutSeconds      int
	LatePaymentWindowMinutes int
	RequestTimeoutSeconds    int
	ReconcileIntervalSeconds int
	ReconcileBatchSize       int
	WebhookClockSkewSeconds  int
}

type USDTPaymentProbeResult struct {
	Ready         bool                     `json:"ready"`
	Rate          string                   `json:"rate"`
	RateUpdatedAt int64                    `json:"rate_updated_at"`
	Networks      []usdtpayment.Capability `json:"networks"`
}

type usdtPaymentStoredSettings struct {
	Enabled                  bool     `json:"enabled"`
	APIBase                  string   `json:"api_base"`
	PublicBaseURL            string   `json:"public_base_url"`
	PublicCallbackBaseURL    string   `json:"public_callback_base_url"`
	KeyID                    string   `json:"key_id"`
	APISecret                string   `json:"api_secret,omitempty"`
	Fiat                     string   `json:"fiat"`
	EnabledNetworks          []string `json:"enabled_networks"`
	OrderTimeoutSeconds      int      `json:"order_timeout_seconds"`
	LatePaymentWindowMinutes int      `json:"late_payment_window_minutes"`
	RequestTimeoutSeconds    int      `json:"request_timeout_seconds"`
	ReconcileIntervalSeconds int      `json:"reconcile_interval_seconds"`
	ReconcileBatchSize       int      `json:"reconcile_batch_size"`
	WebhookClockSkewSeconds  int      `json:"webhook_clock_skew_seconds"`
}

// USDTPaymentSettingsService stores BEpusdt credentials outside the public
// system-settings document and resolves config-file values as a fallback.
type USDTPaymentSettingsService struct {
	settingRepo             SettingRepository
	encryptor               SecretEncryptor
	fallback                config.USDTPaymentConfig
	encryptionKeyConfigured bool
}

func NewUSDTPaymentSettingsService(settingRepo SettingRepository, encryptor SecretEncryptor, fallback config.USDTPaymentConfig, encryptionKeyConfigured bool) *USDTPaymentSettingsService {
	return &USDTPaymentSettingsService{
		settingRepo:             settingRepo,
		encryptor:               encryptor,
		fallback:                normalizeUSDTPaymentConfig(fallback),
		encryptionKeyConfigured: encryptionKeyConfigured,
	}
}

func (s *USDTPaymentSettingsService) GetAdminSettings(ctx context.Context) (*USDTPaymentAdminSettings, error) {
	stored, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		return usdtPaymentAdminSettingsFromConfig(s.fallback), nil
	}
	return &USDTPaymentAdminSettings{
		Enabled:                  stored.Enabled,
		APIBase:                  stored.APIBase,
		PublicBaseURL:            stored.PublicBaseURL,
		PublicCallbackBaseURL:    stored.PublicCallbackBaseURL,
		KeyID:                    stored.KeyID,
		APISecretConfigured:      stored.APISecret != "" || s.fallback.APISecret != "",
		Fiat:                     stored.Fiat,
		EnabledNetworks:          append([]string(nil), stored.EnabledNetworks...),
		OrderTimeoutSeconds:      stored.OrderTimeoutSeconds,
		LatePaymentWindowMinutes: stored.LatePaymentWindowMinutes,
		RequestTimeoutSeconds:    stored.RequestTimeoutSeconds,
		ReconcileIntervalSeconds: stored.ReconcileIntervalSeconds,
		ReconcileBatchSize:       stored.ReconcileBatchSize,
		WebhookClockSkewSeconds:  stored.WebhookClockSkewSeconds,
	}, nil
}

// Update replaces editable fields. An empty APISecret preserves the stored
// secret (or the environment/config fallback on the first save).
func (s *USDTPaymentSettingsService) Update(ctx context.Context, in USDTPaymentAdminSettings) (*USDTPaymentAdminSettings, error) {
	stored, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	if stored == nil {
		stored = &usdtPaymentStoredSettings{}
	}

	secret := strings.TrimSpace(in.APISecret)
	if secret == "" {
		secret = stored.APISecret
	} else {
		if !s.encryptionKeyConfigured || s.encryptor == nil {
			return nil, ErrUSDTPaymentSecretEncryptionKeyNotConfigured
		}
		secret, err = s.encryptor.Encrypt(secret)
		if err != nil {
			return nil, fmt.Errorf("encrypt USDT API secret: %w", err)
		}
	}

	next := usdtPaymentStoredSettings{
		Enabled:                  in.Enabled,
		APIBase:                  strings.TrimRight(strings.TrimSpace(in.APIBase), "/"),
		PublicBaseURL:            strings.TrimRight(strings.TrimSpace(in.PublicBaseURL), "/"),
		PublicCallbackBaseURL:    strings.TrimRight(strings.TrimSpace(in.PublicCallbackBaseURL), "/"),
		KeyID:                    strings.TrimSpace(in.KeyID),
		APISecret:                secret,
		Fiat:                     strings.ToUpper(strings.TrimSpace(in.Fiat)),
		EnabledNetworks:          normalizeUSDTPaymentNetworks(in.EnabledNetworks),
		OrderTimeoutSeconds:      in.OrderTimeoutSeconds,
		LatePaymentWindowMinutes: in.LatePaymentWindowMinutes,
		RequestTimeoutSeconds:    in.RequestTimeoutSeconds,
		ReconcileIntervalSeconds: in.ReconcileIntervalSeconds,
		ReconcileBatchSize:       in.ReconcileBatchSize,
		WebhookClockSkewSeconds:  in.WebhookClockSkewSeconds,
	}
	applyUSDTPaymentDefaults(&next, s.fallback)
	if err := validateUSDTPaymentSettings(next, secret != "" || s.fallback.APISecret != ""); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(next)
	if err != nil {
		return nil, fmt.Errorf("marshal USDT payment settings: %w", err)
	}
	if err := s.settingRepo.Set(ctx, settingKeyUSDTPaymentConfig, string(raw)); err != nil {
		return nil, fmt.Errorf("save USDT payment settings: %w", err)
	}
	return s.GetAdminSettings(ctx)
}

func (s *USDTPaymentSettingsService) ValidateUpdate(ctx context.Context, in USDTPaymentAdminSettings) error {
	if strings.TrimSpace(in.APISecret) != "" && !s.encryptionKeyConfigured {
		return ErrUSDTPaymentSecretEncryptionKeyNotConfigured
	}
	effective, err := s.EffectiveConfig(ctx)
	if err != nil {
		return err
	}
	secretAvailable := strings.TrimSpace(in.APISecret) != "" || effective.APISecret != ""
	validation := usdtPaymentStoredSettings{
		Enabled: in.Enabled, APIBase: in.APIBase, PublicBaseURL: in.PublicBaseURL,
		PublicCallbackBaseURL: in.PublicCallbackBaseURL, KeyID: in.KeyID,
		Fiat: in.Fiat, EnabledNetworks: in.EnabledNetworks,
		OrderTimeoutSeconds: in.OrderTimeoutSeconds, LatePaymentWindowMinutes: in.LatePaymentWindowMinutes,
		RequestTimeoutSeconds: in.RequestTimeoutSeconds, ReconcileIntervalSeconds: in.ReconcileIntervalSeconds,
		ReconcileBatchSize: in.ReconcileBatchSize, WebhookClockSkewSeconds: in.WebhookClockSkewSeconds,
	}
	validation.APIBase = strings.TrimRight(strings.TrimSpace(validation.APIBase), "/")
	validation.PublicBaseURL = strings.TrimRight(strings.TrimSpace(validation.PublicBaseURL), "/")
	validation.PublicCallbackBaseURL = strings.TrimRight(strings.TrimSpace(validation.PublicCallbackBaseURL), "/")
	validation.KeyID = strings.TrimSpace(validation.KeyID)
	validation.Fiat = strings.ToUpper(strings.TrimSpace(validation.Fiat))
	validation.EnabledNetworks = normalizeUSDTPaymentNetworks(validation.EnabledNetworks)
	applyUSDTPaymentDefaults(&validation, s.fallback)
	return validateUSDTPaymentSettings(validation, secretAvailable)
}

// EffectiveConfig resolves the live runtime configuration. It is intentionally
// shaped as a standalone method so the USDT package can depend on it without a
// package cycle.
func (s *USDTPaymentSettingsService) EffectiveConfig(ctx context.Context) (config.USDTPaymentConfig, error) {
	stored, err := s.load(ctx)
	if err != nil {
		return config.USDTPaymentConfig{}, err
	}
	if stored == nil {
		return s.fallback, nil
	}
	secret := stored.APISecret
	if secret == "" {
		secret = s.fallback.APISecret
	} else {
		if s.encryptor == nil {
			return config.USDTPaymentConfig{}, errors.New("USDT API secret encryptor is unavailable")
		}
		secret, err = s.encryptor.Decrypt(secret)
		if err != nil {
			return config.USDTPaymentConfig{}, fmt.Errorf("decrypt USDT API secret: %w", err)
		}
	}
	return normalizeUSDTPaymentConfig(config.USDTPaymentConfig{
		Enabled: stored.Enabled,
		// The legacy cashier token is environment-only. Keep it available when
		// an older persisted settings row does not contain this server secret.
		LegacyToken:              s.fallback.LegacyToken,
		APIBase:                  stored.APIBase,
		PublicBaseURL:            stored.PublicBaseURL,
		PublicCallbackBaseURL:    stored.PublicCallbackBaseURL,
		KeyID:                    stored.KeyID,
		APISecret:                secret,
		Fiat:                     stored.Fiat,
		EnabledNetworks:          stored.EnabledNetworks,
		OrderTimeoutSeconds:      stored.OrderTimeoutSeconds,
		LatePaymentWindowMinutes: stored.LatePaymentWindowMinutes,
		RequestTimeoutSeconds:    stored.RequestTimeoutSeconds,
		ReconcileIntervalSeconds: stored.ReconcileIntervalSeconds,
		ReconcileBatchSize:       stored.ReconcileBatchSize,
		WebhookClockSkewSeconds:  stored.WebhookClockSkewSeconds,
	}), nil
}

func (s *USDTPaymentSettingsService) TestConnection(ctx context.Context, in USDTPaymentAdminSettings) (*USDTPaymentProbeResult, error) {
	effective, err := s.EffectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	if secret := strings.TrimSpace(in.APISecret); secret != "" {
		effective.APISecret = secret
	}
	effective.Enabled = true
	effective.APIBase = in.APIBase
	effective.PublicBaseURL = in.PublicBaseURL
	effective.PublicCallbackBaseURL = in.PublicCallbackBaseURL
	effective.KeyID = in.KeyID
	effective.Fiat = in.Fiat
	effective.EnabledNetworks = in.EnabledNetworks
	effective.OrderTimeoutSeconds = in.OrderTimeoutSeconds
	effective.LatePaymentWindowMinutes = in.LatePaymentWindowMinutes
	effective.RequestTimeoutSeconds = in.RequestTimeoutSeconds
	effective.ReconcileIntervalSeconds = in.ReconcileIntervalSeconds
	effective.ReconcileBatchSize = in.ReconcileBatchSize
	effective.WebhookClockSkewSeconds = in.WebhookClockSkewSeconds
	effective = normalizeUSDTPaymentConfig(effective)

	validation := usdtPaymentStoredSettings{
		Enabled: true, APIBase: effective.APIBase, PublicBaseURL: effective.PublicBaseURL,
		PublicCallbackBaseURL: effective.PublicCallbackBaseURL, KeyID: effective.KeyID,
		Fiat: effective.Fiat, EnabledNetworks: effective.EnabledNetworks,
		OrderTimeoutSeconds: effective.OrderTimeoutSeconds, LatePaymentWindowMinutes: effective.LatePaymentWindowMinutes,
		RequestTimeoutSeconds: effective.RequestTimeoutSeconds, ReconcileIntervalSeconds: effective.ReconcileIntervalSeconds,
		ReconcileBatchSize: effective.ReconcileBatchSize, WebhookClockSkewSeconds: effective.WebhookClockSkewSeconds,
	}
	if err := validateUSDTPaymentSettings(validation, effective.APISecret != ""); err != nil {
		return nil, err
	}

	client := usdtpayment.NewClient(&config.Config{USDTPayment: effective})
	networks, err := client.Capabilities(ctx, effective.EnabledNetworks)
	if err != nil {
		return nil, fmt.Errorf("probe BEpusdt capabilities: %w", err)
	}
	rate, err := client.ExchangeRate(ctx)
	if err != nil {
		return nil, fmt.Errorf("probe BEpusdt rate: %w", err)
	}
	ready := false
	for _, network := range networks {
		if network.AcceptingOrders {
			ready = true
			break
		}
	}
	return &USDTPaymentProbeResult{
		Ready: ready, Rate: rate.Rate, RateUpdatedAt: rate.UpdatedAt, Networks: networks,
	}, nil
}

func (s *USDTPaymentSettingsService) load(ctx context.Context) (*usdtPaymentStoredSettings, error) {
	if s == nil || s.settingRepo == nil {
		return nil, nil //nolint:nilnil // no repository means no stored override
	}
	raw, err := s.settingRepo.GetValue(ctx, settingKeyUSDTPaymentConfig)
	if errors.Is(err, ErrSettingNotFound) || strings.TrimSpace(raw) == "" {
		return nil, nil //nolint:nilnil // an absent override is a valid state
	}
	if err != nil {
		return nil, fmt.Errorf("load USDT payment settings: %w", err)
	}
	var stored usdtPaymentStoredSettings
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return nil, fmt.Errorf("parse USDT payment settings: %w", err)
	}
	stored.APIBase = strings.TrimRight(strings.TrimSpace(stored.APIBase), "/")
	stored.PublicBaseURL = strings.TrimRight(strings.TrimSpace(stored.PublicBaseURL), "/")
	stored.PublicCallbackBaseURL = strings.TrimRight(strings.TrimSpace(stored.PublicCallbackBaseURL), "/")
	stored.KeyID = strings.TrimSpace(stored.KeyID)
	stored.Fiat = strings.ToUpper(strings.TrimSpace(stored.Fiat))
	stored.EnabledNetworks = normalizeUSDTPaymentNetworks(stored.EnabledNetworks)
	var rawFields map[string]json.RawMessage
	_ = json.Unmarshal([]byte(raw), &rawFields)
	applyUSDTPaymentDefaults(&stored, s.fallback)
	// Zero is a valid explicit late-payment window. Only use the fallback when
	// the legacy JSON omitted the field altogether.
	if _, ok := rawFields["late_payment_window_minutes"]; ok {
		var late int
		if err := json.Unmarshal(rawFields["late_payment_window_minutes"], &late); err == nil {
			stored.LatePaymentWindowMinutes = late
		}
	}
	return &stored, nil
}

func usdtPaymentAdminSettingsFromConfig(cfg config.USDTPaymentConfig) *USDTPaymentAdminSettings {
	cfg = normalizeUSDTPaymentConfig(cfg)
	return &USDTPaymentAdminSettings{
		Enabled:                  cfg.Enabled,
		APIBase:                  cfg.APIBase,
		PublicBaseURL:            cfg.PublicBaseURL,
		PublicCallbackBaseURL:    cfg.PublicCallbackBaseURL,
		KeyID:                    cfg.KeyID,
		APISecretConfigured:      cfg.APISecret != "",
		Fiat:                     cfg.Fiat,
		EnabledNetworks:          append([]string(nil), cfg.EnabledNetworks...),
		OrderTimeoutSeconds:      cfg.OrderTimeoutSeconds,
		LatePaymentWindowMinutes: cfg.LatePaymentWindowMinutes,
		RequestTimeoutSeconds:    cfg.RequestTimeoutSeconds,
		ReconcileIntervalSeconds: cfg.ReconcileIntervalSeconds,
		ReconcileBatchSize:       cfg.ReconcileBatchSize,
		WebhookClockSkewSeconds:  cfg.WebhookClockSkewSeconds,
	}
}

func normalizeUSDTPaymentConfig(cfg config.USDTPaymentConfig) config.USDTPaymentConfig {
	cfg.APIBase = strings.TrimRight(strings.TrimSpace(cfg.APIBase), "/")
	cfg.PublicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")
	cfg.PublicCallbackBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicCallbackBaseURL), "/")
	cfg.KeyID = strings.TrimSpace(cfg.KeyID)
	cfg.APISecret = strings.TrimSpace(cfg.APISecret)
	cfg.Fiat = strings.ToUpper(strings.TrimSpace(cfg.Fiat))
	cfg.EnabledNetworks = normalizeUSDTPaymentNetworks(cfg.EnabledNetworks)
	if cfg.Fiat == "" {
		cfg.Fiat = "CNY"
	}
	if cfg.OrderTimeoutSeconds == 0 {
		cfg.OrderTimeoutSeconds = 900
	}
	if cfg.RequestTimeoutSeconds == 0 {
		cfg.RequestTimeoutSeconds = 6
	}
	if cfg.ReconcileIntervalSeconds == 0 {
		cfg.ReconcileIntervalSeconds = 2
	}
	if cfg.ReconcileBatchSize == 0 {
		cfg.ReconcileBatchSize = 100
	}
	if cfg.WebhookClockSkewSeconds == 0 {
		cfg.WebhookClockSkewSeconds = 300
	}
	return cfg
}

func applyUSDTPaymentDefaults(stored *usdtPaymentStoredSettings, fallback config.USDTPaymentConfig) {
	fallback = normalizeUSDTPaymentConfig(fallback)
	if stored.Fiat == "" {
		stored.Fiat = fallback.Fiat
	}
	if stored.OrderTimeoutSeconds == 0 {
		stored.OrderTimeoutSeconds = fallback.OrderTimeoutSeconds
	}
	if stored.RequestTimeoutSeconds == 0 {
		stored.RequestTimeoutSeconds = fallback.RequestTimeoutSeconds
	}
	if stored.ReconcileIntervalSeconds == 0 {
		stored.ReconcileIntervalSeconds = fallback.ReconcileIntervalSeconds
	}
	if stored.ReconcileBatchSize == 0 {
		stored.ReconcileBatchSize = fallback.ReconcileBatchSize
	}
	if stored.WebhookClockSkewSeconds == 0 {
		stored.WebhookClockSkewSeconds = fallback.WebhookClockSkewSeconds
	}
}

func validateUSDTPaymentSettings(stored usdtPaymentStoredSettings, secretAvailable bool) error {
	if stored.OrderTimeoutSeconds < 180 || stored.OrderTimeoutSeconds > 3600 {
		return infraerrors.BadRequest("USDT_PAYMENT_ORDER_TIMEOUT_INVALID", "order timeout must be between 180 and 3600 seconds")
	}
	if stored.LatePaymentWindowMinutes < 0 || stored.LatePaymentWindowMinutes > 1440 {
		return infraerrors.BadRequest("USDT_PAYMENT_LATE_WINDOW_INVALID", "late payment window must be between 0 and 1440 minutes")
	}
	if stored.RequestTimeoutSeconds < 1 || stored.RequestTimeoutSeconds > 30 {
		return infraerrors.BadRequest("USDT_PAYMENT_REQUEST_TIMEOUT_INVALID", "request timeout must be between 1 and 30 seconds")
	}
	if stored.ReconcileIntervalSeconds < 1 || stored.ReconcileIntervalSeconds > 60 {
		return infraerrors.BadRequest("USDT_PAYMENT_RECONCILE_INTERVAL_INVALID", "reconcile interval must be between 1 and 60 seconds")
	}
	if stored.ReconcileBatchSize < 1 || stored.ReconcileBatchSize > 1000 {
		return infraerrors.BadRequest("USDT_PAYMENT_RECONCILE_BATCH_INVALID", "reconcile batch size must be between 1 and 1000")
	}
	if stored.WebhookClockSkewSeconds < 30 || stored.WebhookClockSkewSeconds > 900 {
		return infraerrors.BadRequest("USDT_PAYMENT_WEBHOOK_SKEW_INVALID", "webhook clock skew must be between 30 and 900 seconds")
	}
	if !stored.Enabled {
		return nil
	}
	if stored.KeyID == "" || !secretAvailable {
		return infraerrors.BadRequest("USDT_PAYMENT_CREDENTIALS_REQUIRED", "key ID and API secret are required when USDT payment is enabled")
	}
	for name, value := range map[string]string{
		"api_base": stored.APIBase, "public_callback_base_url": stored.PublicCallbackBaseURL,
	} {
		if err := config.ValidateAbsoluteHTTPURL(value); err != nil {
			return infraerrors.BadRequest("USDT_PAYMENT_URL_INVALID", fmt.Sprintf("%s is invalid: %v", name, err))
		}
	}
	if stored.PublicBaseURL != "" {
		if err := config.ValidateAbsoluteHTTPURL(stored.PublicBaseURL); err != nil {
			return infraerrors.BadRequest("USDT_PAYMENT_URL_INVALID", fmt.Sprintf("public_base_url is invalid: %v", err))
		}
	}
	if stored.Fiat != "CNY" {
		return infraerrors.BadRequest("USDT_PAYMENT_FIAT_INVALID", "USDT payment currently supports only CNY")
	}
	if len(stored.EnabledNetworks) == 0 {
		return infraerrors.BadRequest("USDT_PAYMENT_NETWORKS_REQUIRED", "at least one USDT network is required when enabled")
	}
	allowed := map[string]bool{"tron": true, "bsc": true, "ethereum": true, "polygon": true, "arbitrum": true, "solana": true, "ton": true, "aptos": true, "xlayer": true, "plasma": true}
	for _, network := range stored.EnabledNetworks {
		if !allowed[strings.ToLower(network)] {
			return infraerrors.BadRequest("USDT_PAYMENT_NETWORK_INVALID", fmt.Sprintf("unsupported USDT network %q", network))
		}
	}
	return nil
}

func normalizeUSDTPaymentNetworks(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		network := strings.ToLower(strings.TrimSpace(value))
		if network == "" {
			continue
		}
		if _, ok := seen[network]; ok {
			continue
		}
		seen[network] = struct{}{}
		result = append(result, network)
	}
	return result
}
