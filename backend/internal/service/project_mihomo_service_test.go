//go:build unit

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type projectMihomoAdminServiceStub struct {
	createCalls     []*CreateProxyInput
	updateCalls     []projectMihomoUpdateCall
	deleteCalls     []int64
	bulkUpdateCalls []*BulkUpdateAccountsInput
	exists          map[string]bool
	proxies         []Proxy
	proxyAccounts   map[int64][]ProxyAccountSummary
}

type projectMihomoUpdateCall struct {
	ID    int64
	Input UpdateProxyInput
}

type projectMihomoSettingRepoStub struct {
	values map[string]string
}

func (s *projectMihomoSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *projectMihomoSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if val, ok := s.values[key]; ok {
		return val, nil
	}
	return "", ErrSettingNotFound
}

func (s *projectMihomoSettingRepoStub) Set(_ context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *projectMihomoSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if val, ok := s.values[key]; ok {
			out[key] = val
		}
	}
	return out, nil
}

func (s *projectMihomoSettingRepoStub) SetMultiple(_ context.Context, values map[string]string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	for key, value := range values {
		s.values[key] = value
	}
	return nil
}

func (s *projectMihomoSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return s.values, nil
}

func (s *projectMihomoSettingRepoStub) Delete(_ context.Context, key string) error {
	delete(s.values, key)
	return nil
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
func (s *projectMihomoAdminServiceStub) BulkUpdateAccounts(_ context.Context, input *BulkUpdateAccountsInput) (*BulkUpdateAccountsResult, error) {
	s.bulkUpdateCalls = append(s.bulkUpdateCalls, input)
	if input.ProxyID != nil && *input.ProxyID == 0 {
		for proxyID, accounts := range s.proxyAccounts {
			filtered := accounts[:0]
			for _, account := range accounts {
				remove := false
				for _, id := range input.AccountIDs {
					if account.ID == id {
						remove = true
						break
					}
				}
				if !remove {
					filtered = append(filtered, account)
				}
			}
			s.proxyAccounts[proxyID] = filtered
		}
	}
	result := &BulkUpdateAccountsResult{
		Success:    len(input.AccountIDs),
		SuccessIDs: append([]int64(nil), input.AccountIDs...),
	}
	return result, nil
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
	if len(s.proxyAccounts[id]) > 0 {
		return ErrProxyInUse
	}
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
func (s *projectMihomoAdminServiceStub) GetProxyAccounts(_ context.Context, proxyID int64) ([]ProxyAccountSummary, error) {
	if s.proxyAccounts == nil {
		return nil, nil
	}
	out := make([]ProxyAccountSummary, len(s.proxyAccounts[proxyID]))
	copy(out, s.proxyAccounts[proxyID])
	return out, nil
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
	require.Equal(t, 61000, settings.StartPort)
	require.Equal(t, 4, settings.ListenerCount)
	require.Equal(t, []int{61000, 61001, 61002, 61003}, settings.ListenerPorts)
	require.Equal(t, []string{"project-mihomo-01", "project-mihomo-02", "project-mihomo-03", "project-mihomo-04"}, settings.ListenerNames)
	require.Len(t, settings.ListenerRegions, 4)
	require.Empty(t, settings.SubscriptionURLs)
	require.Empty(t, settings.SubscriptionNames)
	require.False(t, settings.NodeExcludeEnabled)
	require.Contains(t, settings.NodeExcludeKeywords, "香港")
	require.Contains(t, settings.NodeExcludeKeywords, "台湾")
}

func TestProjectMihomoNormalizeRewritesLoopbackControllerInsideContainer(t *testing.T) {
	t.Setenv("PROJECT_MIHOMO_CONTAINER_RUNTIME", "true")

	settings := ProjectMihomoSettings{
		TargetHost:    "127.0.0.1",
		ControllerURL: "http://127.0.0.1:9097",
	}

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	svc.normalize(&settings)

	require.Equal(t, "mihomo-sub2api", settings.TargetHost)
	require.Equal(t, "http://mihomo-sub2api:9097", settings.ControllerURL)
}

func TestProjectMihomoBuildProxies(t *testing.T) {
	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})

	proxies := svc.buildProxies(&ProjectMihomoSettings{
		Protocol:        "socks5h",
		TargetHost:      "mihomo-sub2api",
		StartPort:       61000,
		ListenerCount:   3,
		ProxyNamePrefix: "project-mihomo",
	})
	require.Len(t, proxies, 3)
	require.Equal(t, "project-mihomo-01", proxies[0].Name)
	require.Equal(t, 61002, proxies[2].Port)
}

