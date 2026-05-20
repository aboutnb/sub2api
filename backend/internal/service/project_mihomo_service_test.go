//go:build unit

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type projectMihomoAdminServiceStub struct {
	createCalls []*CreateProxyInput
	updateCalls []projectMihomoUpdateCall
	deleteCalls []int64
	exists      map[string]bool
	proxies     []Proxy
}

type projectMihomoUpdateCall struct {
	ID    int64
	Input UpdateProxyInput
}

func (s *projectMihomoAdminServiceStub) proxyKey(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func (s *projectMihomoAdminServiceStub) ListUsers(context.Context, int, int, UserListFilters, string, string) ([]User, int64, error) {
	panic("unexpected ListUsers call")
}
func (s *projectMihomoAdminServiceStub) GetUser(context.Context, int64) (*User, error) {
	panic("unexpected GetUser call")
}
func (s *projectMihomoAdminServiceStub) CreateUser(context.Context, *CreateUserInput) (*User, error) {
	panic("unexpected CreateUser call")
}
func (s *projectMihomoAdminServiceStub) UpdateUser(context.Context, int64, *UpdateUserInput) (*User, error) {
	panic("unexpected UpdateUser call")
}
func (s *projectMihomoAdminServiceStub) DeleteUser(context.Context, int64) error {
	panic("unexpected DeleteUser call")
}
func (s *projectMihomoAdminServiceStub) UpdateUserBalance(context.Context, int64, float64, string, string) (*User, error) {
	panic("unexpected UpdateUserBalance call")
}
func (s *projectMihomoAdminServiceStub) BatchUpdateConcurrency(context.Context, []int64, int, string) (int, error) {
	panic("unexpected BatchUpdateConcurrency call")
}
func (s *projectMihomoAdminServiceStub) GetUserAPIKeys(context.Context, int64, int, int, string, string) ([]APIKey, int64, error) {
	panic("unexpected GetUserAPIKeys call")
}
func (s *projectMihomoAdminServiceStub) GetUserUsageStats(context.Context, int64, string) (any, error) {
	panic("unexpected GetUserUsageStats call")
}
func (s *projectMihomoAdminServiceStub) GetUserRPMStatus(context.Context, int64) (*UserRPMStatus, error) {
	panic("unexpected GetUserRPMStatus call")
}
func (s *projectMihomoAdminServiceStub) GetUserBalanceHistory(context.Context, int64, int, int, string) ([]RedeemCode, int64, float64, error) {
	panic("unexpected GetUserBalanceHistory call")
}
func (s *projectMihomoAdminServiceStub) BindUserAuthIdentity(context.Context, int64, AdminBindAuthIdentityInput) (*AdminBoundAuthIdentity, error) {
	panic("unexpected BindUserAuthIdentity call")
}
func (s *projectMihomoAdminServiceStub) ListGroups(context.Context, int, int, string, string, string, *bool, string, string) ([]Group, int64, error) {
	panic("unexpected ListGroups call")
}
func (s *projectMihomoAdminServiceStub) GetAllGroups(context.Context) ([]Group, error) {
	panic("unexpected GetAllGroups call")
}
func (s *projectMihomoAdminServiceStub) GetAllGroupsByPlatform(context.Context, string) ([]Group, error) {
	panic("unexpected GetAllGroupsByPlatform call")
}
func (s *projectMihomoAdminServiceStub) GetGroup(context.Context, int64) (*Group, error) {
	panic("unexpected GetGroup call")
}
func (s *projectMihomoAdminServiceStub) CreateGroup(context.Context, *CreateGroupInput) (*Group, error) {
	panic("unexpected CreateGroup call")
}
func (s *projectMihomoAdminServiceStub) UpdateGroup(context.Context, int64, *UpdateGroupInput) (*Group, error) {
	panic("unexpected UpdateGroup call")
}
func (s *projectMihomoAdminServiceStub) DeleteGroup(context.Context, int64) error {
	panic("unexpected DeleteGroup call")
}
func (s *projectMihomoAdminServiceStub) GetGroupAPIKeys(context.Context, int64, int, int) ([]APIKey, int64, error) {
	panic("unexpected GetGroupAPIKeys call")
}
func (s *projectMihomoAdminServiceStub) GetGroupRateMultipliers(context.Context, int64) ([]UserGroupRateEntry, error) {
	panic("unexpected GetGroupRateMultipliers call")
}
func (s *projectMihomoAdminServiceStub) ClearGroupRateMultipliers(context.Context, int64) error {
	panic("unexpected ClearGroupRateMultipliers call")
}
func (s *projectMihomoAdminServiceStub) BatchSetGroupRateMultipliers(context.Context, int64, []GroupRateMultiplierInput) error {
	panic("unexpected BatchSetGroupRateMultipliers call")
}
func (s *projectMihomoAdminServiceStub) ClearGroupRPMOverrides(context.Context, int64) error {
	panic("unexpected ClearGroupRPMOverrides call")
}
func (s *projectMihomoAdminServiceStub) BatchSetGroupRPMOverrides(context.Context, int64, []GroupRPMOverrideInput) error {
	panic("unexpected BatchSetGroupRPMOverrides call")
}
func (s *projectMihomoAdminServiceStub) UpdateGroupSortOrders(context.Context, []GroupSortOrderUpdate) error {
	panic("unexpected UpdateGroupSortOrders call")
}
func (s *projectMihomoAdminServiceStub) AdminUpdateAPIKeyGroupID(context.Context, int64, *int64) (*AdminUpdateAPIKeyGroupIDResult, error) {
	panic("unexpected AdminUpdateAPIKeyGroupID call")
}
func (s *projectMihomoAdminServiceStub) AdminResetAPIKeyRateLimitUsage(context.Context, int64) (*APIKey, error) {
	panic("unexpected AdminResetAPIKeyRateLimitUsage call")
}
func (s *projectMihomoAdminServiceStub) ReplaceUserGroup(context.Context, int64, int64, int64) (*ReplaceUserGroupResult, error) {
	panic("unexpected ReplaceUserGroup call")
}
func (s *projectMihomoAdminServiceStub) ListAccounts(context.Context, int, int, string, string, string, string, int64, string, string, string) ([]Account, int64, error) {
	panic("unexpected ListAccounts call")
}
func (s *projectMihomoAdminServiceStub) GetAccount(context.Context, int64) (*Account, error) {
	panic("unexpected GetAccount call")
}
func (s *projectMihomoAdminServiceStub) GetAccountsByIDs(context.Context, []int64) ([]*Account, error) {
	panic("unexpected GetAccountsByIDs call")
}
func (s *projectMihomoAdminServiceStub) CreateAccount(context.Context, *CreateAccountInput) (*Account, error) {
	panic("unexpected CreateAccount call")
}
func (s *projectMihomoAdminServiceStub) UpdateAccount(context.Context, int64, *UpdateAccountInput) (*Account, error) {
	panic("unexpected UpdateAccount call")
}
func (s *projectMihomoAdminServiceStub) DeleteAccount(context.Context, int64) error {
	panic("unexpected DeleteAccount call")
}
func (s *projectMihomoAdminServiceStub) RefreshAccountCredentials(context.Context, int64) (*Account, error) {
	panic("unexpected RefreshAccountCredentials call")
}
func (s *projectMihomoAdminServiceStub) ClearAccountError(context.Context, int64) (*Account, error) {
	panic("unexpected ClearAccountError call")
}
func (s *projectMihomoAdminServiceStub) SetAccountError(context.Context, int64, string) error {
	panic("unexpected SetAccountError call")
}
func (s *projectMihomoAdminServiceStub) EnsureOpenAIPrivacy(context.Context, *Account) string {
	panic("unexpected EnsureOpenAIPrivacy call")
}
func (s *projectMihomoAdminServiceStub) EnsureAntigravityPrivacy(context.Context, *Account) string {
	panic("unexpected EnsureAntigravityPrivacy call")
}
func (s *projectMihomoAdminServiceStub) ForceOpenAIPrivacy(context.Context, *Account) string {
	panic("unexpected ForceOpenAIPrivacy call")
}
func (s *projectMihomoAdminServiceStub) ForceAntigravityPrivacy(context.Context, *Account) string {
	panic("unexpected ForceAntigravityPrivacy call")
}
func (s *projectMihomoAdminServiceStub) SetAccountSchedulable(context.Context, int64, bool) (*Account, error) {
	panic("unexpected SetAccountSchedulable call")
}
func (s *projectMihomoAdminServiceStub) BulkUpdateAccounts(context.Context, *BulkUpdateAccountsInput) (*BulkUpdateAccountsResult, error) {
	panic("unexpected BulkUpdateAccounts call")
}
func (s *projectMihomoAdminServiceStub) CheckMixedChannelRisk(context.Context, int64, string, []int64) error {
	panic("unexpected CheckMixedChannelRisk call")
}
func (s *projectMihomoAdminServiceStub) ListProxies(context.Context, int, int, string, string, string, string, string) ([]Proxy, int64, error) {
	panic("unexpected ListProxies call")
}
func (s *projectMihomoAdminServiceStub) ListProxiesWithAccountCount(context.Context, int, int, string, string, string, string, string) ([]ProxyWithAccountCount, int64, error) {
	panic("unexpected ListProxiesWithAccountCount call")
}
func (s *projectMihomoAdminServiceStub) GetAllProxies(context.Context) ([]Proxy, error) {
	out := make([]Proxy, len(s.proxies))
	copy(out, s.proxies)
	return out, nil
}
func (s *projectMihomoAdminServiceStub) GetAllProxiesWithAccountCount(context.Context) ([]ProxyWithAccountCount, error) {
	panic("unexpected GetAllProxiesWithAccountCount call")
}
func (s *projectMihomoAdminServiceStub) GetProxy(context.Context, int64) (*Proxy, error) {
	panic("unexpected GetProxy call")
}
func (s *projectMihomoAdminServiceStub) GetProxiesByIDs(context.Context, []int64) ([]Proxy, error) {
	panic("unexpected GetProxiesByIDs call")
}
func (s *projectMihomoAdminServiceStub) CreateProxy(_ context.Context, input *CreateProxyInput) (*Proxy, error) {
	s.createCalls = append(s.createCalls, input)
	if s.exists == nil {
		s.exists = map[string]bool{}
	}
	s.exists[s.proxyKey(input.Host, input.Port)] = true
	id := int64(len(s.proxies) + 1)
	proxy := Proxy{
		ID:       id,
		Name:     input.Name,
		Protocol: input.Protocol,
		Host:     input.Host,
		Port:     input.Port,
		Status:   StatusActive,
	}
	s.proxies = append(s.proxies, proxy)
	return &proxy, nil
}
func (s *projectMihomoAdminServiceStub) UpdateProxy(_ context.Context, id int64, input *UpdateProxyInput) (*Proxy, error) {
	s.updateCalls = append(s.updateCalls, projectMihomoUpdateCall{ID: id, Input: *input})
	for i := range s.proxies {
		if s.proxies[i].ID != id {
			continue
		}
		if input.Name != "" {
			s.proxies[i].Name = input.Name
		}
		if input.Protocol != "" {
			s.proxies[i].Protocol = input.Protocol
		}
		if input.Host != "" {
			s.proxies[i].Host = input.Host
		}
		if input.Port != 0 {
			s.proxies[i].Port = input.Port
		}
		if input.Status != "" {
			s.proxies[i].Status = input.Status
		}
		return &s.proxies[i], nil
	}
	return nil, fmt.Errorf("proxy %d not found", id)
}
func (s *projectMihomoAdminServiceStub) DeleteProxy(_ context.Context, id int64) error {
	s.deleteCalls = append(s.deleteCalls, id)
	for i := range s.proxies {
		if s.proxies[i].ID != id {
			continue
		}
		s.proxies = append(s.proxies[:i], s.proxies[i+1:]...)
		return nil
	}
	return nil
}
func (s *projectMihomoAdminServiceStub) BatchDeleteProxies(context.Context, []int64) (*ProxyBatchDeleteResult, error) {
	panic("unexpected BatchDeleteProxies call")
}
func (s *projectMihomoAdminServiceStub) GetProxyAccounts(context.Context, int64) ([]ProxyAccountSummary, error) {
	panic("unexpected GetProxyAccounts call")
}
func (s *projectMihomoAdminServiceStub) CheckProxyExists(_ context.Context, host string, port int, username, password string) (bool, error) {
	return s.exists[s.proxyKey(host, port)], nil
}
func (s *projectMihomoAdminServiceStub) TestProxy(context.Context, int64) (*ProxyTestResult, error) {
	panic("unexpected TestProxy call")
}
func (s *projectMihomoAdminServiceStub) CheckProxyQuality(context.Context, int64) (*ProxyQualityCheckResult, error) {
	panic("unexpected CheckProxyQuality call")
}
func (s *projectMihomoAdminServiceStub) ListRedeemCodes(context.Context, int, int, string, string, string, string, string) ([]RedeemCode, int64, error) {
	panic("unexpected ListRedeemCodes call")
}
func (s *projectMihomoAdminServiceStub) GetRedeemCode(context.Context, int64) (*RedeemCode, error) {
	panic("unexpected GetRedeemCode call")
}
func (s *projectMihomoAdminServiceStub) GenerateRedeemCodes(context.Context, *GenerateRedeemCodesInput) ([]RedeemCode, error) {
	panic("unexpected GenerateRedeemCodes call")
}
func (s *projectMihomoAdminServiceStub) DeleteRedeemCode(context.Context, int64) error {
	panic("unexpected DeleteRedeemCode call")
}
func (s *projectMihomoAdminServiceStub) BatchDeleteRedeemCodes(context.Context, []int64) (int64, error) {
	panic("unexpected BatchDeleteRedeemCodes call")
}
func (s *projectMihomoAdminServiceStub) ExpireRedeemCode(context.Context, int64) (*RedeemCode, error) {
	panic("unexpected ExpireRedeemCode call")
}
func (s *projectMihomoAdminServiceStub) ResetAccountQuota(context.Context, int64) error {
	panic("unexpected ResetAccountQuota call")
}

func TestProjectMihomoGetSettingsDefaults(t *testing.T) {
	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})

	settings, err := svc.GetSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, "socks5h", settings.Protocol)
	require.Equal(t, 41001, settings.StartPort)
	require.Equal(t, 4, settings.ListenerCount)
	require.Len(t, settings.ListenerRegions, 4)
	require.Empty(t, settings.SubscriptionURLs)
}

