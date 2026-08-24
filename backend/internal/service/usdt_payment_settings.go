package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
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
	MinimumAmount            float64
	OrderTimeoutSeconds      int
	LatePaymentWindowMinutes int
	RequestTimeoutSeconds    int
	ReconcileIntervalSeconds int
	ReconcileBatchSize       int
	WebhookClockSkewSeconds  int
	ConfigWarnings           []string
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
	MinimumAmount            float64  `json:"minimum_amount"`
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
		return nil, classifyUSDTPaymentConfigError(err)
	}
	if stored == nil {
		settings := usdtPaymentAdminSettingsFromConfig(s.fallback)
		settings.ConfigWarnings = usdtPaymentConfigWarnings(nil, s.fallback)
		return settings, nil
	}
	settings := &USDTPaymentAdminSettings{
		Enabled:                  stored.Enabled,
		APIBase:                  resolveUSDTPaymentEndpoint(stored.APIBase, s.fallback.APIBase),
		PublicBaseURL:            resolveUSDTPaymentEndpoint(stored.PublicBaseURL, s.fallback.PublicBaseURL),
		PublicCallbackBaseURL:    resolveUSDTPaymentEndpoint(stored.PublicCallbackBaseURL, s.fallback.PublicCallbackBaseURL),
		KeyID:                    stored.KeyID,
		APISecretConfigured:      stored.APISecret != "" || s.fallback.APISecret != "",
		Fiat:                     stored.Fiat,
		EnabledNetworks:          append([]string(nil), stored.EnabledNetworks...),
		MinimumAmount:            stored.MinimumAmount,
		OrderTimeoutSeconds:      stored.OrderTimeoutSeconds,
		LatePaymentWindowMinutes: stored.LatePaymentWindowMinutes,
		RequestTimeoutSeconds:    stored.RequestTimeoutSeconds,
		ReconcileIntervalSeconds: stored.ReconcileIntervalSeconds,
		ReconcileBatchSize:       stored.ReconcileBatchSize,
		WebhookClockSkewSeconds:  stored.WebhookClockSkewSeconds,
		ConfigWarnings:           usdtPaymentConfigWarnings(stored, s.fallback),
	}
	if warning := s.usdtPaymentSecretWarning(stored); warning != "" {
		settings.ConfigWarnings = append(settings.ConfigWarnings, warning)
	}
	return settings, nil
}

// Update replaces editable fields. An empty APISecret preserves the stored
// secret (or the environment/config fallback on the first save).
func (s *USDTPaymentSettingsService) Update(ctx context.Context, in USDTPaymentAdminSettings) (*USDTPaymentAdminSettings, error) {
	stored, err := s.load(ctx)
	if err != nil {
		return nil, classifyUSDTPaymentConfigError(err)
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
		MinimumAmount:            in.MinimumAmount,
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
		MinimumAmount:       in.MinimumAmount,
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
		return config.USDTPaymentConfig{}, classifyUSDTPaymentConfigError(err)
	}
	if stored == nil {
		return s.fallback, nil
	}
	secret := stored.APISecret
	if secret == "" {
		secret = s.fallback.APISecret
	} else {
		if s.encryptor == nil {
			if s.fallback.APISecret != "" {
				secret = s.fallback.APISecret
			} else {
				return config.USDTPaymentConfig{}, classifyUSDTPaymentConfigError(errors.New("USDT API secret encryptor is unavailable"))
			}
		} else if decrypted, decryptErr := s.encryptor.Decrypt(secret); decryptErr != nil {
			// A database copied from another deployment can contain ciphertext
			// encrypted with a different TOTP key. Prefer an explicitly configured
			// deployment secret so the migrated row cannot take the payment gateway
			// offline; the admin warning tells the operator to replace the stale row.
			if s.fallback.APISecret != "" {
				secret = s.fallback.APISecret
			} else {
				return config.USDTPaymentConfig{}, classifyUSDTPaymentConfigError(fmt.Errorf("decrypt USDT API secret: %w", decryptErr))
			}
		} else {
			secret = decrypted
		}
	}
	return normalizeUSDTPaymentConfig(config.USDTPaymentConfig{
		Enabled: stored.Enabled,
		// The legacy cashier token is environment-only. Keep it available when
		// an older persisted settings row does not contain this server secret.
		LegacyToken:              s.fallback.LegacyToken,
		APIBase:                  resolveUSDTPaymentEndpoint(stored.APIBase, s.fallback.APIBase),
		PublicBaseURL:            resolveUSDTPaymentEndpoint(stored.PublicBaseURL, s.fallback.PublicBaseURL),
		PublicCallbackBaseURL:    resolveUSDTPaymentEndpoint(stored.PublicCallbackBaseURL, s.fallback.PublicCallbackBaseURL),
		KeyID:                    stored.KeyID,
		APISecret:                secret,
		Fiat:                     stored.Fiat,
		EnabledNetworks:          stored.EnabledNetworks,
		MinimumAmount:            stored.MinimumAmount,
		OrderTimeoutSeconds:      stored.OrderTimeoutSeconds,
		LatePaymentWindowMinutes: stored.LatePaymentWindowMinutes,
		RequestTimeoutSeconds:    stored.RequestTimeoutSeconds,
		ReconcileIntervalSeconds: stored.ReconcileIntervalSeconds,
		ReconcileBatchSize:       stored.ReconcileBatchSize,
		WebhookClockSkewSeconds:  stored.WebhookClockSkewSeconds,
	}), nil
}