func TestProjectMihomoBuildProxiesUsesStoredPortsAndNames(t *testing.T) {
	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})

	proxies := svc.buildProxies(&ProjectMihomoSettings{
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   2,
		ListenerPorts:   []int{61000, 61002},
		ListenerNames:   []string{"project-mihomo-01", "project-mihomo-03"},
		ProxyNamePrefix: "project-mihomo",
	})
	require.Equal(t, []ProjectMihomoProxy{
		{Name: "project-mihomo-01", Protocol: "socks5h", Host: "127.0.0.1", Port: 61000},
		{Name: "project-mihomo-03", Protocol: "socks5h", Host: "127.0.0.1", Port: 61002},
	}, proxies)
}

func TestProjectMihomoRenderConfigUsesAutoRouteFilters(t *testing.T) {
	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	settings := &ProjectMihomoSettings{
		SubscriptionURLs:    []string{"https://a.example/sub", "https://b.example/sub"},
		SubscriptionUA:      "sub2api/mihomo",
		Protocol:            "socks5h",
		TargetHost:          "127.0.0.1",
		StartPort:           61000,
		ListenerCount:       2,
		ListenerPorts:       []int{61000, 61001},
		ListenerNames:       []string{"project-mihomo-01", "project-mihomo-02"},
		ControllerURL:       "http://127.0.0.1:9097",
		ProxyNamePrefix:     "project-mihomo",
		ListenerRegions:     []string{"japan", projectMihomoNodeKey("project-subscription-02", "美国-01")},
		AutoRouteEnabled:    true,
		AutoRouteTolerance:  120,
		AutoRouteInterval:   180,
		NodeExcludeEnabled:  true,
		NodeExcludeKeywords: copyProjectMihomoDefaultNodeExcludeKeywords(),
	}

	content, err := svc.renderConfig(settings)
	require.NoError(t, err)

	var payload struct {
		ProxyProviders map[string]map[string]any `yaml:"proxy-providers"`
		ProxyGroups    []map[string]any          `yaml:"proxy-groups"`
	}
	require.NoError(t, yaml.Unmarshal(content, &payload))
	require.Contains(t, fmt.Sprint(payload.ProxyProviders["project-subscription-01"]["exclude-filter"]), "香港")
	require.Len(t, payload.ProxyGroups, 2)

	require.Equal(t, "url-test", payload.ProxyGroups[0]["type"])
	require.Equal(t, []any{"project-subscription-01", "project-subscription-02"}, payload.ProxyGroups[0]["use"])
	require.Equal(t, 120, payload.ProxyGroups[0]["tolerance"])
	require.Equal(t, 180, payload.ProxyGroups[0]["interval"])
	require.Contains(t, fmt.Sprint(payload.ProxyGroups[0]["filter"]), "japan")

	require.Equal(t, "url-test", payload.ProxyGroups[1]["type"])
	require.Equal(t, []any{"project-subscription-02"}, payload.ProxyGroups[1]["use"])
	require.Equal(t, "^美国-01$", payload.ProxyGroups[1]["filter"])
}

func TestBuildProjectMihomoProviderRefs(t *testing.T) {
	refs := buildProjectMihomoProviderRefs(&ProjectMihomoSettings{
		SubscriptionURLs: []string{"https://a.example/sub", "https://b.example/sub"},
	})

	require.Len(t, refs, 2)
	require.Equal(t, "project-subscription-01", refs[0].Name)
	require.Equal(t, "./providers/project-subscription-01.yaml", refs[0].Path)
	require.Equal(t, "project-subscription-02", refs[1].Name)
	require.Equal(t, "./providers/project-subscription-02.yaml", refs[1].Path)
	require.Equal(t, "https://a.example/sub", refs[0].URL)
	require.Equal(t, "https://b.example/sub", refs[1].URL)
}

