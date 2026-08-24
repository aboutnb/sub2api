package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func testUSDTPaymentAdminSettings() USDTPaymentAdminSettings {
	return USDTPaymentAdminSettings{
		Enabled:                  true,
		APIBase:                  "http://bepusdt:8080/",
		PublicBaseURL:            "https://pay.example.test/",
		PublicCallbackBaseURL:    "https://app.example.test/",
		KeyID:                    "sub2api",
		APISecret:                "secret-a",
		Fiat:                     "cny",
		EnabledNetworks:          []string{"BSC", "bsc"},
		OrderTimeoutSeconds:      900,
		LatePaymentWindowMinutes: 60,
		RequestTimeoutSeconds:    6,
		ReconcileIntervalSeconds: 2,
		ReconcileBatchSize:       100,
		WebhookClockSkewSeconds:  300,
	}
}

func TestUSDTPaymentSettingsFallbackDoesNotExposeSecret(t *testing.T) {
	ctx := context.Background()
	settings := NewUSDTPaymentSettingsService(
		newInvoiceSettingsTestRepo(),
		invoiceSettingsTestEncryptor{},
		config.USDTPaymentConfig{
			Enabled: true, APIBase: "http://bepusdt:8080/", PublicBaseURL: "https://pay.example.test/",
			PublicCallbackBaseURL: "https://app.example.test/", KeyID: "sub2api", APISecret: "fallback-secret",
			Fiat: "CNY", EnabledNetworks: []string{"bsc"}, OrderTimeoutSeconds: 900,
			LatePaymentWindowMinutes: 60, RequestTimeoutSeconds: 6, ReconcileIntervalSeconds: 2,
			ReconcileBatchSize: 100, WebhookClockSkewSeconds: 300,
		},
		true,
	)

	admin, err := settings.GetAdminSettings(ctx)
	if err != nil {
		t.Fatalf("get admin settings: %v", err)
	}
	if admin.APISecret != "" || !admin.APISecretConfigured {
		t.Fatalf("secret leaked or configured state lost: %#v", admin)
	}

	effective, err := settings.EffectiveConfig(ctx)
	if err != nil {
		t.Fatalf("get effective config: %v", err)
	}
	if effective.APISecret != "fallback-secret" || effective.APIBase != "http://bepusdt:8080" {
		t.Fatalf("unexpected fallback config: %#v", effective)
	}
}

func TestUSDTPaymentSettingsEncryptsAndPreservesSecret(t *testing.T) {
	ctx := context.Background()
	repo := newInvoiceSettingsTestRepo()
	settings := NewUSDTPaymentSettingsService(repo, invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{}, true)

	input := testUSDTPaymentAdminSettings()
	if _, err := settings.Update(ctx, input); err != nil {
		t.Fatalf("save USDT settings: %v", err)
	}

	var stored usdtPaymentStoredSettings
	if err := json.Unmarshal([]byte(repo.data[settingKeyUSDTPaymentConfig]), &stored); err != nil {
		t.Fatalf("decode stored settings: %v", err)
	}
	if stored.APISecret != "encrypted:secret-a" {
		t.Fatalf("API secret was not encrypted: %q", stored.APISecret)
	}
	if stored.MinimumAmount != config.DefaultUSDTPaymentMinimumAmount {
		t.Fatalf("minimum amount = %v, want default %v", stored.MinimumAmount, config.DefaultUSDTPaymentMinimumAmount)
	}
	if len(stored.EnabledNetworks) != 1 || stored.EnabledNetworks[0] != "bsc" {
		t.Fatalf("networks were not normalized: %#v", stored.EnabledNetworks)
	}

	input.APISecret = ""
	input.APIBase = "http://bepusdt-v2:8080"
	if _, err := settings.Update(ctx, input); err != nil {
		t.Fatalf("update without replacing secret: %v", err)
	}
	effective, err := settings.EffectiveConfig(ctx)
	if err != nil {
		t.Fatalf("get effective config: %v", err)
	}
	if effective.APISecret != "secret-a" || effective.APIBase != "http://bepusdt-v2:8080" {
		t.Fatalf("unexpected effective config: %#v", effective)
	}
}