func TestProjectMihomoBuildProxies(t *testing.T) {
	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})

	proxies := svc.buildProxies(&ProjectMihomoSettings{
		Protocol:        "socks5h",
		TargetHost:      "mihomo-sub2api",
		StartPort:       41001,
		ListenerCount:   3,
		ProxyNamePrefix: "project-mihomo",
	})
	require.Len(t, proxies, 3)
	require.Equal(t, "project-mihomo-01", proxies[0].Name)
	require.Equal(t, 41003, proxies[2].Port)
}

func TestBuildProjectMihomoProviderRefs(t *testing.T) {
	refs := buildProjectMihomoProviderRefs(&ProjectMihomoSettings{
		SubscriptionURLs: []string{"https://a.example/sub", "https://b.example/sub"},
	})

	require.Len(t, refs, 2)
	require.Equal(t, "project-subscription-01", refs[0].Name)
	require.Equal(t, "./providers/project-subscription-01.yaml", refs[0].Path)
	require.Equal(t, "project-subscription-02", refs[1].Name)
}

func TestProjectMihomoAssignProviderNodes(t *testing.T) {
	selected := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"placeholder","alive":false},{"name":"node-a","alive":true},{"name":"node-b","alive":true}]}}}`))
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/proxies/"):
			var payload struct {
				Name string `json:"name"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			selected[strings.TrimPrefix(r.URL.Path, "/proxies/")] = payload.Name
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	settings := &ProjectMihomoSettings{
		ControllerURL:   server.URL,
		SubscriptionURL: "https://example.com/sub",
	}
	proxies := []ProjectMihomoProxy{{Name: "project-mihomo-01"}, {Name: "project-mihomo-02"}, {Name: "project-mihomo-03"}}

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies)
	require.NoError(t, err)
	require.Equal(t, 3, assigned)
	require.Equal(t, "node-a", selected["project-mihomo-01"])
	require.Equal(t, "node-b", selected["project-mihomo-02"])
	require.Equal(t, "node-a", selected["project-mihomo-03"])
}