func TestProjectMihomoSetSettingsCachesAndCleansProviderFiles(t *testing.T) {
	served := map[string]string{
		"/sub-a": "proxies:\n  - name: 日本-01\n    server: jp.example.com\n",
		"/sub-b": "proxies:\n  - name: 美国-01\n    server: us.example.com\n",
		"/sub-c": "proxies:\n  - name: 香港-01\n    server: hk.example.com\n",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, ok := served[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	t.Setenv("DATA_DIR", t.TempDir())
	repo := &projectMihomoSettingRepoStub{values: map[string]string{}}
	svc := NewProjectMihomoService(repo, &projectMihomoAdminServiceStub{})
	legacyPath := svc.providerCachePath()
	require.NoError(t, os.MkdirAll(filepath.Dir(legacyPath), 0o755))
	require.NoError(t, os.WriteFile(legacyPath, []byte("legacy"), 0o644))

	_, err := svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionURLs: []string{server.URL + "/sub-a", server.URL + "/sub-b"},
		SubscriptionUA:   "sub2api/mihomo",
		Protocol:         "socks5h",
		TargetHost:       "127.0.0.1",
		StartPort:        61000,
		ListenerCount:    2,
		ControllerURL:    "http://127.0.0.1:9097",
		ProxyNamePrefix:  "project-mihomo",
	})
	require.NoError(t, err)
	firstPath := svc.providerCachePathFor("./providers/project-subscription-01.yaml")
	secondPath := svc.providerCachePathFor("./providers/project-subscription-02.yaml")
	content, err := os.ReadFile(firstPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "日本-01")
	content, err = os.ReadFile(secondPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "美国-01")
	_, err = os.Stat(legacyPath)
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionURLs: []string{server.URL + "/sub-c"},
		SubscriptionUA:   "sub2api/mihomo",
		Protocol:         "socks5h",
		TargetHost:       "127.0.0.1",
		StartPort:        61000,
		ListenerCount:    2,
		ControllerURL:    "http://127.0.0.1:9097",
		ProxyNamePrefix:  "project-mihomo",
	})
	require.NoError(t, err)
	content, err = os.ReadFile(svc.providerCachePath())
	require.NoError(t, err)
	require.Contains(t, string(content), "香港-01")
	_, err = os.Stat(firstPath)
	require.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(secondPath)
	require.ErrorIs(t, err, os.ErrNotExist)
	content, err = os.ReadFile(legacyPath)
	require.NoError(t, err)
	require.Contains(t, string(content), "香港-01")

	_, err = svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionURLs: nil,
		SubscriptionUA:   "sub2api/mihomo",
		Protocol:         "socks5h",
		TargetHost:       "127.0.0.1",
		StartPort:        61000,
		ListenerCount:    2,
		ControllerURL:    "http://127.0.0.1:9097",
		ProxyNamePrefix:  "project-mihomo",
	})
	require.NoError(t, err)
	_, err = os.Stat(legacyPath)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestProjectMihomoSetSettingsReturnsBadRequestWhenSubscriptionFetchFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
	}))
	defer server.Close()

	t.Setenv("DATA_DIR", t.TempDir())
	svc := NewProjectMihomoService(&projectMihomoSettingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	_, err := svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionURL: server.URL + "/expired",
		SubscriptionUA:  "sub2api/mihomo",
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   1,
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
	})
	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err))
	require.ErrorContains(t, err, "PROJECT_MIHOMO_SUBSCRIPTION_FETCH_FAILED")
}

func TestProjectMihomoSetSettingsAllowsEmptySubscriptions(t *testing.T) {
	repo := &projectMihomoSettingRepoStub{values: map[string]string{}}
	svc := NewProjectMihomoService(repo, &projectMihomoAdminServiceStub{})

	settings, err := svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionURLs: nil,
		SubscriptionUA:   "sub2api/mihomo",
		Protocol:         "socks5h",
		TargetHost:       "127.0.0.1",
		StartPort:        61000,
		ListenerCount:    2,
		ControllerURL:    "http://127.0.0.1:9097",
		ProxyNamePrefix:  "project-mihomo",
	})
	require.NoError(t, err)
	require.Empty(t, settings.SubscriptionURL)
	require.Empty(t, settings.SubscriptionURLs)
	require.Empty(t, settings.SubscriptionNames)

	var saved ProjectMihomoSettings
	require.NoError(t, json.Unmarshal([]byte(repo.values[SettingKeyProjectMihomoSettings]), &saved))
	require.Empty(t, saved.SubscriptionURL)
	require.Empty(t, saved.SubscriptionURLs)
	require.Empty(t, saved.SubscriptionNames)
}