func (s *USDTPaymentSettingsService) TestConnection(ctx context.Context, in USDTPaymentAdminSettings) (*USDTPaymentProbeResult, error) {
	// Prefer a newly entered secret before resolving the persisted one. A local
	// database may contain ciphertext created with a different TOTP key; that
	// stale ciphertext must not prevent an administrator from testing and
	// replacing the production merchant secret.
	var effective config.USDTPaymentConfig
	if strings.TrimSpace(in.APISecret) != "" {
		effective = s.fallback
		effective.APISecret = strings.TrimSpace(in.APISecret)
	} else {
		var err error
		effective, err = s.EffectiveConfig(ctx)
		if err != nil {
			return nil, classifyUSDTPaymentConfigError(err)
		}
	}
	effective.Enabled = true
	effective.APIBase = in.APIBase
	effective.PublicBaseURL = in.PublicBaseURL
	effective.PublicCallbackBaseURL = in.PublicCallbackBaseURL
	effective.KeyID = in.KeyID
	effective.Fiat = in.Fiat
	effective.EnabledNetworks = in.EnabledNetworks
	effective.MinimumAmount = in.MinimumAmount
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
		MinimumAmount:       effective.MinimumAmount,
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
		return nil, classifyUSDTPaymentProbeError("capabilities", err)
	}
	rate, err := client.ExchangeRate(ctx)
	if err != nil {
		return nil, classifyUSDTPaymentProbeError("rate", err)
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

func classifyUSDTPaymentConfigError(err error) error {
	if err == nil {
		return nil
	}
	if reason := infraerrors.Reason(err); strings.HasPrefix(reason, "USDT_PAYMENT_") {
		return err
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "decrypt usdt api secret") ||
		strings.Contains(message, "message authentication failed") ||
		strings.Contains(message, "cipher: message authentication failed") {
		return infraerrors.BadRequest(
			"USDT_PAYMENT_SECRET_DECRYPT_FAILED",
			"the saved USDT secret cannot be decrypted; restore the TOTP_ENCRYPTION_KEY used with this database or enter a new merchant HMAC Secret and save the settings",
		).WithCause(err)
	}
	if strings.Contains(message, "encryptor is unavailable") || strings.Contains(message, "encryption key") {
		return infraerrors.BadRequest(
			"USDT_PAYMENT_SECRET_ENCRYPTION_UNAVAILABLE",
			"USDT secret encryption is unavailable; configure a fixed TOTP_ENCRYPTION_KEY before testing or saving the merchant secret",
		).WithCause(err)
	}
	return infraerrors.ServiceUnavailable(
		"USDT_PAYMENT_CONFIG_UNAVAILABLE",
		"USDT payment configuration could not be loaded; check the database settings row and server encryption key",
	).WithCause(err)
}

func (s *USDTPaymentSettingsService) usdtPaymentSecretWarning(stored *usdtPaymentStoredSettings) string {
	if s == nil || stored == nil || strings.TrimSpace(stored.APISecret) == "" {
		return ""
	}
	if s.encryptor == nil {
		return "database USDT merchant secret cannot be decrypted because the server encryptor is unavailable; enter a new merchant HMAC Secret"
	}
	if _, err := s.encryptor.Decrypt(stored.APISecret); err == nil {
		return ""
	}
	if s.fallback.APISecret != "" {
		return "database USDT merchant secret is encrypted with a different TOTP_ENCRYPTION_KEY; the deployment secret is being used until you replace the migrated value"
	}
	return "database USDT merchant secret cannot be decrypted; restore the original TOTP_ENCRYPTION_KEY or enter a new merchant HMAC Secret"
}

func classifyUSDTPaymentProbeError(phase string, err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "http 401") || strings.Contains(message, "unauthorized") || strings.Contains(message, "invalid hmac"):
		return infraerrors.BadRequest("USDT_PAYMENT_CREDENTIALS_REJECTED", fmt.Sprintf("BEpusdt rejected the merchant credentials during %s probe; check Key ID and merchant HMAC Secret", phase)).WithCause(err)
	case strings.Contains(message, "http 403") || strings.Contains(message, "1010") || strings.Contains(message, "cloudflare"):
		return infraerrors.ServiceUnavailable("USDT_PAYMENT_UPSTREAM_BLOCKED", fmt.Sprintf("BEpusdt %s probe was blocked by the upstream edge (HTTP 403); use the private BEpusdt Docker/host URL for API Base instead of the Cloudflare public URL", phase)).WithCause(err)
	case strings.Contains(message, "http 404") || strings.Contains(message, "not found"):
		return infraerrors.BadRequest("USDT_PAYMENT_MERCHANT_API_MISSING", fmt.Sprintf("BEpusdt does not expose the merchant %s API at this address; deploy the merchant-enabled BEpusdt build or correct API Base", phase)).WithCause(err)
	case strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout"):
		return infraerrors.ServiceUnavailable("USDT_PAYMENT_UPSTREAM_TIMEOUT", fmt.Sprintf("BEpusdt %s probe timed out; check the server-to-server API Base and network/firewall", phase)).WithCause(err)
	default:
		return infraerrors.ServiceUnavailable("USDT_PAYMENT_UPSTREAM_UNAVAILABLE", fmt.Sprintf("BEpusdt %s probe failed; check API Base and server-side connectivity", phase)).WithCause(err)
	}
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
		MinimumAmount:            cfg.MinimumAmount,
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
	if cfg.MinimumAmount == 0 {
		cfg.MinimumAmount = config.DefaultUSDTPaymentMinimumAmount
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
	if stored.MinimumAmount == 0 {
		stored.MinimumAmount = fallback.MinimumAmount
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
	if stored.MinimumAmount < 0.01 || stored.MinimumAmount > 1_000_000 {
		return infraerrors.BadRequest("USDT_PAYMENT_MINIMUM_AMOUNT_INVALID", "minimum amount must be between 0.01 and 1000000 USDT")
	}
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

// usdtPaymentConfigWarnings detects values that commonly survive an unsafe
// local database migration. They remain warnings rather than hard failures:
// Docker-internal API bases are valid when BEpusdt is on the same network, but
// loopback/private callback URLs can never receive production webhooks.
func usdtPaymentConfigWarnings(stored *usdtPaymentStoredSettings, fallback config.USDTPaymentConfig) []string {
	values := []struct {
		field  string
		value  string
		public bool
	}{
		{"api_base", fallback.APIBase, false},
		{"public_base_url", fallback.PublicBaseURL, true},
		{"public_callback_base_url", fallback.PublicCallbackBaseURL, true},
	}
	if stored != nil {
		values[0].value = stored.APIBase
		values[1].value = stored.PublicBaseURL
		values[2].value = stored.PublicCallbackBaseURL
	}
	warnings := make([]string, 0, 4)
	for _, item := range values {
		host, local := localPaymentEndpointHost(item.value)
		if !local {
			continue
		}
		if item.public {
			warnings = append(warnings, fmt.Sprintf("%s uses local/private host %q; BEpusdt cannot reach production browser redirects or webhooks", item.field, host))
		} else {
			warnings = append(warnings, fmt.Sprintf("%s uses local/private host %q; this is valid only when BEpusdt is reachable from the same trusted Docker/host network", item.field, host))
		}
	}
	if stored != nil {
		for _, item := range []struct {
			field, stored, fallback string
		}{
			{"api_base", stored.APIBase, fallback.APIBase},
			{"public_base_url", stored.PublicBaseURL, fallback.PublicBaseURL},
			{"public_callback_base_url", stored.PublicCallbackBaseURL, fallback.PublicCallbackBaseURL},
		} {
			if item.fallback == "" || item.stored == "" || strings.EqualFold(item.stored, item.fallback) {
				continue
			}
			if _, storedLocal := localPaymentEndpointHost(item.stored); storedLocal {
				if _, fallbackLocal := localPaymentEndpointHost(item.fallback); !fallbackLocal {
					warnings = append(warnings, fmt.Sprintf("database value for %s overrides the non-local environment/config fallback; review the migrated usdt_payment_config row", item.field))
				}
			}
		}
	}
	return warnings
}

func localPaymentEndpointHost(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Hostname() == "" {
		return "", false
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if host == "localhost" || host == "localhost.localdomain" || host == "host.docker.internal" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return host, true
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified()) {
		return host, true
	}
	return host, false
}

// resolveUSDTPaymentEndpoint prevents a copied local database row from
// overriding a deployment's production endpoint. Explicit non-local database
// values remain authoritative; local values fall back only when the deployment
// provides a non-local value. This keeps local Docker setups working while
// allowing a production container to recover from a migrated localhost row.
func resolveUSDTPaymentEndpoint(stored, fallback string) string {
	stored = strings.TrimRight(strings.TrimSpace(stored), "/")
	fallback = strings.TrimRight(strings.TrimSpace(fallback), "/")
	if stored == "" {
		return fallback
	}
	if _, local := localPaymentEndpointHost(stored); local {
		if _, fallbackLocal := localPaymentEndpointHost(fallback); fallback != "" && !fallbackLocal {
			return fallback
		}
	}
	return stored
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