func TestUSDTPaymentSettingsPersistsCustomMinimumAmount(t *testing.T) {
	settings := NewUSDTPaymentSettingsService(newInvoiceSettingsTestRepo(), invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{}, true)
	input := testUSDTPaymentAdminSettings()
	input.MinimumAmount = 12.5
	if _, err := settings.Update(context.Background(), input); err != nil {
		t.Fatalf("save custom minimum amount: %v", err)
	}
	admin, err := settings.GetAdminSettings(context.Background())
	if err != nil {
		t.Fatalf("read custom minimum amount: %v", err)
	}
	if admin.MinimumAmount != 12.5 {
		t.Fatalf("minimum amount = %v, want 12.5", admin.MinimumAmount)
	}
}

func TestUSDTPaymentSettingsRejectsInvalidMinimumAmount(t *testing.T) {
	settings := NewUSDTPaymentSettingsService(newInvoiceSettingsTestRepo(), invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{}, true)
	input := testUSDTPaymentAdminSettings()
	input.MinimumAmount = 0.001
	if err := settings.ValidateUpdate(context.Background(), input); err == nil {
		t.Fatal("accepted minimum amount below 0.01")
	}
	input.MinimumAmount = 1_000_001
	if err := settings.ValidateUpdate(context.Background(), input); err == nil {
		t.Fatal("accepted excessive minimum amount")
	}
}

func TestUSDTPaymentSettingsRequiresFixedEncryptionKeyForNewSecret(t *testing.T) {
	settings := NewUSDTPaymentSettingsService(
		newInvoiceSettingsTestRepo(),
		invoiceSettingsTestEncryptor{},
		config.USDTPaymentConfig{},
		false,
	)
	_, err := settings.Update(context.Background(), testUSDTPaymentAdminSettings())
	if err != ErrUSDTPaymentSecretEncryptionKeyNotConfigured {
		t.Fatalf("error = %v, want fixed encryption key error", err)
	}
}

func TestUSDTPaymentSettingsValidateUpdateRejectsInvalidEnabledConfigBeforeWrite(t *testing.T) {
	repo := newInvoiceSettingsTestRepo()
	settings := NewUSDTPaymentSettingsService(repo, invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{}, true)
	input := testUSDTPaymentAdminSettings()
	input.EnabledNetworks = nil
	if err := settings.ValidateUpdate(context.Background(), input); err == nil {
		t.Fatal("expected enabled settings without a network to fail validation")
	}
	if len(repo.data) != 0 {
		t.Fatalf("validation wrote settings: %#v", repo.data)
	}
}

func TestUSDTPaymentSettingsAllowsInlineCheckoutWithoutPublicBEpusdtURL(t *testing.T) {
	settings := NewUSDTPaymentSettingsService(newInvoiceSettingsTestRepo(), invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{}, true)
	input := testUSDTPaymentAdminSettings()
	input.PublicBaseURL = ""
	if err := settings.ValidateUpdate(context.Background(), input); err != nil {
		t.Fatalf("inline checkout should not require public BEpusdt URL: %v", err)
	}
}