func TestProjectMihomoNormalizeSubscriptionNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("proxies:\n  - name: 日本-01\n    server: jp.example.com\n"))
	}))
	defer server.Close()

	t.Setenv("DATA_DIR", t.TempDir())
	svc := NewProjectMihomoService(&projectMihomoSettingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})

	settings, err := svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionURLs:  []string{server.URL + "/sub-a", server.URL + "/sub-b"},
		SubscriptionNames: []string{"  Japan  ", "", "ignored"},
		SubscriptionUA:    "sub2api/mihomo",
		Protocol:          "socks5h",
		TargetHost:        "127.0.0.1",
		StartPort:         61000,
		ListenerCount:     2,
		ControllerURL:     "http://127.0.0.1:9097",
		ProxyNamePrefix:   "project-mihomo",
	})

	require.NoError(t, err)
	require.Equal(t, []string{"Japan", ""}, settings.SubscriptionNames)
}

func TestProjectMihomoNormalizeNodeExcludeKeywords(t *testing.T) {
	svc := NewProjectMihomoService(&projectMihomoSettingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})

	settings, err := svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionUA:      "sub2api/mihomo",
		Protocol:            "socks5h",
		TargetHost:          "127.0.0.1",
		StartPort:           61000,
		ListenerCount:       2,
		ControllerURL:       "http://127.0.0.1:9097",
		ProxyNamePrefix:     "project-mihomo",
		NodeExcludeEnabled:  true,
		NodeExcludeKeywords: []string{" 香港 ", "", "Hong Kong", "hong-kong", "台湾"},
	})

	require.NoError(t, err)
	require.True(t, settings.NodeExcludeEnabled)
	require.Equal(t, []string{"香港", "Hong Kong", "台湾"}, settings.NodeExcludeKeywords)
}

func TestProjectMihomoSetSettingsAppendsListenersWithoutResettingExisting(t *testing.T) {
	previousSettings := ProjectMihomoSettings{
		SubscriptionUA:  "sub2api/mihomo",
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   2,
		ListenerPorts:   []int{61000, 61002},
		ListenerNames:   []string{"project-mihomo-01", "project-mihomo-03"},
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: []string{"日本", "美国"},
	}
	raw, err := json.Marshal(previousSettings)
	require.NoError(t, err)
	repo := &projectMihomoSettingRepoStub{values: map[string]string{
		SettingKeyProjectMihomoSettings: string(raw),
	}}
	svc := NewProjectMihomoService(repo, &projectMihomoAdminServiceStub{})

	settings, err := svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionUA:  "sub2api/mihomo",
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   3,
		ListenerPorts:   []int{61000, 61002},
		ListenerNames:   []string{"project-mihomo-01", "project-mihomo-03"},
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: []string{"日本", "美国"},
	})

	require.NoError(t, err)
	require.Equal(t, []int{61000, 61002, 61001}, settings.ListenerPorts)
	require.Equal(t, []string{"project-mihomo-01", "project-mihomo-03", "project-mihomo-04"}, settings.ListenerNames)
	require.Equal(t, []string{"日本", "美国", ""}, settings.ListenerRegions)
}

