package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newProjectMihomoSettingServiceForTest(t *testing.T, raw string) *service.SettingService {
	t.Helper()
	return service.NewSettingService(&serviceSettingRepoStub{
		values: map[string]string{
			service.SettingKeyProjectMihomoSettings: raw,
		},
	}, &config.Config{})
}

type serviceSettingRepoStub struct {
	values map[string]string
}

func (s *serviceSettingRepoStub) Get(_ context.Context, _ string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *serviceSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if v, ok := s.values[key]; ok {
		return v, nil
	}
	return "", service.ErrSettingNotFound
}

func (s *serviceSettingRepoStub) Set(_ context.Context, _, _ string) error {
	panic("unexpected Set call")
}

func (s *serviceSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string, len(keys))
	for _, key := range keys {
		if v, ok := s.values[key]; ok {
			result[key] = v
		}
	}
	return result, nil
}

func (s *serviceSettingRepoStub) SetMultiple(_ context.Context, _ map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *serviceSettingRepoStub) GetAll(_ context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *serviceSettingRepoStub) Delete(_ context.Context, _ string) error {
	panic("unexpected Delete call")
}

func setupAccountCreateRouterWithMihomo(t *testing.T, adminSvc *stubAdminService, rawSettings string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAccountHandler(
		adminSvc,
		newProjectMihomoSettingServiceForTest(t, rawSettings),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	router.POST("/api/v1/admin/accounts", handler.Create)
	router.POST("/api/v1/admin/accounts/batch", handler.BatchCreate)
	router.POST("/api/v1/admin/accounts/data", handler.ImportData)
	return router
}

func defaultProjectMihomoSettingsJSON() string {
	return `{
		"protocol":"socks5h",
		"target_host":"mihomo-sub2api",
		"listener_count":3,
		"listener_ports":[61000,61001,61002],
		"listener_names":["project-mihomo-01","project-mihomo-02","project-mihomo-03"],
		"proxy_name_prefix":"project-mihomo"
	}`
}

func TestAccountHandlerCreate_ProjectMihomoAllocatesLeastLoadedProxy(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.proxyCounts = []service.ProxyWithAccountCount{
		{
			Proxy:        service.Proxy{ID: 11, Name: "project-mihomo-01", Protocol: "socks5h", Host: "mihomo-sub2api", Port: 61000, Status: service.StatusActive},
			AccountCount: 5,
		},
		{
			Proxy:        service.Proxy{ID: 12, Name: "project-mihomo-02", Protocol: "socks5h", Host: "mihomo-sub2api", Port: 61001, Status: service.StatusActive},
			AccountCount: 1,
		},
		{
			Proxy:        service.Proxy{ID: 13, Name: "project-mihomo-03", Protocol: "socks5h", Host: "mihomo-sub2api", Port: 61002, Status: service.StatusActive},
			AccountCount: 3,
		},
	}
	router := setupAccountCreateRouterWithMihomo(t, adminSvc, defaultProjectMihomoSettingsJSON())

	body, err := json.Marshal(map[string]any{
		"name":           "acc-1",
		"platform":       service.PlatformOpenAI,
		"type":           service.AccountTypeOAuth,
		"credentials":    map[string]any{"token": "x"},
		"proxy_provider": proxyProviderProjectMihomo,
		"concurrency":    3,
		"priority":       10,
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, adminSvc.createdAccounts, 1)
	require.NotNil(t, adminSvc.createdAccounts[0].ProxyID)
	require.Equal(t, int64(12), *adminSvc.createdAccounts[0].ProxyID)
}

func TestAccountHandlerBatchCreate_ProjectMihomoBalancesAcrossPool(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.proxyCounts = []service.ProxyWithAccountCount{
		{
			Proxy:        service.Proxy{ID: 21, Name: "project-mihomo-01", Protocol: "socks5h", Host: "mihomo-sub2api", Port: 61000, Status: service.StatusActive},
			AccountCount: 0,
		},
		{
			Proxy:        service.Proxy{ID: 22, Name: "project-mihomo-02", Protocol: "socks5h", Host: "mihomo-sub2api", Port: 61001, Status: service.StatusActive},
			AccountCount: 0,
		},
	}
	router := setupAccountCreateRouterWithMihomo(t, adminSvc, defaultProjectMihomoSettingsJSON())

	body, err := json.Marshal(map[string]any{
		"accounts": []map[string]any{
			{
				"name":           "acc-1",
				"platform":       service.PlatformOpenAI,
				"type":           service.AccountTypeOAuth,
				"credentials":    map[string]any{"token": "a"},
				"proxy_provider": proxyProviderProjectMihomo,
				"concurrency":    3,
				"priority":       10,
			},
			{
				"name":           "acc-2",
				"platform":       service.PlatformOpenAI,
				"type":           service.AccountTypeOAuth,
				"credentials":    map[string]any{"token": "b"},
				"proxy_provider": proxyProviderProjectMihomo,
				"concurrency":    3,
				"priority":       10,
			},
			{
				"name":           "acc-3",
				"platform":       service.PlatformOpenAI,
				"type":           service.AccountTypeOAuth,
				"credentials":    map[string]any{"token": "c"},
				"proxy_provider": proxyProviderProjectMihomo,
				"concurrency":    3,
				"priority":       10,
			},
		},
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, adminSvc.createdAccounts, 3)

	var assigned []int64
	for _, item := range adminSvc.createdAccounts {
		require.NotNil(t, item.ProxyID)
		assigned = append(assigned, *item.ProxyID)
	}
	require.Equal(t, []int64{21, 22, 21}, assigned)
}

func TestImportData_ProjectMihomoOverridesSourceProxyKeyAndBalances(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.proxies = []service.Proxy{
		{ID: 1, Name: "legacy-proxy", Protocol: "socks5", Host: "1.2.3.4", Port: 1080, Username: "u", Password: "p", Status: service.StatusActive},
	}
	adminSvc.proxyCounts = []service.ProxyWithAccountCount{
		{
			Proxy:        service.Proxy{ID: 31, Name: "project-mihomo-01", Protocol: "socks5h", Host: "mihomo-sub2api", Port: 61000, Status: service.StatusActive},
			AccountCount: 0,
		},
		{
			Proxy:        service.Proxy{ID: 32, Name: "project-mihomo-02", Protocol: "socks5h", Host: "mihomo-sub2api", Port: 61001, Status: service.StatusActive},
			AccountCount: 0,
		},
	}
	router := setupAccountCreateRouterWithMihomo(t, adminSvc, defaultProjectMihomoSettingsJSON())

	payload := map[string]any{
		"data": map[string]any{
			"type":    dataType,
			"version": dataVersion,
			"proxies": []map[string]any{
				{
					"proxy_key": "socks5|1.2.3.4|1080|u|p",
					"name":      "legacy-proxy",
					"protocol":  "socks5",
					"host":      "1.2.3.4",
					"port":      1080,
					"username":  "u",
					"password":  "p",
					"status":    "active",
				},
			},
			"accounts": []map[string]any{
				{
					"name":        "acc-a",
					"platform":    service.PlatformOpenAI,
					"type":        service.AccountTypeOAuth,
					"credentials": map[string]any{"token": "a"},
					"proxy_key":   "socks5|1.2.3.4|1080|u|p",
					"concurrency": 3,
					"priority":    10,
				},
				{
					"name":        "acc-b",
					"platform":    service.PlatformOpenAI,
					"type":        service.AccountTypeOAuth,
					"credentials": map[string]any{"token": "b"},
					"proxy_key":   "socks5|1.2.3.4|1080|u|p",
					"concurrency": 3,
					"priority":    10,
				},
			},
		},
		"skip_default_group_bind": true,
		"proxy_provider":          proxyProviderProjectMihomo,
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/accounts/data", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, adminSvc.createdAccounts, 2)
	require.NotNil(t, adminSvc.createdAccounts[0].ProxyID)
	require.NotNil(t, adminSvc.createdAccounts[1].ProxyID)
	require.Equal(t, int64(31), *adminSvc.createdAccounts[0].ProxyID)
	require.Equal(t, int64(32), *adminSvc.createdAccounts[1].ProxyID)
}