func TestProjectMihomoAssignProviderNodesByRegion(t *testing.T) {
	selected := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"日本-01","alive":true},{"name":"美国-01","alive":true},{"name":"香港-01","alive":true}]}}}`))
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/proxies/"):
			var payload struct {
				Name string `json:"name"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			selected[strings.TrimPrefix(r.URL.Path, "/proxies/")] = payload.Name
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	settings := &ProjectMihomoSettings{
		ControllerURL:   server.URL,
		SubscriptionURL: "https://example.com/sub",
		ListenerRegions: []string{"japan", "usa", ""},
	}
	proxies := []ProjectMihomoProxy{{Name: "project-mihomo-01"}, {Name: "project-mihomo-02"}, {Name: "project-mihomo-03"}}

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies)
	require.NoError(t, err)
	require.Equal(t, 3, assigned)
	require.Equal(t, "日本-01", selected["project-mihomo-01"])
	require.Equal(t, "美国-01", selected["project-mihomo-02"])
}

func TestProjectMihomoAssignProviderNodesAcrossMultipleProviders(t *testing.T) {
	selected := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription-01":{"proxies":[{"name":"日本-01","alive":true}]},"project-subscription-02":{"proxies":[{"name":"美国-01","alive":true}]}}}`))
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/proxies/"):
			var payload struct {
				Name string `json:"name"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			selected[strings.TrimPrefix(r.URL.Path, "/proxies/")] = payload.Name
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	settings := &ProjectMihomoSettings{
		ControllerURL:    server.URL,
		SubscriptionURLs: []string{"https://a.example/sub", "https://b.example/sub"},
	}
	proxies := []ProjectMihomoProxy{{Name: "project-mihomo-01"}, {Name: "project-mihomo-02"}}

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies)
	require.NoError(t, err)
	require.Equal(t, 2, assigned)
	require.Equal(t, "日本-01", selected["project-mihomo-01"])
	require.Equal(t, "美国-01", selected["project-mihomo-02"])
}

func TestExtractProjectMihomoRegions(t *testing.T) {
	regions := extractProjectMihomoRegions([]string{
		"日本01[X中转1]x2.0",
		"美国01[X中转1]x2.0",
		"香港[X中转1]x2.0",
		"新加坡[直连]x0.8",
		"美国02[X中转1]x2.0",
		"",
	})

	require.Equal(t, []string{"日本", "美国", "香港", "新加坡"}, regions)
}

func TestProjectMihomoGetStatusResolvesLegacyRegionAliases(t *testing.T) {
	settings := ProjectMihomoSettings{
		SubscriptionURL: "https://example.com/sub",
		SubscriptionUA:  "sub2api/mihomo",
		UpdateInterval:  3600,
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       41001,
		ListenerCount:   4,
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: []string{"japan", "usa", "hongkong", ""},
	}
	raw, err := json.Marshal(settings)
	require.NoError(t, err)

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{
		SettingKeyProjectMihomoSettings: string(raw),
	}}, &projectMihomoAdminServiceStub{})

	t.Setenv("DATA_DIR", t.TempDir())
	cacheDir := svc.configDir()
	require.NoError(t, os.MkdirAll(filepath.Join(cacheDir, "providers"), 0o755))
	require.NoError(t, os.WriteFile(
		svc.providerCachePath(),
		[]byte("proxies:\n  - name: 日本01[X中转1]x2.0\n    server: jp.example.com\n  - name: 美国01[X中转1]x2.0\n    server: us.example.com\n  - name: 香港01[X中转1]x2.0\n    server: hk.example.com\n"),
		0o644,
	))

	status, err := svc.GetStatus(context.Background())
	require.NoError(t, err)
	require.Equal(t, []string{"日本", "美国", "香港"}, status.AvailableRegions)
	require.Equal(t, []string{"日本", "美国", "香港", ""}, status.Settings.ListenerRegions)
}

func TestProjectMihomoAvailableRegionsAcrossMultipleProviderCaches(t *testing.T) {
	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})

	t.Setenv("DATA_DIR", t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Join(svc.configDir(), "providers"), 0o755))
	require.NoError(t, os.WriteFile(
		svc.providerCachePathFor("./providers/project-subscription-01.yaml"),
		[]byte("proxies:\n  - name: 日本01[X中转1]x2.0\n    server: jp.example.com\n"),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		svc.providerCachePathFor("./providers/project-subscription-02.yaml"),
		[]byte("proxies:\n  - name: 美国01[X中转1]x2.0\n    server: us.example.com\n"),
		0o644,
	))

	regions := svc.availableRegions(&ProjectMihomoSettings{
		SubscriptionURLs: []string{"https://a.example/sub", "https://b.example/sub"},
	})
	require.Equal(t, []string{"日本", "美国"}, regions)
}

func TestProjectMihomoSyncProxyRowsUpdatesAndCleansManagedDuplicates(t *testing.T) {
	stub := &projectMihomoAdminServiceStub{
		proxies: []Proxy{
			{ID: 1, Name: "project-mihomo-01", Protocol: "socks5h", Host: "127.0.0.1", Port: 41001, Status: StatusActive},
			{ID: 2, Name: "project-mihomo-04", Protocol: "http", Host: "127.0.0.1", Port: 41004, Status: StatusActive},
			{ID: 3, Name: "project-mihomo-04", Protocol: "socks5h", Host: "127.0.0.1", Port: 41004, Status: StatusActive},
			{ID: 4, Name: "project-mihomo-05", Protocol: "socks5h", Host: "127.0.0.1", Port: 41005, Status: StatusActive},
		},
	}
	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, stub)
	settings := &ProjectMihomoSettings{
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       41001,
		ListenerCount:   4,
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: make([]string, 4),
	}

	created, reused, err := svc.syncProxyRows(context.Background(), settings, svc.buildProxies(settings))
	require.NoError(t, err)
	require.Equal(t, 2, created)
	require.Equal(t, 2, reused)

	require.Len(t, stub.proxies, 4)
	for i := range stub.proxies {
		require.Equal(t, "socks5h", stub.proxies[i].Protocol)
		require.True(t, strings.HasPrefix(stub.proxies[i].Name, "project-mihomo-"))
	}

	remaining := map[string]struct{}{}
	for i := range stub.proxies {
		remaining[stub.proxies[i].Name] = struct{}{}
	}
	require.Contains(t, remaining, "project-mihomo-01")
	require.Contains(t, remaining, "project-mihomo-02")
	require.Contains(t, remaining, "project-mihomo-03")
	require.Contains(t, remaining, "project-mihomo-04")
	require.NotContains(t, remaining, "project-mihomo-05")
	require.Contains(t, stub.deleteCalls, int64(2))
	require.Contains(t, stub.deleteCalls, int64(4))
}

func TestParseProjectMihomoSubscriptionNodeNames_Base64URIsSkipsPlaceholderHost(t *testing.T) {
	content := "dmxlc3M6Ly91c2VyQERvbnQudXNlLnRoaXMubm9kZTo4ODg4IyVFNSU4OSVBOSVFNCVCRCU5OSVFNiVCNSU4MSVFOSU4NyU4Rg0Kc3M6Ly9ZMmhoWTJoaE1qQXRAcmVhbC5leGFtcGxlLmNvbToxMTAxMiMlRTklQTYlOTklRTYlQjglQUY="
	names := parseProjectMihomoSubscriptionNodeNames([]byte(content))
	require.Equal(t, []string{"香港"}, names)
}