func TestProjectMihomoSetSettingsReusesCachedProviderFileWhenURLUnchanged(t *testing.T) {
	t.Setenv("DATA_DIR", t.TempDir())
	sourceURL := "https://example.com/sub"
	previousSettings := ProjectMihomoSettings{
		SubscriptionURL:  sourceURL,
		SubscriptionURLs: []string{sourceURL},
		SubscriptionUA:   "sub2api/mihomo",
		Protocol:         "socks5h",
		TargetHost:       "127.0.0.1",
		StartPort:        61000,
		ListenerCount:    2,
		ControllerURL:    "http://127.0.0.1:9097",
		ProxyNamePrefix:  "project-mihomo",
	}
	raw, err := json.Marshal(previousSettings)
	require.NoError(t, err)
	repo := &projectMihomoSettingRepoStub{values: map[string]string{
		SettingKeyProjectMihomoSettings: string(raw),
	}}
	svc := NewProjectMihomoService(repo, &projectMihomoAdminServiceStub{})

	require.NoError(t, os.MkdirAll(svc.providerDir(), 0o755))
	oldPath := svc.providerCachePath()
	require.NoError(t, os.WriteFile(oldPath, []byte("proxies:\n  - name: 日本-01\n    server: jp.example.com\n"), 0o644))

	_, err = svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionURLs: []string{sourceURL, "https://example.com/sub-2"},
		SubscriptionUA:   "sub2api/mihomo",
		Protocol:         "socks5h",
		TargetHost:       "127.0.0.1",
		StartPort:        61000,
		ListenerCount:    2,
		ControllerURL:    "http://127.0.0.1:9097",
		ProxyNamePrefix:  "project-mihomo",
	})

	require.Error(t, err)
	require.True(t, infraerrors.IsBadRequest(err))

	newPath := svc.providerCachePathFor("./providers/project-subscription-01.yaml")
	content, readErr := os.ReadFile(newPath)
	require.NoError(t, readErr)
	require.Contains(t, string(content), "日本-01")
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

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies, nil, 0)
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

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies, nil, 0)
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

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies, nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, assigned)
	require.Equal(t, "日本-01", selected["project-mihomo-01"])
	require.Equal(t, "美国-01", selected["project-mihomo-02"])
}

func TestProjectMihomoAssignProviderNodesPreservesExistingPortsWhenAdding(t *testing.T) {
	selected := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"日本-01","alive":true},{"name":"美国-01","alive":true},{"name":"新加坡-01","alive":true}]}}}`))
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

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies, []string{"日本-01", "美国-01"}, 2)
	require.NoError(t, err)
	require.Equal(t, 3, assigned)
	require.Equal(t, "日本-01", selected["project-mihomo-01"])
	require.Equal(t, "美国-01", selected["project-mihomo-02"])
	require.Equal(t, "日本-01", selected["project-mihomo-03"])
}

func TestProjectMihomoAssignProviderNodesAssignsFirstSyncPorts(t *testing.T) {
	selected := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"日本-01","alive":true},{"name":"美国-01","alive":true}]}}}`))
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
	proxies := []ProjectMihomoProxy{{Name: "project-mihomo-01"}, {Name: "project-mihomo-02"}}

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies, nil, 0)
	require.NoError(t, err)
	require.Equal(t, 2, assigned)
	require.Equal(t, "日本-01", selected["project-mihomo-01"])
	require.Equal(t, "美国-01", selected["project-mihomo-02"])
}

