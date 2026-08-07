package service

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type invoiceSettingsTestRepo struct {
	mu   sync.Mutex
	data map[string]string
}

func newInvoiceSettingsTestRepo() *invoiceSettingsTestRepo {
	return &invoiceSettingsTestRepo{data: make(map[string]string)}
}

func (r *invoiceSettingsTestRepo) Get(_ context.Context, key string) (*Setting, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value, ok := r.data[key]
	if !ok {
		return nil, ErrSettingNotFound
	}
	return &Setting{Key: key, Value: value}, nil
}

func (r *invoiceSettingsTestRepo) GetValue(_ context.Context, key string) (string, error) {
	setting, err := r.Get(context.Background(), key)
	if err != nil {
		return "", err
	}
	return setting.Value, nil
}

func (r *invoiceSettingsTestRepo) Set(_ context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
	return nil
}

func (r *invoiceSettingsTestRepo) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := r.data[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (r *invoiceSettingsTestRepo) SetMultiple(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if err := r.Set(ctx, key, value); err != nil {
			return err
		}
	}
	return nil
}

func (r *invoiceSettingsTestRepo) GetAll(_ context.Context) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	values := make(map[string]string, len(r.data))
	for key, value := range r.data {
		values[key] = value
	}
	return values, nil
}

func (r *invoiceSettingsTestRepo) Delete(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}

type invoiceSettingsTestEncryptor struct{}

func (invoiceSettingsTestEncryptor) Encrypt(value string) (string, error) {
	return "encrypted:" + value, nil
}

func (invoiceSettingsTestEncryptor) Decrypt(value string) (string, error) {
	if !strings.HasPrefix(value, "encrypted:") {
		return "", ErrSettingNotFound
	}
	return strings.TrimPrefix(value, "encrypted:"), nil
}

func TestInvoiceSettingsFallbackDoesNotExposeSecret(t *testing.T) {
	ctx := context.Background()
	service := NewInvoiceSettingsService(
		newInvoiceSettingsTestRepo(),
		invoiceSettingsTestEncryptor{},
		config.InvoiceIntegrationConfig{
			Enabled: true, BaseURL: "https://invoice.example.test/", ClientID: "fallback-client", ClientSecret: "fallback-secret", TimeoutSeconds: 20,
		},
		true,
	)

	admin, err := service.GetAdminSettings(ctx)
	if err != nil {
		t.Fatalf("get admin settings: %v", err)
	}
	if admin.ClientSecret != "" || !admin.ClientSecretConfigured {
		t.Fatalf("secret leaked or configured state lost: %#v", admin)
	}
	effective, err := service.EffectiveConfig(ctx)
	if err != nil {
		t.Fatalf("get effective config: %v", err)
	}
	if effective.ClientSecret != "fallback-secret" || effective.BaseURL != "https://invoice.example.test" {
		t.Fatalf("unexpected fallback config: %#v", effective)
	}
}

func TestInvoiceSettingsEncryptsAndPreservesClientSecret(t *testing.T) {
	ctx := context.Background()
	repo := newInvoiceSettingsTestRepo()
	service := NewInvoiceSettingsService(repo, invoiceSettingsTestEncryptor{}, config.InvoiceIntegrationConfig{}, true)

	_, err := service.Update(ctx, InvoiceAdminSettings{
		Enabled: true, BaseURL: "https://invoice.example.test", ClientID: "client-a", ClientSecret: "secret-a", TimeoutSeconds: 15,
	})
	if err != nil {
		t.Fatalf("save invoice settings: %v", err)
	}
	var stored invoiceStoredSettings
	if err := json.Unmarshal([]byte(repo.data[settingKeyInvoiceIntegrationConfig]), &stored); err != nil {
		t.Fatalf("decode stored settings: %v", err)
	}
	if stored.ClientSecret != "encrypted:secret-a" {
		t.Fatalf("client secret was not encrypted: %q", stored.ClientSecret)
	}

	_, err = service.Update(ctx, InvoiceAdminSettings{
		Enabled: true, BaseURL: "https://invoice.example.test/v2", ClientID: "client-b", TimeoutSeconds: 25,
	})
	if err != nil {
		t.Fatalf("update without replacing secret: %v", err)
	}
	effective, err := service.EffectiveConfig(ctx)
	if err != nil {
		t.Fatalf("get effective config: %v", err)
	}
	if effective.ClientSecret != "secret-a" || effective.ClientID != "client-b" || effective.TimeoutSeconds != 25 {
		t.Fatalf("unexpected effective config: %#v", effective)
	}
}

func TestInvoiceSettingsRejectsIncompleteEnabledConfig(t *testing.T) {
	service := NewInvoiceSettingsService(newInvoiceSettingsTestRepo(), invoiceSettingsTestEncryptor{}, config.InvoiceIntegrationConfig{}, true)
	_, err := service.Update(context.Background(), InvoiceAdminSettings{Enabled: true, BaseURL: "https://invoice.example.test", TimeoutSeconds: 15})
	if err == nil {
		t.Fatal("expected incomplete enabled config to be rejected")
	}
}

func TestInvoiceSettingsRequiresFixedEncryptionKeyForNewSecret(t *testing.T) {
	service := NewInvoiceSettingsService(newInvoiceSettingsTestRepo(), invoiceSettingsTestEncryptor{}, config.InvoiceIntegrationConfig{}, false)
	_, err := service.Update(context.Background(), InvoiceAdminSettings{
		Enabled: true, BaseURL: "https://invoice.example.test", ClientID: "client", ClientSecret: "secret", TimeoutSeconds: 15,
	})
	if err != ErrInvoiceSecretEncryptionKeyNotConfigured {
		t.Fatalf("error = %v, want fixed encryption key error", err)
	}
}