func TestUSDTPaymentSettingsTestConnectionUsesEffectiveSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-BEPUSDT-Key-Id") != "sub2api" || r.Header.Get("X-BEPUSDT-Signature") == "" {
			t.Errorf("missing signed merchant headers")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/merchant/capabilities":
			_, _ = w.Write([]byte(`{"code":"ok","data":{"networks":[{"crypto":"USDT","network":"bsc","network_name":"BSC","trade_type":"usdt.bep20","wallet_count":1,"rpc_endpoint_set":true,"accepting_orders":true}]}}`))
		case "/api/v1/merchant/rate":
			_, _ = w.Write([]byte(`{"code":"ok","data":{"crypto":"USDT","fiat":"CNY","rate":"6.72","updated_at":1700000000}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	settings := NewUSDTPaymentSettingsService(newInvoiceSettingsTestRepo(), invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{
		Enabled: true, APIBase: server.URL, PublicBaseURL: server.URL, PublicCallbackBaseURL: server.URL,
		KeyID: "sub2api", APISecret: "fallback-secret", Fiat: "CNY", EnabledNetworks: []string{"bsc"},
		OrderTimeoutSeconds: 900, LatePaymentWindowMinutes: 60, RequestTimeoutSeconds: 6,
		ReconcileIntervalSeconds: 2, ReconcileBatchSize: 100, WebhookClockSkewSeconds: 300,
	}, true)
	input := testUSDTPaymentAdminSettings()
	input.APIBase = server.URL
	input.PublicBaseURL = server.URL
	input.PublicCallbackBaseURL = server.URL
	input.APISecret = ""
	result, err := settings.TestConnection(context.Background(), input)
	if err != nil {
		t.Fatalf("test connection: %v", err)
	}
	if !result.Ready || result.Rate != "6.72" || len(result.Networks) != 1 || !result.Networks[0].AcceptingOrders {
		t.Fatalf("unexpected probe result: %#v", result)
	}
}

func TestUSDTPaymentSettingsWarnsAboutMigratedLocalEndpoints(t *testing.T) {
	settings := NewUSDTPaymentSettingsService(
		newInvoiceSettingsTestRepo(), invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{
			APIBase:                  "https://upay.example.test",
			PublicBaseURL:            "https://upay.example.test",
			PublicCallbackBaseURL:    "https://app.example.test",
			Enabled:                  true,
			KeyID:                    "sub2api",
			APISecret:                "fallback-secret",
			Fiat:                     "CNY",
			EnabledNetworks:          []string{"bsc"},
			OrderTimeoutSeconds:      900,
			LatePaymentWindowMinutes: 60,
			RequestTimeoutSeconds:    6,
			ReconcileIntervalSeconds: 2,
			ReconcileBatchSize:       100,
			WebhookClockSkewSeconds:  300,
		}, true)
	repo := settings.settingRepo.(*invoiceSettingsTestRepo)
	legacy := `{"enabled":true,"api_base":"http://host.docker.internal:18080","public_base_url":"http://127.0.0.1:18080","public_callback_base_url":"http://localhost:8080","key_id":"sub2api","fiat":"CNY","enabled_networks":["bsc"],"order_timeout_seconds":900,"late_payment_window_minutes":60,"request_timeout_seconds":6,"reconcile_interval_seconds":2,"reconcile_batch_size":100,"webhook_clock_skew_seconds":300}`
	repo.data[settingKeyUSDTPaymentConfig] = legacy
	admin, err := settings.GetAdminSettings(context.Background())
	if err != nil {
		t.Fatalf("get admin settings: %v", err)
	}
	if len(admin.ConfigWarnings) < 4 {
		t.Fatalf("expected local endpoint and migration warnings, got %#v", admin.ConfigWarnings)
	}
	effective, err := settings.EffectiveConfig(context.Background())
	if err != nil {
		t.Fatalf("effective migrated config: %v", err)
	}
	if effective.APIBase != "https://upay.example.test" || effective.PublicBaseURL != "https://upay.example.test" || effective.PublicCallbackBaseURL != "https://app.example.test" {
		t.Fatalf("production fallback endpoints were not selected: %#v", effective)
	}
}

func TestUSDTPaymentSettingsDoesNotWarnForDockerAPIAndPublicURLs(t *testing.T) {
	settings := NewUSDTPaymentSettingsService(
		newInvoiceSettingsTestRepo(), invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{
			Enabled: true, APIBase: "http://bepusdt:8080", PublicBaseURL: "https://pay.example.test", PublicCallbackBaseURL: "https://app.example.test",
			KeyID: "sub2api", APISecret: "fallback-secret", Fiat: "CNY", EnabledNetworks: []string{"bsc"},
			OrderTimeoutSeconds: 900, LatePaymentWindowMinutes: 60, RequestTimeoutSeconds: 6, ReconcileIntervalSeconds: 2, ReconcileBatchSize: 100, WebhookClockSkewSeconds: 300,
		}, true)
	admin, err := settings.GetAdminSettings(context.Background())
	if err != nil {
		t.Fatalf("get admin settings: %v", err)
	}
	if len(admin.ConfigWarnings) != 0 {
		t.Fatalf("unexpected warnings: %#v", admin.ConfigWarnings)
	}
}

func TestUSDTPaymentSettingsUsesDeploymentSecretWhenMigratedCiphertextCannotDecrypt(t *testing.T) {
	repo := newInvoiceSettingsTestRepo()
	repo.data[settingKeyUSDTPaymentConfig] = `{"enabled":true,"api_base":"http://bepusdt:8080","public_callback_base_url":"https://app.example.test","key_id":"sub2api","api_secret":"ciphertext-from-another-key","fiat":"CNY","enabled_networks":["bsc"],"order_timeout_seconds":900,"late_payment_window_minutes":60,"request_timeout_seconds":6,"reconcile_interval_seconds":2,"reconcile_batch_size":100,"webhook_clock_skew_seconds":300}`
	settings := NewUSDTPaymentSettingsService(repo, invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{
		Enabled: true, APIBase: "http://bepusdt:8080", PublicCallbackBaseURL: "https://app.example.test",
		KeyID: "sub2api", APISecret: "deployment-secret", Fiat: "CNY", EnabledNetworks: []string{"bsc"},
		OrderTimeoutSeconds: 900, LatePaymentWindowMinutes: 60, RequestTimeoutSeconds: 6,
		ReconcileIntervalSeconds: 2, ReconcileBatchSize: 100, WebhookClockSkewSeconds: 300,
	}, true)
	effective, err := settings.EffectiveConfig(context.Background())
	if err != nil {
		t.Fatalf("effective config should fall back to deployment secret: %v", err)
	}
	if effective.APISecret != "deployment-secret" {
		t.Fatalf("effective secret = %q, want deployment fallback", effective.APISecret)
	}
	admin, err := settings.GetAdminSettings(context.Background())
	if err != nil {
		t.Fatalf("get admin settings: %v", err)
	}
	if len(admin.ConfigWarnings) == 0 || !strings.Contains(strings.Join(admin.ConfigWarnings, "\n"), "different TOTP_ENCRYPTION_KEY") {
		t.Fatalf("missing migrated-secret warning: %#v", admin.ConfigWarnings)
	}
}

func TestUSDTPaymentSettingsReturnsStructuredErrorWithoutSecretFallback(t *testing.T) {
	repo := newInvoiceSettingsTestRepo()
	repo.data[settingKeyUSDTPaymentConfig] = `{"enabled":true,"api_base":"http://bepusdt:8080","public_callback_base_url":"https://app.example.test","key_id":"sub2api","api_secret":"ciphertext-from-another-key","fiat":"CNY","enabled_networks":["bsc"],"order_timeout_seconds":900,"late_payment_window_minutes":60,"request_timeout_seconds":6,"reconcile_interval_seconds":2,"reconcile_batch_size":100,"webhook_clock_skew_seconds":300}`
	settings := NewUSDTPaymentSettingsService(repo, invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{}, true)
	_, err := settings.EffectiveConfig(context.Background())
	if infraerrors.Reason(err) != "USDT_PAYMENT_SECRET_DECRYPT_FAILED" {
		t.Fatalf("reason = %q, want USDT_PAYMENT_SECRET_DECRYPT_FAILED; err=%v", infraerrors.Reason(err), err)
	}
}

func TestClassifyUSDTPaymentProbeError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		reason string
	}{
		{"credentials", errors.New("BEpusdt HTTP 401 (Unauthorized): missing HMAC v2 headers"), "USDT_PAYMENT_CREDENTIALS_REJECTED"},
		{"edge", errors.New("BEpusdt HTTP 403 (): cloudflare access denied"), "USDT_PAYMENT_UPSTREAM_BLOCKED"},
		{"missing", errors.New("BEpusdt HTTP 404"), "USDT_PAYMENT_MERCHANT_API_MISSING"},
		{"timeout", errors.New("call BEpusdt: context deadline exceeded"), "USDT_PAYMENT_UPSTREAM_TIMEOUT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyUSDTPaymentProbeError("rate", tt.err)
			if got == nil || infraerrors.Reason(got) != tt.reason {
				t.Fatalf("reason = %q, want %q; error = %v", infraerrors.Reason(got), tt.reason, got)
			}
		})
	}
}

func TestUSDTPaymentTestConnectionUsesNewSecretWhenPersistedCiphertextIsUndecryptable(t *testing.T) {
	repo := newInvoiceSettingsTestRepo()
	repo.data[settingKeyUSDTPaymentConfig] = `{"enabled":true,"api_base":"http://old.local:18080","public_callback_base_url":"https://app.example.test","key_id":"old-key","api_secret":"ciphertext-from-local-key","fiat":"CNY","enabled_networks":["bsc"],"order_timeout_seconds":900,"late_payment_window_minutes":60,"request_timeout_seconds":6,"reconcile_interval_seconds":2,"reconcile_batch_size":100,"webhook_clock_skew_seconds":300}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-BEPUSDT-Key-Id") != "new-key" || r.Header.Get("X-BEPUSDT-Signature") == "" {
			t.Fatalf("new merchant credentials were not used")
		}
		if r.URL.Path == "/api/v1/merchant/capabilities" {
			_, _ = w.Write([]byte(`{"code":"ok","data":{"networks":[{"crypto":"USDT","network":"bsc","network_name":"BSC","trade_type":"usdt.bep20","wallet_count":1,"rpc_endpoint_set":true,"accepting_orders":true}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":"ok","data":{"crypto":"USDT","fiat":"CNY","rate":"6.72","updated_at":1700000000}}`))
	}))
	defer server.Close()
	settings := NewUSDTPaymentSettingsService(repo, invoiceSettingsTestEncryptor{}, config.USDTPaymentConfig{
		APIBase: server.URL, PublicCallbackBaseURL: "https://app.example.test", KeyID: "fallback-key", APISecret: "fallback-secret", Fiat: "CNY", EnabledNetworks: []string{"bsc"},
		OrderTimeoutSeconds: 900, LatePaymentWindowMinutes: 60, RequestTimeoutSeconds: 6, ReconcileIntervalSeconds: 2, ReconcileBatchSize: 100, WebhookClockSkewSeconds: 300,
	}, true)
	input := testUSDTPaymentAdminSettings()
	input.APIBase, input.KeyID, input.APISecret = server.URL, "new-key", "new-secret"
	result, err := settings.TestConnection(context.Background(), input)
	if err != nil || result == nil || result.Rate != "6.72" {
		t.Fatalf("test connection with replacement secret failed: result=%#v err=%v", result, err)
	}
}

func TestClassifyUSDTPaymentConfigErrorForMigratedSecret(t *testing.T) {
	err := classifyUSDTPaymentConfigError(errors.New("decrypt USDT API secret: cipher: message authentication failed"))
	if infraerrors.Reason(err) != "USDT_PAYMENT_SECRET_DECRYPT_FAILED" {
		t.Fatalf("reason = %q", infraerrors.Reason(err))
	}
}