func TestProjectMihomoAssignProviderNodesSkipsOldPortWhenSelectionUnavailable(t *testing.T) {
	selected := map[string]string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"日本-01","alive":true},{"name":"美国-01","alive":true}]}}}`))
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

	assigned, err := svc.assignProviderNodes(context.Background(), settings, proxies, []string{"香港-01", "美国-01"}, 2)
	require.NoError(t, err)
	require.Equal(t, 2, assigned)
	require.NotContains(t, selected, "project-mihomo-01")
	require.Equal(t, "美国-01", selected["project-mihomo-02"])
	require.Equal(t, "日本-01", selected["project-mihomo-03"])
}

func TestProjectMihomoProviderNodesKeepDuplicateNamesAcrossProviders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription-01":{"proxies":[{"name":"日本-01","alive":true,"history":[{"time":"2026-05-21T10:00:00Z","delay":31}]}]},"project-subscription-02":{"proxies":[{"name":"日本-01","alive":true,"history":[{"time":"2026-05-21T10:00:00Z","delay":87}]}]}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	nodes, err := svc.providerNodes(context.Background(), server.Client(), &ProjectMihomoSettings{
		ControllerURL:    server.URL,
		SubscriptionURLs: []string{"https://a.example/sub", "https://b.example/sub"},
	})
	require.NoError(t, err)
	require.Len(t, nodes, 2)
	require.Equal(t, "project-subscription-01", nodes[0].Provider)
	require.Equal(t, "project-subscription-02", nodes[1].Provider)
	require.NotEqual(t, nodes[0].Key, nodes[1].Key)
	require.NotNil(t, nodes[0].LatencyMS)
	require.Equal(t, 31, *nodes[0].LatencyMS)
	require.NotNil(t, nodes[1].LatencyMS)
	require.Equal(t, 87, *nodes[1].LatencyMS)
}

func TestProjectMihomoProviderNodesSkipSubscriptionInfoEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"剩余流量：107.81+GB","alive":true},{"name":"套餐到期：2026-06-16","alive":true},{"name":"香港01[X中转1]x2.0","alive":true}]}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	nodes, err := svc.providerNodes(context.Background(), server.Client(), &ProjectMihomoSettings{
		ControllerURL:   server.URL,
		SubscriptionURL: "https://example.com/sub",
	})
	require.NoError(t, err)
	require.Len(t, nodes, 1)
	require.Equal(t, "香港01[X中转1]x2.0", nodes[0].Name)
}

func TestProjectMihomoProviderNodesExcludeKeywords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"香港-01","alive":true},{"name":"台灣-01","alive":true},{"name":"日本-01","alive":true},{"name":"美国-01","alive":true}]}}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	settings := &ProjectMihomoSettings{
		ControllerURL:       server.URL,
		SubscriptionURL:     "https://example.com/sub",
		NodeExcludeEnabled:  true,
		NodeExcludeKeywords: []string{"香港", "台灣"},
	}
	svc.normalize(settings)

	nodes, err := svc.providerNodes(context.Background(), server.Client(), settings)
	require.NoError(t, err)
	require.Equal(t, []string{"日本-01", "美国-01"}, projectMihomoNodeNames(nodes))
}

func TestProjectMihomoTestNodesReturnsLatencyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/sub":
			_, _ = w.Write([]byte("proxies:\n  - name: 日本-01\n    server: jp.example.com\n  - name: 美国-01\n    server: us.example.com\n"))
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/configs"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"日本-01","alive":true,"history":[{"time":"2026-05-21T10:00:00Z","delay":41}]},{"name":"美国-01","alive":true}]}}}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription/") && strings.HasSuffix(r.URL.Path, "/healthcheck"):
			name, err := url.PathUnescape(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/providers/proxies/project-subscription/"), "/healthcheck"))
			require.NoError(t, err)
			require.Equal(t, projectMihomoDelayURL, r.URL.Query().Get("url"))
			require.Equal(t, "3000", r.URL.Query().Get("timeout"))
			delayByName := map[string]int{
				"美国-01": 81,
			}
			delay, ok := delayByName[name]
			require.True(t, ok)
			_ = json.NewEncoder(w).Encode(map[string]int{"delay": delay})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Setenv("DATA_DIR", t.TempDir())

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	result, err := svc.TestNodes(context.Background(), &ProjectMihomoSettings{
		SubscriptionURL: server.URL + "/sub",
		SubscriptionUA:  "sub2api/mihomo",
		UpdateInterval:  3600,
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   2,
		ControllerURL:   server.URL,
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: []string{"", ""},
	})
	require.NoError(t, err)
	require.Len(t, result.Nodes, 2)
	require.Equal(t, []string{"日本", "美国"}, result.AvailableRegions)
	require.NotNil(t, result.Nodes[0].LatencyMS)
	require.Equal(t, 41, *result.Nodes[0].LatencyMS)
	require.Equal(t, "success", result.Nodes[0].LatencyStatus)
	require.NotNil(t, result.Nodes[1].LatencyMS)
	require.Equal(t, 81, *result.Nodes[1].LatencyMS)
	require.Equal(t, "success", result.Nodes[1].LatencyStatus)
}

func TestProjectMihomoTestNodeReturnsSingleLatencyResult(t *testing.T) {
	var tested []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/sub":
			_, _ = w.Write([]byte("proxies:\n  - name: 日本-01\n    server: jp.example.com\n  - name: 美国-01\n    server: us.example.com\n"))
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/configs"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/providers/proxies":
			_, _ = w.Write([]byte(`{"providers":{"project-subscription":{"proxies":[{"name":"日本-01","alive":true,"history":[{"time":"2026-05-21T10:00:00Z","delay":41}]},{"name":"美国-01","alive":true,"history":[{"time":"2026-05-21T10:00:00Z","delay":91}]}]}}}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/providers/proxies/project-subscription/") && strings.HasSuffix(r.URL.Path, "/healthcheck"):
			name, err := url.PathUnescape(strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/providers/proxies/project-subscription/"), "/healthcheck"))
			require.NoError(t, err)
			tested = append(tested, name)
			_ = json.NewEncoder(w).Encode(map[string]int{"delay": 63})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Setenv("DATA_DIR", t.TempDir())

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, &projectMihomoAdminServiceStub{})
	result, err := svc.TestNode(context.Background(), &ProjectMihomoSettings{
		SubscriptionURL: server.URL + "/sub",
		SubscriptionUA:  "sub2api/mihomo",
		UpdateInterval:  3600,
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   2,
		ControllerURL:   server.URL,
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: []string{"", ""},
	}, ProjectMihomoNode{
		Key: projectMihomoNodeKey("project-subscription", "美国-01"),
	})
	require.NoError(t, err)
	require.Equal(t, []string{"美国-01"}, tested)
	require.Equal(t, "美国-01", result.Name)
	require.NotNil(t, result.LatencyMS)
	require.Equal(t, 63, *result.LatencyMS)
	require.Equal(t, "success", result.LatencyStatus)
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
		StartPort:       61000,
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
	require.Len(t, status.AvailableNodes, 3)
	require.Equal(t, "日本01[X中转1]x2.0", status.AvailableNodes[0].Name)
	require.Equal(t, []string{"日本", "美国", "香港", ""}, status.Settings.ListenerRegions)
}

