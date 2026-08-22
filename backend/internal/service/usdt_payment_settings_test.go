package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
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