func TestProjectMihomoGetStatusResolvesLegacyNodeKeyToCurrentProvider(t *testing.T) {
	settings := ProjectMihomoSettings{
		SubscriptionURL: "https://example.com/sub",
		SubscriptionUA:  "sub2api/mihomo",
		UpdateInterval:  3600,
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   1,
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: []string{projectMihomoNodeKey("project-subscription-02", "日本01[X中转1]x2.0")},
	}
	raw, err := json.Marshal(settings)
	require.NoError(t, err)

	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{
		SettingKeyProjectMihomoSettings: string(raw),
	}}, &projectMihomoAdminServiceStub{})

	t.Setenv("DATA_DIR", t.TempDir())
	require.NoError(t, os.MkdirAll(filepath.Join(svc.configDir(), "providers"), 0o755))
	require.NoError(t, os.WriteFile(
		svc.providerCachePath(),
		[]byte("proxies:\n  - name: 日本01[X中转1]x2.0\n    server: jp.example.com\n"),
		0o644,
	))

	status, err := svc.GetStatus(context.Background())
	require.NoError(t, err)
	require.Len(t, status.AvailableNodes, 1)
	require.Equal(t, status.AvailableNodes[0].Key, status.Settings.ListenerRegions[0])
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
			{ID: 1, Name: "project-mihomo-01", Protocol: "socks5h", Host: "127.0.0.1", Port: 61000, Status: StatusActive},
			{ID: 2, Name: "project-mihomo-04", Protocol: "http", Host: "127.0.0.1", Port: 61003, Status: StatusActive},
			{ID: 3, Name: "project-mihomo-04", Protocol: "socks5h", Host: "127.0.0.1", Port: 61003, Status: StatusActive},
			{ID: 4, Name: "project-mihomo-05", Protocol: "socks5h", Host: "127.0.0.1", Port: 61004, Status: StatusActive},
		},
	}
	svc := NewProjectMihomoService(&settingRepoStub{values: map[string]string{}}, stub)
	settings := &ProjectMihomoSettings{
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
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

func TestProjectMihomoSetSettingsBlocksRemovingListenerInUse(t *testing.T) {
	previousSettings := ProjectMihomoSettings{
		SubscriptionUA:  "sub2api/mihomo",
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   2,
		ListenerPorts:   []int{61000, 61001},
		ListenerNames:   []string{"project-mihomo-01", "project-mihomo-02"},
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: make([]string, 2),
	}
	raw, err := json.Marshal(previousSettings)
	require.NoError(t, err)
	stub := &projectMihomoAdminServiceStub{
		proxies: []Proxy{
			{ID: 1, Name: "project-mihomo-01", Protocol: "socks5h", Host: "127.0.0.1", Port: 61000, Status: StatusActive},
			{ID: 2, Name: "project-mihomo-02", Protocol: "socks5h", Host: "127.0.0.1", Port: 61001, Status: StatusActive},
		},
		proxyAccounts: map[int64][]ProxyAccountSummary{
			2: {{ID: 11, Name: "account-a", Platform: PlatformOpenAI, Type: AccountTypeOAuth}},
		},
	}
	svc := NewProjectMihomoService(&projectMihomoSettingRepoStub{values: map[string]string{
		SettingKeyProjectMihomoSettings: string(raw),
	}}, stub)

	_, err = svc.SetSettings(context.Background(), &ProjectMihomoSettings{
		SubscriptionUA:  "sub2api/mihomo",
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   1,
		ListenerPorts:   []int{61000},
		ListenerNames:   []string{"project-mihomo-01"},
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: []string{""},
	})

	require.ErrorIs(t, err, ErrProjectMihomoProxyInUse)
	appErr := infraerrors.FromError(err)
	require.Equal(t, "1", appErr.Metadata["account_count"])
	require.Contains(t, appErr.Metadata["proxies"], "project-mihomo-02")
	require.Empty(t, stub.bulkUpdateCalls)
	require.Empty(t, stub.deleteCalls)
}

func TestProjectMihomoSetSettingsForceRemovingListenerClearsAccounts(t *testing.T) {
	previousSettings := ProjectMihomoSettings{
		SubscriptionUA:  "sub2api/mihomo",
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   2,
		ListenerPorts:   []int{61000, 61001},
		ListenerNames:   []string{"project-mihomo-01", "project-mihomo-02"},
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: make([]string, 2),
	}
	raw, err := json.Marshal(previousSettings)
	require.NoError(t, err)
	stub := &projectMihomoAdminServiceStub{
		proxies: []Proxy{
			{ID: 1, Name: "project-mihomo-01", Protocol: "socks5h", Host: "127.0.0.1", Port: 61000, Status: StatusActive},
			{ID: 2, Name: "project-mihomo-02", Protocol: "socks5h", Host: "127.0.0.1", Port: 61001, Status: StatusActive},
		},
		proxyAccounts: map[int64][]ProxyAccountSummary{
			2: {{ID: 11, Name: "account-a", Platform: PlatformOpenAI, Type: AccountTypeOAuth}},
		},
	}
	svc := NewProjectMihomoService(&projectMihomoSettingRepoStub{values: map[string]string{
		SettingKeyProjectMihomoSettings: string(raw),
	}}, stub)

	settings, err := svc.SetSettingsWithOptions(context.Background(), &ProjectMihomoSettings{
		SubscriptionUA:  "sub2api/mihomo",
		Protocol:        "socks5h",
		TargetHost:      "127.0.0.1",
		StartPort:       61000,
		ListenerCount:   1,
		ListenerPorts:   []int{61000},
		ListenerNames:   []string{"project-mihomo-01"},
		ControllerURL:   "http://127.0.0.1:9097",
		ProxyNamePrefix: "project-mihomo",
		ListenerRegions: []string{""},
	}, ProjectMihomoSaveOptions{ForceRemoveInUse: true})

	require.NoError(t, err)
	require.Equal(t, []string{"project-mihomo-01"}, settings.ListenerNames)
	require.Len(t, stub.bulkUpdateCalls, 1)
	require.Equal(t, []int64{11}, stub.bulkUpdateCalls[0].AccountIDs)
	require.NotNil(t, stub.bulkUpdateCalls[0].ProxyID)
	require.Zero(t, *stub.bulkUpdateCalls[0].ProxyID)
	require.Contains(t, stub.deleteCalls, int64(2))
	require.Len(t, stub.proxies, 1)
	require.Equal(t, "project-mihomo-01", stub.proxies[0].Name)
}

func TestParseProjectMihomoSubscriptionNodeNames_Base64URIsSkipsPlaceholderHost(t *testing.T) {
	content := "dmxlc3M6Ly91c2VyQERvbnQudXNlLnRoaXMubm9kZTo4ODg4IyVFNSU4OSVBOSVFNCVCRCU5OSVFNiVCNSU4MSVFOSU4NyU4Rg0Kc3M6Ly9ZMmhoWTJoaE1qQXRAcmVhbC5leGFtcGxlLmNvbToxMTAxMiMlRTklQTYlOTklRTYlQjglQUY="
	names := parseProjectMihomoSubscriptionNodeNames([]byte(content))
	require.Equal(t, []string{"香港"}, names)
}
