package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/httpclient"
	"gopkg.in/yaml.v3"
)

const (
	projectMihomoConfigFilename  = "config.yaml"
	projectMihomoProviderName    = "project-subscription"
	projectMihomoProviderPath    = "./providers/project-subscription.yaml"
	projectMihomoHTTPTimeout     = 10 * time.Second
	projectMihomoReloadPath      = "/root/.config/mihomo/config.yaml"
	projectMihomoAssignWait      = 8 * time.Second
	projectMihomoAssignPoll      = 500 * time.Millisecond
	projectMihomoDelayURL        = "https://www.gstatic.com/generate_204"
	projectMihomoDelayTimeoutMS  = 3000
	projectMihomoDelayWorkers    = 32
	projectMihomoProviderMaxSize = 16 << 20
	projectMihomoPlaceholderHost = "dont.use.this.node"
)

var (
	ErrProjectMihomoSubscriptionRequired = infraerrors.BadRequest("PROJECT_MIHOMO_SUBSCRIPTION_REQUIRED", "subscription_url is required")
	ErrProjectMihomoControllerRequired   = infraerrors.BadRequest("PROJECT_MIHOMO_CONTROLLER_REQUIRED", "controller_url is required")
	ErrProjectMihomoProtocolInvalid      = infraerrors.BadRequest("PROJECT_MIHOMO_PROTOCOL_INVALID", "protocol must be one of http, https, socks5, socks5h")
	ErrProjectMihomoListenerCountInvalid = infraerrors.BadRequest("PROJECT_MIHOMO_LISTENER_COUNT_INVALID", "listener_count must be between 1 and 32")
	ErrProjectMihomoStartPortInvalid     = infraerrors.BadRequest("PROJECT_MIHOMO_START_PORT_INVALID", "start_port must be between 1 and 65535")
	ErrProjectMihomoPortRangeInvalid     = infraerrors.BadRequest("PROJECT_MIHOMO_PORT_RANGE_INVALID", "listener ports exceed valid range")
	ErrProjectMihomoSubscriptionFetch    = infraerrors.BadRequest("PROJECT_MIHOMO_SUBSCRIPTION_FETCH_FAILED", "failed to fetch project mihomo subscription")
)

var projectMihomoRegionAliases = map[string][]string{
	"hongkong":  {"香港", "hongkong", "hong kong"},
	"japan":     {"日本", "japan", "tokyo", "osaka"},
	"usa":       {"美国", "美國", "usa", "unitedstates", "losangeles", "sanjose", "sanfrancisco", "seattle", "newyork"},
	"singapore": {"新加坡", "singapore"},
	"taiwan":    {"台湾", "台灣", "taiwan"},
	"korea":     {"韩国", "韓國", "korea", "southkorea", "seoul"},
	"uk":        {"英国", "英國", "uk", "unitedkingdom", "britain", "london"},
	"germany":   {"德国", "德國", "germany", "frankfurt"},
}

type ProjectMihomoSettings struct {
	SubscriptionURL   string   `json:"subscription_url"`
	SubscriptionURLs  []string `json:"subscription_urls"`
	SubscriptionNames []string `json:"subscription_names"`
	SubscriptionUA    string   `json:"subscription_user_agent"`
	UpdateInterval    int      `json:"update_interval"`
	Protocol          string   `json:"protocol"`
	TargetHost        string   `json:"target_host"`
	StartPort         int      `json:"start_port"`
	ListenerCount     int      `json:"listener_count"`
	ControllerURL     string   `json:"controller_url"`
	ControllerSecret  string   `json:"controller_secret"`
	ProxyNamePrefix   string   `json:"proxy_name_prefix"`
	ListenerRegions   []string `json:"listener_regions"`
}

type ProjectMihomoProxy struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

type ProjectMihomoNode struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	Region         string `json:"region"`
	Alive          bool   `json:"alive"`
	Provider       string `json:"provider,omitempty"`
	ProviderLabel  string `json:"provider_label,omitempty"`
	LatencyMS      *int   `json:"latency_ms,omitempty"`
	LatencyStatus  string `json:"latency_status,omitempty"`
	LatencyMessage string `json:"latency_message,omitempty"`
}

type ProjectMihomoStatus struct {
	Settings         ProjectMihomoSettings `json:"settings"`
	ConfigPath       string                `json:"config_path"`
	Proxies          []ProjectMihomoProxy  `json:"proxies"`
	AvailableNodes   []ProjectMihomoNode   `json:"available_nodes"`
	AvailableRegions []string              `json:"available_regions"`
}

type ProjectMihomoNodeTestResult struct {
	Nodes            []ProjectMihomoNode `json:"nodes"`
	AvailableRegions []string            `json:"available_regions"`
}

type ProjectMihomoSyncResult struct {
	ConfigPath string               `json:"config_path"`
	Proxies    []ProjectMihomoProxy `json:"proxies"`
	Created    int                  `json:"created"`
	Reused     int                  `json:"reused"`
	Assigned   int                  `json:"assigned"`
	Reloaded   bool                 `json:"reloaded"`
}

type ProjectMihomoService struct {
	settingRepo  SettingRepository
	adminService AdminService
}

type projectMihomoProviderRef struct {
	Name        string
	Path        string
	URL         string
	DisplayName string
}

func NewProjectMihomoService(settingRepo SettingRepository, adminService AdminService) *ProjectMihomoService {
	return &ProjectMihomoService{
		settingRepo:  settingRepo,
		adminService: adminService,
	}
}

func DefaultProjectMihomoSettings() ProjectMihomoSettings {
	targetHost := "mihomo-sub2api"
	controllerURL := "http://mihomo-sub2api:9097"
	if _, err := os.Stat("/app/data"); err != nil {
		targetHost = "127.0.0.1"
		controllerURL = "http://127.0.0.1:9097"
	}
	return ProjectMihomoSettings{
		SubscriptionURL:   "",
		SubscriptionURLs:  nil,
		SubscriptionNames: nil,
		SubscriptionUA:    "sub2api/mihomo",
		UpdateInterval:    3600,
		Protocol:          "socks5h",
		TargetHost:        targetHost,
		StartPort:         61000,
		ListenerCount:     4,
		ControllerURL:     controllerURL,
		ControllerSecret:  "",
		ProxyNamePrefix:   "project-mihomo",
		ListenerRegions:   make([]string, 4),
	}
}

func (s *ProjectMihomoService) GetSettings(ctx context.Context) (*ProjectMihomoSettings, error) {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyProjectMihomoSettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			defaults := DefaultProjectMihomoSettings()
			return &defaults, nil
		}
		return nil, fmt.Errorf("get project mihomo settings: %w", err)
	}

	settings := DefaultProjectMihomoSettings()
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &settings); err != nil {
			defaults := DefaultProjectMihomoSettings()
			return &defaults, nil
		}
	}
	s.normalize(&settings)
	return &settings, nil
}

func (s *ProjectMihomoService) SetSettings(ctx context.Context, settings *ProjectMihomoSettings) (*ProjectMihomoSettings, error) {
	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	previous, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	normalized := *settings
	s.normalize(&normalized)
	if err := s.validate(&normalized, false); err != nil {
		return nil, err
	}
	if err := s.ensureProviderFiles(ctx, previous, &normalized); err != nil {
		return nil, err
	}

	data, err := json.Marshal(&normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal project mihomo settings: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyProjectMihomoSettings, string(data)); err != nil {
		return nil, fmt.Errorf("save project mihomo settings: %w", err)
	}
	if err := s.cleanupProviderFiles(previous, &normalized); err != nil {
		return nil, err
	}
	return &normalized, nil
}

func (s *ProjectMihomoService) GetStatus(ctx context.Context) (*ProjectMihomoStatus, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	availableNodes := s.availableNodes(settings)
	availableRegions := extractProjectMihomoRegions(projectMihomoNodeNames(availableNodes))
	statusSettings := *settings
	statusSettings.ListenerRegions = resolveProjectMihomoListenerSelections(statusSettings.ListenerRegions, availableNodes, availableRegions)

	return &ProjectMihomoStatus{
		Settings:         statusSettings,
		ConfigPath:       s.configPath(),
		Proxies:          s.buildProxies(settings),
		AvailableNodes:   availableNodes,
		AvailableRegions: availableRegions,
	}, nil
}

func (s *ProjectMihomoService) Sync(ctx context.Context, settings *ProjectMihomoSettings) (*ProjectMihomoSyncResult, error) {
	saved, err := s.SetSettings(ctx, settings)
	if err != nil {
		return nil, err
	}

	configPath, err := s.writeConfig(saved)
	if err != nil {
		return nil, err
	}

	proxies := s.buildProxies(saved)
	created, reused, err := s.syncProxyRows(ctx, saved, proxies)
	if err != nil {
		return nil, err
	}

	reloaded := false
	assigned := 0
	if err := s.reloadConfig(ctx, saved, projectMihomoReloadPath); err == nil {
		reloaded = true
		if count, err := s.assignProviderNodes(ctx, saved, proxies); err == nil {
			assigned = count
		}
	}

	return &ProjectMihomoSyncResult{
		ConfigPath: configPath,
		Proxies:    proxies,
		Created:    created,
		Reused:     reused,
		Assigned:   assigned,
		Reloaded:   reloaded,
	}, nil
}

func (s *ProjectMihomoService) TestNodes(ctx context.Context, settings *ProjectMihomoSettings) (*ProjectMihomoNodeTestResult, error) {
	if settings == nil {
		saved, err := s.GetSettings(ctx)
		if err != nil {
			return nil, err
		}
		settings = saved
	}

	normalized := *settings
	s.normalize(&normalized)
	if err := s.validate(&normalized, true); err != nil {
		return nil, err
	}
	previous, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.ensureProviderFiles(ctx, previous, &normalized); err != nil {
		return nil, err
	}
	if _, err := s.writeConfig(&normalized); err != nil {
		return nil, err
	}

	client, err := httpclient.GetClient(httpclient.Options{
		Timeout: projectMihomoHTTPTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("build controller client: %w", err)
	}
	if err := s.reloadConfig(ctx, &normalized, projectMihomoReloadPath); err != nil {
		return nil, err
	}
	_ = s.refreshProvider(ctx, client, &normalized)

	nodes, err := s.waitProviderNodes(ctx, client, &normalized)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		nodes = s.availableNodes(&normalized)
	}
	nodes = s.testNodeDelays(ctx, client, &normalized, nodes)

	return &ProjectMihomoNodeTestResult{
		Nodes:            nodes,
		AvailableRegions: extractProjectMihomoRegions(projectMihomoNodeNames(nodes)),
	}, nil
}

func (s *ProjectMihomoService) TestNode(ctx context.Context, settings *ProjectMihomoSettings, node ProjectMihomoNode) (*ProjectMihomoNode, error) {
	if settings == nil {
		saved, err := s.GetSettings(ctx)
		if err != nil {
			return nil, err
		}
		settings = saved
	}

	normalized := *settings
	s.normalize(&normalized)
	if err := s.validate(&normalized, true); err != nil {
		return nil, err
	}
	previous, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.ensureProviderFiles(ctx, previous, &normalized); err != nil {
		return nil, err
	}
	if _, err := s.writeConfig(&normalized); err != nil {
		return nil, err
	}

	client, err := httpclient.GetClient(httpclient.Options{
		Timeout: projectMihomoHTTPTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("build controller client: %w", err)
	}
	if err := s.reloadConfig(ctx, &normalized, projectMihomoReloadPath); err != nil {
		return nil, err
	}
	_ = s.refreshProvider(ctx, client, &normalized)

	target := node
	if target.Name == "" {
		_, name, ok := parseProjectMihomoNodeKey(target.Key)
		if !ok {
			return nil, infraerrors.BadRequest("PROJECT_MIHOMO_NODE_REQUIRED", "node is required")
		}
		target.Name = name
	}
	nodes, err := s.waitProviderNodes(ctx, client, &normalized)
	if err != nil {
		return nil, err
	}
	if matched := findProjectMihomoNode(nodes, firstNonEmptyString(target.Key, projectMihomoNodeKey(target.Provider, target.Name), target.Name)); matched != nil {
		target = *matched
	}

	latency, err := s.testNodeDelay(ctx, client, &normalized, target)
	if err != nil {
		target.LatencyMS = nil
		target.LatencyStatus = "failed"
		target.LatencyMessage = err.Error()
		return &target, nil
	}
	target.LatencyMS = &latency
	target.LatencyStatus = "success"
	target.LatencyMessage = ""
	target.Alive = true
	return &target, nil
}

func (s *ProjectMihomoService) normalize(settings *ProjectMihomoSettings) {
	defaults := DefaultProjectMihomoSettings()
	settings.SubscriptionURL = strings.TrimSpace(settings.SubscriptionURL)
	settings.SubscriptionURLs = normalizeProjectMihomoSubscriptionURLs(settings.SubscriptionURLs, settings.SubscriptionURL)
	settings.SubscriptionNames = normalizeProjectMihomoSubscriptionNames(settings.SubscriptionNames, len(settings.SubscriptionURLs))
	if len(settings.SubscriptionURLs) > 0 {
		settings.SubscriptionURL = settings.SubscriptionURLs[0]
	} else {
		settings.SubscriptionURL = ""
	}
	settings.SubscriptionUA = strings.TrimSpace(settings.SubscriptionUA)
	if settings.SubscriptionUA == "" {
		settings.SubscriptionUA = defaults.SubscriptionUA
	}
	if settings.UpdateInterval <= 0 {
		settings.UpdateInterval = defaults.UpdateInterval
	}
	settings.Protocol = strings.ToLower(strings.TrimSpace(settings.Protocol))
	if settings.Protocol == "" {
		settings.Protocol = defaults.Protocol
	}
	settings.TargetHost = strings.TrimSpace(settings.TargetHost)
	if settings.TargetHost == "" {
		settings.TargetHost = defaults.TargetHost
	}
	if settings.StartPort == 0 {
		settings.StartPort = defaults.StartPort
	}
	if settings.ListenerCount == 0 {
		settings.ListenerCount = defaults.ListenerCount
	}
	settings.ControllerURL = strings.TrimSpace(settings.ControllerURL)
	if settings.ControllerURL == "" {
		settings.ControllerURL = defaults.ControllerURL
	}
	if !strings.Contains(settings.ControllerURL, "://") {
		settings.ControllerURL = "http://" + settings.ControllerURL
	}
	settings.ControllerSecret = strings.TrimSpace(settings.ControllerSecret)
	settings.ProxyNamePrefix = strings.TrimSpace(settings.ProxyNamePrefix)
	if settings.ProxyNamePrefix == "" {
		settings.ProxyNamePrefix = defaults.ProxyNamePrefix
	}
	settings.ListenerRegions = normalizeProjectMihomoListenerRegions(settings.ListenerRegions, settings.ListenerCount)
}

func (s *ProjectMihomoService) validate(settings *ProjectMihomoSettings, requireSubscription bool) error {
	if requireSubscription && len(settings.SubscriptionURLs) == 0 {
		return ErrProjectMihomoSubscriptionRequired
	}
	if settings.ControllerURL == "" {
		return ErrProjectMihomoControllerRequired
	}
	switch settings.Protocol {
	case "http", "https", "socks5", "socks5h":
	default:
		return ErrProjectMihomoProtocolInvalid
	}
	if settings.ListenerCount < 1 || settings.ListenerCount > 32 {
		return ErrProjectMihomoListenerCountInvalid
	}
	if settings.StartPort < 1 || settings.StartPort > 65535 {
		return ErrProjectMihomoStartPortInvalid
	}
	if settings.StartPort+settings.ListenerCount-1 > 65535 {
		return ErrProjectMihomoPortRangeInvalid
	}
	return nil
}

func (s *ProjectMihomoService) buildProxies(settings *ProjectMihomoSettings) []ProjectMihomoProxy {
	out := make([]ProjectMihomoProxy, 0, settings.ListenerCount)
	for i := 0; i < settings.ListenerCount; i++ {
		out = append(out, ProjectMihomoProxy{
			Name:     fmt.Sprintf("%s-%02d", settings.ProxyNamePrefix, i+1),
			Protocol: settings.Protocol,
			Host:     settings.TargetHost,
			Port:     settings.StartPort + i,
		})
	}
	return out
}

func normalizeProjectMihomoListenerRegions(regions []string, count int) []string {
	if count <= 0 {
		return nil
	}
	out := make([]string, count)
	for i := 0; i < count; i++ {
		if i < len(regions) {
			out[i] = strings.TrimSpace(regions[i])
		}
	}
	return out
}

func normalizeProjectMihomoSubscriptionURLs(urls []string, fallback string) []string {
	seen := make(map[string]struct{}, len(urls)+1)
	out := make([]string, 0, len(urls)+1)
	appendValue := func(value string) {
		for _, part := range strings.FieldsFunc(value, func(r rune) bool {
			return r == '\n' || r == '\r'
		}) {
			item := strings.TrimSpace(part)
			if item == "" {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			out = append(out, item)
		}
	}
	for i := range urls {
		appendValue(urls[i])
	}
	appendValue(fallback)
	return out
}

func normalizeProjectMihomoSubscriptionNames(names []string, count int) []string {
	if count <= 0 {
		return nil
	}

	out := make([]string, count)
	for i := 0; i < count; i++ {
		if i < len(names) {
			out[i] = strings.TrimSpace(names[i])
		}
	}
	return out
}

func resolveProjectMihomoListenerRegions(regions []string, availableRegions []string) []string {
	if len(regions) == 0 {
		return nil
	}
	out := make([]string, len(regions))
	for i := range regions {
		out[i] = resolveProjectMihomoListenerRegion(regions[i], availableRegions)
	}
	return out
}

func resolveProjectMihomoListenerRegion(region string, availableRegions []string) string {
	region = strings.TrimSpace(region)
	if region == "" || len(availableRegions) == 0 {
		return region
	}
	matches := filterProjectMihomoNodesByRegion(availableRegions, region)
	if len(matches) == 0 {
		return region
	}
	return matches[0]
}

func resolveProjectMihomoListenerSelections(values []string, availableNodes []ProjectMihomoNode, availableRegions []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	for i := range values {
		out[i] = resolveProjectMihomoListenerSelection(values[i], availableNodes, availableRegions)
	}
	return out
}

func resolveProjectMihomoListenerSelection(value string, availableNodes []ProjectMihomoNode, availableRegions []string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if matched := findProjectMihomoNode(availableNodes, value); matched != nil {
		return matched.Key
	}
	return resolveProjectMihomoListenerRegion(value, availableRegions)
}

func (s *ProjectMihomoService) syncProxyRows(ctx context.Context, settings *ProjectMihomoSettings, proxies []ProjectMihomoProxy) (int, int, error) {
	existing, err := s.adminService.GetAllProxies(ctx)
	if err != nil {
		return 0, 0, err
	}

	expectedPorts := make(map[int]struct{}, len(proxies))
	managed := make([]*Proxy, 0, len(existing))
	for i := range proxies {
		expectedPorts[proxies[i].Port] = struct{}{}
	}
	for i := range existing {
		if isProjectMihomoManagedProxy(settings, &existing[i], expectedPorts) {
			managed = append(managed, &existing[i])
		}
	}

	created := 0
	reused := 0
	staleIDs := make([]int64, 0)
	matchedIDs := make(map[int64]struct{}, len(managed))
	for i := range proxies {
		expected := proxies[i]
		matches := make([]*Proxy, 0, 2)
		for j := range managed {
			if projectMihomoProxyMatchesExpected(managed[j], expected) {
				matches = append(matches, managed[j])
			}
		}

		if len(matches) == 0 {
			if _, err := s.adminService.CreateProxy(ctx, &CreateProxyInput{
				Name:     expected.Name,
				Protocol: expected.Protocol,
				Host:     expected.Host,
				Port:     expected.Port,
			}); err != nil {
				return created, reused, err
			}
			created++
			continue
		}

		keep := selectProjectMihomoProxyMatch(matches, expected)
		for j := range matches {
			item := matches[j]
			matchedIDs[item.ID] = struct{}{}
			if projectMihomoProxyNeedsUpdate(item, expected) {
				if _, err := s.adminService.UpdateProxy(ctx, item.ID, &UpdateProxyInput{
					Name:     expected.Name,
					Protocol: expected.Protocol,
					Host:     expected.Host,
					Port:     expected.Port,
					Status:   StatusActive,
				}); err != nil {
					return created, reused, err
				}
			}
			if item.ID != keep.ID {
				staleIDs = append(staleIDs, item.ID)
			}
		}
		reused++
	}

	for i := range managed {
		if _, ok := matchedIDs[managed[i].ID]; ok {
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(managed[i].Name), settings.ProxyNamePrefix+"-") {
			staleIDs = append(staleIDs, managed[i].ID)
		}
	}

	for _, id := range dedupeProjectMihomoProxyIDs(staleIDs) {
		_ = s.adminService.DeleteProxy(ctx, id)
	}
	return created, reused, nil
}

func isProjectMihomoManagedProxy(settings *ProjectMihomoSettings, proxy *Proxy, expectedPorts map[int]struct{}) bool {
	if proxy == nil {
		return false
	}
	if strings.HasPrefix(strings.TrimSpace(proxy.Name), settings.ProxyNamePrefix+"-") {
		return true
	}
	if strings.TrimSpace(proxy.Host) != settings.TargetHost {
		return false
	}
	_, ok := expectedPorts[proxy.Port]
	return ok
}

func projectMihomoProxyMatchesExpected(existing *Proxy, expected ProjectMihomoProxy) bool {
	if existing == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(existing.Name), expected.Name) {
		return true
	}
	return strings.TrimSpace(existing.Host) == expected.Host && existing.Port == expected.Port
}

func projectMihomoProxyNeedsUpdate(existing *Proxy, expected ProjectMihomoProxy) bool {
	if existing == nil {
		return false
	}
	if strings.TrimSpace(existing.Name) != expected.Name {
		return true
	}
	if strings.ToLower(strings.TrimSpace(existing.Protocol)) != expected.Protocol {
		return true
	}
	if strings.TrimSpace(existing.Host) != expected.Host {
		return true
	}
	if existing.Port != expected.Port {
		return true
	}
	return existing.Status != StatusActive
}

func selectProjectMihomoProxyMatch(matches []*Proxy, expected ProjectMihomoProxy) *Proxy {
	for i := range matches {
		if !projectMihomoProxyNeedsUpdate(matches[i], expected) {
			return matches[i]
		}
	}
	return matches[0]
}

func dedupeProjectMihomoProxyIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func (s *ProjectMihomoService) configDir() string {
	if dataDir := strings.TrimSpace(os.Getenv("DATA_DIR")); dataDir != "" {
		return filepath.Join(dataDir, "mihomo")
	}
	if _, err := os.Stat("/app/data"); err == nil {
		return "/app/data/mihomo"
	}
	return filepath.Join(".", "data", "mihomo")
}

func (s *ProjectMihomoService) configPath() string {
	return filepath.Join(s.configDir(), projectMihomoConfigFilename)
}

func (s *ProjectMihomoService) providerCachePath() string {
	return s.providerCachePathFor(projectMihomoProviderPath)
}

func (s *ProjectMihomoService) providerCachePathFor(providerPath string) string {
	return filepath.Join(s.configDir(), "providers", filepath.Base(providerPath))
}

func (s *ProjectMihomoService) providerDir() string {
	return filepath.Join(s.configDir(), "providers")
}

func (s *ProjectMihomoService) availableRegions(settings *ProjectMihomoSettings) []string {
	return extractProjectMihomoRegions(projectMihomoNodeNames(s.availableNodes(settings)))
}

func (s *ProjectMihomoService) availableNodes(settings *ProjectMihomoSettings) []ProjectMihomoNode {
	providers := buildProjectMihomoProviderRefs(settings)
	if len(providers) == 0 {
		return nil
	}

	nodes := make([]ProjectMihomoNode, 0)
	for i := range providers {
		providerLabel := projectMihomoProviderDisplayName(providers[i], i, len(providers))
		cachedNames, err := s.providerNodeNamesFromCachePath(s.providerCachePathFor(providers[i].Path))
		if err != nil {
			continue
		}
		for j := range cachedNames {
			name := strings.TrimSpace(cachedNames[j])
			if !isProjectMihomoSelectableNodeName(name) {
				continue
			}
			nodes = append(nodes, ProjectMihomoNode{
				Key:           projectMihomoNodeKey(providers[i].Name, name),
				Name:          name,
				Region:        extractProjectMihomoRegion(name),
				Alive:         true,
				Provider:      providers[i].Name,
				ProviderLabel: providerLabel,
				LatencyStatus: "unknown",
			})
		}
	}
	return nodes
}

func (s *ProjectMihomoService) writeConfig(settings *ProjectMihomoSettings) (string, error) {
	if err := os.MkdirAll(s.configDir(), 0o755); err != nil {
		return "", fmt.Errorf("create mihomo dir: %w", err)
	}
	if err := os.MkdirAll(s.providerDir(), 0o755); err != nil {
		return "", fmt.Errorf("create mihomo provider dir: %w", err)
	}

	content, err := s.renderConfig(settings)
	if err != nil {
		return "", err
	}

	configPath := s.configPath()
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		return "", fmt.Errorf("write mihomo config: %w", err)
	}
	return configPath, nil
}

func (s *ProjectMihomoService) renderConfig(settings *ProjectMihomoSettings) ([]byte, error) {
	proxies := s.buildProxies(settings)
	providers := buildProjectMihomoProviderRefs(settings)
	listeners := make([]map[string]any, 0, len(proxies))
	groups := make([]map[string]any, 0, len(proxies))
	providerNames := make([]string, 0, len(providers))
	providerConfigs := make(map[string]any, len(providers))
	for i := range providers {
		item := providers[i]
		providerNames = append(providerNames, item.Name)
		providerConfigs[item.Name] = map[string]any{
			"type": "file",
			"path": item.Path,
			"health-check": map[string]any{
				"enable":   true,
				"url":      "https://www.gstatic.com/generate_204",
				"interval": 300,
				"timeout":  5000,
				"lazy":     true,
			},
		}
	}
	for i := range proxies {
		item := proxies[i]
		listeners = append(listeners, map[string]any{
			"name":   item.Name,
			"type":   "mixed",
			"port":   item.Port,
			"listen": "0.0.0.0",
			"udp":    true,
			"proxy":  item.Name,
		})
		groups = append(groups, map[string]any{
			"name": item.Name,
			"type": "select",
			"use":  providerNames,
		})
	}

	root := map[string]any{
		"mode":                "rule",
		"allow-lan":           true,
		"bind-address":        "*",
		"external-controller": controllerListenAddress(settings.ControllerURL),
		"secret":              settings.ControllerSecret,
		"log-level":           "info",
		"proxy-providers":     providerConfigs,
		"proxy-groups":        groups,
		"listeners":           listeners,
	}

	data, err := yaml.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("marshal mihomo config: %w", err)
	}
	return data, nil
}

func controllerListenAddress(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "0.0.0.0:9097"
	}

	parseTarget := value
	if !strings.Contains(parseTarget, "://") {
		parseTarget = "http://" + parseTarget
	}
	if parsed, err := url.Parse(parseTarget); err == nil {
		if port := parsed.Port(); port != "" {
			return net.JoinHostPort("0.0.0.0", port)
		}
	}

	address := controllerAddress(value)
	if _, port, err := net.SplitHostPort(address); err == nil && port != "" {
		return net.JoinHostPort("0.0.0.0", port)
	}
	return address
}

func controllerAddress(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "https://")
	return strings.TrimSuffix(value, "/")
}

func (s *ProjectMihomoService) reloadConfig(ctx context.Context, settings *ProjectMihomoSettings, configPath string) error {
	client, err := httpclient.GetClient(httpclient.Options{
		Timeout: projectMihomoHTTPTimeout,
	})
	if err != nil {
		return fmt.Errorf("build controller client: %w", err)
	}

	body, err := json.Marshal(map[string]any{
		"path": configPath,
	})
	if err != nil {
		return fmt.Errorf("marshal reload payload: %w", err)
	}

	target := strings.TrimRight(settings.ControllerURL, "/") + "/configs?force=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build reload request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if settings.ControllerSecret != "" {
		req.Header.Set("Authorization", "Bearer "+settings.ControllerSecret)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("reload mihomo config: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("reload mihomo config: unexpected status %d", resp.StatusCode)
	}
	return nil
}

type mihomoProvidersResponse struct {
	Providers map[string]mihomoProvider `json:"providers"`
}

type mihomoProvider struct {
	Proxies []mihomoProviderProxy `json:"proxies"`
}

type mihomoProviderProxy struct {
	Name    string `json:"name"`
	Alive   bool   `json:"alive"`
	History []struct {
		Time  string `json:"time"`
		Delay int    `json:"delay"`
	} `json:"history"`
}

func (s *ProjectMihomoService) assignProviderNodes(ctx context.Context, settings *ProjectMihomoSettings, proxies []ProjectMihomoProxy) (int, error) {
	client, err := httpclient.GetClient(httpclient.Options{
		Timeout: projectMihomoHTTPTimeout,
	})
	if err != nil {
		return 0, fmt.Errorf("build controller client: %w", err)
	}

	_ = s.refreshProvider(ctx, client, settings)
	deadline := time.Now().Add(projectMihomoAssignWait)
	nodes := make([]ProjectMihomoNode, 0)
	for {
		nodes, err = s.providerNodes(ctx, client, settings)
		if err != nil {
			return 0, err
		}
		if len(nodes) > 0 || time.Now().After(deadline) {
			break
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(projectMihomoAssignPoll):
		}
	}
	if len(nodes) == 0 {
		return 0, nil
	}

	assigned := 0
	fallbackIndex := 0
	regionOffsets := make(map[string]int, len(proxies))
	for i := range proxies {
		var selectedNode *ProjectMihomoNode
		regionFilter := ""
		if i < len(settings.ListenerRegions) {
			regionFilter = settings.ListenerRegions[i]
		}
		if matched := findProjectMihomoNode(nodes, regionFilter); matched != nil {
			selectedNode = matched
		}
		candidates := filterProjectMihomoNodeEntriesByRegion(nodes, regionFilter)
		if len(candidates) == 0 {
			if selectedNode == nil {
				selectedNode = &nodes[fallbackIndex%len(nodes)]
				fallbackIndex++
			}
		} else if selectedNode == nil {
			key := normalizeProjectMihomoRegionText(regionFilter)
			selectedNode = &candidates[regionOffsets[key]%len(candidates)]
			regionOffsets[key]++
		}
		if selectedNode == nil {
			continue
		}
		if err := s.selectProxyGroup(ctx, client, settings, proxies[i].Name, selectedNode.Name); err != nil {
			return assigned, err
		}
		assigned++
	}
	return assigned, nil
}

func (s *ProjectMihomoService) refreshProvider(ctx context.Context, client *http.Client, settings *ProjectMihomoSettings) error {
	for _, provider := range buildProjectMihomoProviderRefs(settings) {
		req, err := s.controllerRequest(ctx, http.MethodPut, settings, "/providers/proxies/"+url.PathEscape(provider.Name), nil)
		if err != nil {
			return err
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("refresh mihomo provider: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("refresh mihomo provider: unexpected status %d", resp.StatusCode)
		}
	}
	return nil
}

func (s *ProjectMihomoService) providerNodeNames(ctx context.Context, client *http.Client, settings *ProjectMihomoSettings) ([]string, error) {
	nodes, err := s.providerNodes(ctx, client, settings)
	if err != nil {
		return nil, err
	}
	return projectMihomoNodeNames(nodes), nil
}

func (s *ProjectMihomoService) waitProviderNodes(ctx context.Context, client *http.Client, settings *ProjectMihomoSettings) ([]ProjectMihomoNode, error) {
	deadline := time.Now().Add(projectMihomoAssignWait)
	for {
		nodes, err := s.providerNodes(ctx, client, settings)
		if err != nil {
			return nil, err
		}
		if len(nodes) > 0 || time.Now().After(deadline) {
			return nodes, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(projectMihomoAssignPoll):
		}
	}
}

func (s *ProjectMihomoService) providerNodes(ctx context.Context, client *http.Client, settings *ProjectMihomoSettings) ([]ProjectMihomoNode, error) {
	req, err := s.controllerRequest(ctx, http.MethodGet, settings, "/providers/proxies", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get mihomo providers: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("get mihomo providers: unexpected status %d", resp.StatusCode)
	}

	var payload mihomoProvidersResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode mihomo providers: %w", err)
	}

	allNodes := make([]ProjectMihomoNode, 0)
	providerRefs := buildProjectMihomoProviderRefs(settings)
	for i, providerRef := range providerRefs {
		provider, ok := payload.Providers[providerRef.Name]
		if !ok {
			continue
		}
		providerLabel := projectMihomoProviderDisplayName(providerRef, i, len(providerRefs))

		aliveSet := make(map[string]struct{}, len(provider.Proxies))
		providerNodes := make([]ProjectMihomoNode, 0, len(provider.Proxies))
		for i := range provider.Proxies {
			item := provider.Proxies[i]
			name := strings.TrimSpace(item.Name)
			if !isProjectMihomoSelectableNodeName(name) {
				continue
			}
			if _, ok := aliveSet[name]; ok {
				continue
			}
			if item.Alive {
				aliveSet[name] = struct{}{}
			}
			latency := latestProjectMihomoDelay(item)
			providerNodes = append(providerNodes, ProjectMihomoNode{
				Key:           projectMihomoNodeKey(providerRef.Name, name),
				Name:          name,
				Region:        extractProjectMihomoRegion(name),
				Alive:         item.Alive,
				Provider:      providerRef.Name,
				ProviderLabel: providerLabel,
				LatencyMS:     latency,
				LatencyStatus: projectMihomoLatencyStatus(item.Alive, latency),
			})
		}

		cachedNames, err := s.providerNodeNamesFromCachePath(s.providerCachePathFor(providerRef.Path))
		if err == nil && len(cachedNames) > 0 {
			byName := make(map[string]ProjectMihomoNode, len(providerNodes))
			for i := range providerNodes {
				byName[providerNodes[i].Name] = providerNodes[i]
			}
			filtered := make([]ProjectMihomoNode, 0, len(cachedNames))
			for i := range cachedNames {
				name := cachedNames[i]
				if !isProjectMihomoSelectableNodeName(name) {
					continue
				}
				node, ok := byName[name]
				if !ok || !node.Alive {
					continue
				}
				filtered = append(filtered, node)
			}
			if len(filtered) > 0 {
				providerNodes = filtered
			}
		}

		for i := range providerNodes {
			node := providerNodes[i]
			if !node.Alive {
				continue
			}
			allNodes = append(allNodes, node)
		}
	}
	return allNodes, nil
}

func (s *ProjectMihomoService) providerNodeNamesFromCache() ([]string, error) {
	return s.providerNodeNamesFromCachePath(s.providerCachePath())
}

func (s *ProjectMihomoService) providerNodeNamesFromCachePath(cachePath string) ([]string, error) {
	content, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	return parseProjectMihomoSubscriptionNodeNames(content), nil
}

func latestProjectMihomoDelay(proxy mihomoProviderProxy) *int {
	for i := len(proxy.History) - 1; i >= 0; i-- {
		delay := proxy.History[i].Delay
		if delay > 0 {
			return &delay
		}
	}
	return nil
}

func projectMihomoLatencyStatus(alive bool, latency *int) string {
	if latency != nil {
		return "success"
	}
	if !alive {
		return "failed"
	}
	return "unknown"
}

func projectMihomoNodeNames(nodes []ProjectMihomoNode) []string {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]string, 0, len(nodes))
	seen := make(map[string]struct{}, len(nodes))
	for i := range nodes {
		name := strings.TrimSpace(nodes[i].Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func filterProjectMihomoNodeEntriesByRegion(nodes []ProjectMihomoNode, regionFilter string) []ProjectMihomoNode {
	if len(nodes) == 0 {
		return nil
	}
	if provider, name, ok := parseProjectMihomoNodeKey(regionFilter); ok {
		for i := range nodes {
			if strings.TrimSpace(nodes[i].Provider) == provider && strings.TrimSpace(nodes[i].Name) == name {
				return []ProjectMihomoNode{nodes[i]}
			}
		}
		fallback := make([]ProjectMihomoNode, 0, 1)
		for i := range nodes {
			if strings.TrimSpace(nodes[i].Name) == name {
				fallback = append(fallback, nodes[i])
			}
		}
		if len(fallback) > 0 {
			return fallback
		}
		regionFilter = name
	}

	filteredNames := filterProjectMihomoNodesByRegion(projectMihomoNodeNames(nodes), regionFilter)
	if len(filteredNames) == 0 {
		return nil
	}

	out := make([]ProjectMihomoNode, 0, len(filteredNames))
	for i := range filteredNames {
		target := filteredNames[i]
		for j := range nodes {
			if strings.TrimSpace(nodes[j].Name) == target {
				out = append(out, nodes[j])
			}
		}
	}
	return out
}

func findProjectMihomoNode(nodes []ProjectMihomoNode, value string) *ProjectMihomoNode {
	target := strings.TrimSpace(value)
	if target == "" {
		return nil
	}
	for i := range nodes {
		if strings.TrimSpace(nodes[i].Key) == target {
			return &nodes[i]
		}
	}
	if provider, name, ok := parseProjectMihomoNodeKey(target); ok {
		for i := range nodes {
			if strings.TrimSpace(nodes[i].Provider) == provider && strings.TrimSpace(nodes[i].Name) == name {
				return &nodes[i]
			}
		}
		target = name
	}
	for i := range nodes {
		name := strings.TrimSpace(nodes[i].Name)
		if name == target {
			return &nodes[i]
		}
	}
	normalizedTarget := normalizeProjectMihomoRegionText(target)
	for i := range nodes {
		name := strings.TrimSpace(nodes[i].Name)
		if normalizeProjectMihomoRegionText(name) == normalizedTarget {
			return &nodes[i]
		}
	}
	return nil
}

func (s *ProjectMihomoService) testNodeDelays(ctx context.Context, client *http.Client, settings *ProjectMihomoSettings, nodes []ProjectMihomoNode) []ProjectMihomoNode {
	if len(nodes) == 0 {
		return nodes
	}

	out := make([]ProjectMihomoNode, len(nodes))
	copy(out, nodes)
	jobs := make(chan int)
	var wg sync.WaitGroup
	workerCount := projectMihomoDelayWorkers
	if len(out) < workerCount {
		workerCount = len(out)
	}
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for index := range jobs {
				if out[index].LatencyMS != nil {
					out[index].LatencyStatus = "success"
					out[index].LatencyMessage = ""
					continue
				}
				latency, err := s.testNodeDelay(ctx, client, settings, out[index])
				if err != nil {
					out[index].LatencyMS = nil
					out[index].LatencyStatus = "failed"
					out[index].LatencyMessage = err.Error()
					continue
				}
				out[index].LatencyMS = &latency
				out[index].LatencyStatus = "success"
				out[index].LatencyMessage = ""
				out[index].Alive = true
			}
		}()
	}
	for i := range out {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return out
		case jobs <- i:
		}
	}
	close(jobs)
	wg.Wait()
	return out
}

func (s *ProjectMihomoService) testNodeDelay(ctx context.Context, client *http.Client, settings *ProjectMihomoSettings, node ProjectMihomoNode) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(projectMihomoDelayTimeoutMS+1000)*time.Millisecond)
	defer cancel()

	nodeName := strings.TrimSpace(node.Name)
	paths := []string{"/proxies/" + url.PathEscape(nodeName) + "/delay"}
	if strings.TrimSpace(node.Provider) != "" {
		paths = append([]string{
			"/providers/proxies/" + url.PathEscape(strings.TrimSpace(node.Provider)) + "/" + url.PathEscape(nodeName) + "/healthcheck",
		}, paths...)
	}
	query := url.Values{}
	query.Set("url", projectMihomoDelayURL)
	query.Set("timeout", fmt.Sprintf("%d", projectMihomoDelayTimeoutMS))

	var lastErr error
	for i, path := range paths {
		latency, err := s.testNodeDelayPath(ctx, client, settings, path+"?"+query.Encode())
		if err == nil {
			return latency, nil
		}
		lastErr = err
		if i == 0 && !isProjectMihomoDelayPathFallbackError(err) {
			break
		}
	}
	return 0, lastErr
}

func (s *ProjectMihomoService) testNodeDelayPath(ctx context.Context, client *http.Client, settings *ProjectMihomoSettings, path string) (int, error) {
	req, err := s.controllerRequest(ctx, http.MethodGet, settings, path, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("test mihomo node delay: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("test mihomo node delay: unexpected status %d", resp.StatusCode)
	}
	var payload struct {
		Delay int `json:"delay"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, fmt.Errorf("decode mihomo node delay: %w", err)
	}
	if payload.Delay <= 0 {
		return 0, fmt.Errorf("test mihomo node delay: empty delay")
	}
	return payload.Delay, nil
}

func (s *ProjectMihomoService) ensureProviderFiles(ctx context.Context, previous, current *ProjectMihomoSettings) error {
	if current == nil {
		return nil
	}
	providers := buildProjectMihomoProviderRefs(current)
	if len(providers) == 0 {
		return nil
	}
	if err := os.MkdirAll(s.providerDir(), 0o755); err != nil {
		return fmt.Errorf("create mihomo provider dir: %w", err)
	}

	previousURLs := make(map[string]string, len(buildProjectMihomoProviderRefs(previous)))
	previousRefs := buildProjectMihomoProviderRefs(previous)
	for _, ref := range previousRefs {
		previousURLs[ref.Name] = strings.TrimSpace(ref.URL)
	}

	client, err := httpclient.GetClient(httpclient.Options{
		Timeout: projectMihomoHTTPTimeout,
	})
	if err != nil {
		return fmt.Errorf("build subscription client: %w", err)
	}

	for _, provider := range providers {
		targetPath := s.providerCachePathFor(provider.Path)
		currentURL := strings.TrimSpace(provider.URL)
		previousURL := previousURLs[provider.Name]
		if currentURL == previousURL {
			if _, err := os.Stat(targetPath); err == nil {
				continue
			}
		}
		if reused, err := s.restoreProviderFileFromPrevious(previousRefs, currentURL, targetPath); err != nil {
			return err
		} else if reused {
			continue
		}
		content, err := s.fetchProviderContent(ctx, client, current.SubscriptionUA, currentURL)
		if err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, content, 0o644); err != nil {
			return fmt.Errorf("write mihomo provider file: %w", err)
		}
	}
	return s.pruneProviderFiles(current)
}

func (s *ProjectMihomoService) restoreProviderFileFromPrevious(previousRefs []projectMihomoProviderRef, currentURL, targetPath string) (bool, error) {
	currentURL = strings.TrimSpace(currentURL)
	if currentURL == "" {
		return false, nil
	}
	for _, ref := range previousRefs {
		if strings.TrimSpace(ref.URL) != currentURL {
			continue
		}
		sourcePath := s.providerCachePathFor(ref.Path)
		content, err := os.ReadFile(sourcePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return false, fmt.Errorf("read cached mihomo provider file: %w", err)
		}
		if err := os.WriteFile(targetPath, content, 0o644); err != nil {
			return false, fmt.Errorf("write mihomo provider file: %w", err)
		}
		return true, nil
	}
	return false, nil
}

func (s *ProjectMihomoService) cleanupProviderFiles(previous, current *ProjectMihomoSettings) error {
	currentRefs := make(map[string]struct{}, len(buildProjectMihomoProviderRefs(current)))
	for _, ref := range buildProjectMihomoProviderRefs(current) {
		currentRefs[ref.Name] = struct{}{}
	}
	entries, err := os.ReadDir(s.providerDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read mihomo provider dir: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if !strings.HasPrefix(name, projectMihomoProviderName) {
			continue
		}
		if _, ok := currentRefs[name]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(s.providerDir(), entry.Name())); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove mihomo provider file: %w", err)
		}
	}
	return nil
}

func (s *ProjectMihomoService) pruneProviderFiles(current *ProjectMihomoSettings) error {
	expected := make(map[string]struct{}, len(buildProjectMihomoProviderRefs(current)))
	for _, ref := range buildProjectMihomoProviderRefs(current) {
		expected[filepath.Base(ref.Path)] = struct{}{}
	}
	entries, err := os.ReadDir(s.providerDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("list mihomo provider files: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, projectMihomoProviderName) || !strings.HasSuffix(strings.ToLower(name), ".yaml") {
			continue
		}
		if _, ok := expected[name]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(s.providerDir(), name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove stale mihomo provider file: %w", err)
		}
	}
	return nil
}

func (s *ProjectMihomoService) fetchProviderContent(ctx context.Context, client *http.Client, userAgent, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(rawURL), nil)
	if err != nil {
		return nil, ErrProjectMihomoSubscriptionFetch.WithCause(fmt.Errorf("build subscription request: %w", err))
	}
	if strings.TrimSpace(userAgent) != "" {
		req.Header.Set("User-Agent", strings.TrimSpace(userAgent))
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrProjectMihomoSubscriptionFetch.WithCause(fmt.Errorf("request subscription: %w", err))
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ErrProjectMihomoSubscriptionFetch.WithCause(fmt.Errorf("request subscription: unexpected status %d", resp.StatusCode))
	}
	limited := io.LimitReader(resp.Body, projectMihomoProviderMaxSize+1)
	content, err := io.ReadAll(limited)
	if err != nil {
		return nil, ErrProjectMihomoSubscriptionFetch.WithCause(fmt.Errorf("read subscription: %w", err))
	}
	if len(content) == 0 {
		return nil, ErrProjectMihomoSubscriptionFetch.WithCause(fmt.Errorf("read subscription: empty response"))
	}
	if len(content) > projectMihomoProviderMaxSize {
		return nil, ErrProjectMihomoSubscriptionFetch.WithCause(fmt.Errorf("read subscription: response too large"))
	}
	return content, nil
}

func isProjectMihomoDelayPathFallbackError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "unexpected status 404") ||
		strings.Contains(message, "unexpected status 405")
}

func buildProjectMihomoProviderRefs(settings *ProjectMihomoSettings) []projectMihomoProviderRef {
	if settings == nil {
		return nil
	}
	urls := normalizeProjectMihomoSubscriptionURLs(settings.SubscriptionURLs, settings.SubscriptionURL)
	if len(urls) == 0 {
		return nil
	}

	out := make([]projectMihomoProviderRef, 0, len(urls))
	if len(urls) == 1 {
		label := strings.TrimSpace("")
		if len(settings.SubscriptionNames) > 0 {
			label = strings.TrimSpace(settings.SubscriptionNames[0])
		}
		return []projectMihomoProviderRef{{
			Name:        projectMihomoProviderName,
			Path:        projectMihomoProviderPath,
			URL:         urls[0],
			DisplayName: label,
		}}
	}

	for i := range urls {
		suffix := fmt.Sprintf("-%02d", i+1)
		label := ""
		if i < len(settings.SubscriptionNames) {
			label = strings.TrimSpace(settings.SubscriptionNames[i])
		}
		out = append(out, projectMihomoProviderRef{
			Name:        projectMihomoProviderName + suffix,
			Path:        "./providers/" + projectMihomoProviderName + suffix + ".yaml",
			URL:         urls[i],
			DisplayName: label,
		})
	}
	return out
}

func projectMihomoProviderDisplayName(providerRef projectMihomoProviderRef, index int, total int) string {
	if providerRef.DisplayName != "" {
		return providerRef.DisplayName
	}
	host := strings.TrimSpace(projectMihomoProviderHost(providerRef.URL))
	prefix := ""
	if total > 1 {
		prefix = fmt.Sprintf("#%d ", index+1)
	}
	if host != "" {
		return prefix + host
	}
	return prefix + providerRef.Name
}

func projectMihomoProviderHost(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Hostname())
}

func projectMihomoNodeKey(providerName string, nodeName string) string {
	name := strings.TrimSpace(nodeName)
	if name == "" {
		return ""
	}
	provider := strings.TrimSpace(providerName)
	if provider == "" {
		return name
	}
	return url.QueryEscape(provider) + "::" + url.QueryEscape(name)
}

func parseProjectMihomoNodeKey(value string) (string, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(value), "::", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	provider, err := url.QueryUnescape(parts[0])
	if err != nil {
		return "", "", false
	}
	name, err := url.QueryUnescape(parts[1])
	if err != nil {
		return "", "", false
	}
	provider = strings.TrimSpace(provider)
	name = strings.TrimSpace(name)
	if provider == "" || name == "" {
		return "", "", false
	}
	return provider, name, true
}

func isProjectMihomoSelectableNodeName(name string) bool {
	value := strings.TrimSpace(name)
	if value == "" {
		return false
	}
	lower := strings.ToLower(value)
	blockedPrefixes := []string{
		"剩余流量", "距离下次重置", "套餐到期", "防走失", "如果用不了",
		"traffic:", "expire:", "subscription", "官网", "更新订阅",
	}
	for i := range blockedPrefixes {
		if strings.HasPrefix(lower, strings.ToLower(blockedPrefixes[i])) {
			return false
		}
	}
	return true
}

type mihomoSubscriptionConfig struct {
	Proxies []mihomoSubscriptionProxy `yaml:"proxies"`
}

type mihomoSubscriptionProxy struct {
	Name   string `yaml:"name"`
	Server string `yaml:"server"`
}

func parseProjectMihomoSubscriptionNodeNames(content []byte) []string {
	texts := candidateSubscriptionTexts(content)
	for i := range texts {
		if names := parseProjectMihomoYAMLNodeNames(texts[i]); len(names) > 0 {
			return names
		}
		if names := parseProjectMihomoURINodeNames(texts[i]); len(names) > 0 {
			return names
		}
	}
	return nil
}

func candidateSubscriptionTexts(content []byte) []string {
	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return nil
	}

	out := []string{trimmed}
	for _, decoder := range []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	} {
		decoded, err := decoder.DecodeString(trimmed)
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(decoded))
		if text == "" || text == trimmed {
			continue
		}
		out = append([]string{text}, out...)
		break
	}
	return out
}

func parseProjectMihomoYAMLNodeNames(text string) []string {
	var payload mihomoSubscriptionConfig
	if err := yaml.Unmarshal([]byte(text), &payload); err != nil || len(payload.Proxies) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(payload.Proxies))
	out := make([]string, 0, len(payload.Proxies))
	for i := range payload.Proxies {
		item := payload.Proxies[i]
		if strings.EqualFold(strings.TrimSpace(item.Server), projectMihomoPlaceholderHost) {
			continue
		}
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func parseProjectMihomoURINodeNames(text string) []string {
	lines := strings.Split(text, "\n")
	seen := make(map[string]struct{}, len(lines))
	out := make([]string, 0, len(lines))
	for i := range lines {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		parsed, err := url.Parse(line)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(parsed.Hostname()), projectMihomoPlaceholderHost) {
			continue
		}
		name, err := url.QueryUnescape(parsed.Fragment)
		if err != nil {
			name = parsed.Fragment
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func extractProjectMihomoRegions(nodes []string) []string {
	if len(nodes) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(nodes))
	out := make([]string, 0, len(nodes))
	for i := range nodes {
		region := extractProjectMihomoRegion(nodes[i])
		if region == "" {
			continue
		}
		key := normalizeProjectMihomoRegionText(region)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, region)
	}
	return out
}

func extractProjectMihomoRegion(nodeName string) string {
	runes := []rune(strings.TrimSpace(nodeName))
	if len(runes) == 0 {
		return ""
	}

	start := -1
	end := -1
	for i, r := range runes {
		if start < 0 {
			if unicode.IsLetter(r) {
				start = i
				end = i + 1
			}
			continue
		}
		if unicode.IsDigit(r) || isProjectMihomoRegionDelimiter(r) {
			end = i
			break
		}
		end = i + 1
	}
	if start < 0 || end <= start {
		return ""
	}

	region := strings.TrimSpace(string(runes[start:end]))
	region = strings.Trim(region, " -_./:|")
	return strings.TrimSpace(region)
}

func isProjectMihomoRegionDelimiter(r rune) bool {
	switch r {
	case '[', ']', '(', ')', '（', '）', '【', '】', '{', '}', '<', '>', '|', '\\', '/', '#', '@':
		return true
	default:
		return false
	}
}

func filterProjectMihomoNodesByRegion(nodes []string, regionFilter string) []string {
	normalizedFilter := normalizeProjectMihomoRegionText(regionFilter)
	if normalizedFilter == "" {
		return nodes
	}

	tokens := projectMihomoRegionAliases[normalizedFilter]
	if len(tokens) == 0 {
		tokens = []string{regionFilter}
	}

	normalizedTokens := make([]string, 0, len(tokens))
	for i := range tokens {
		token := normalizeProjectMihomoRegionText(tokens[i])
		if token != "" {
			normalizedTokens = append(normalizedTokens, token)
		}
	}
	if len(normalizedTokens) == 0 {
		return nodes
	}

	filtered := make([]string, 0, len(nodes))
	for i := range nodes {
		normalizedName := normalizeProjectMihomoRegionText(nodes[i])
		for j := range normalizedTokens {
			if strings.Contains(normalizedName, normalizedTokens[j]) {
				filtered = append(filtered, nodes[i])
				break
			}
		}
	}
	return filtered
}

func normalizeProjectMihomoRegionText(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		" ", "",
		"-", "",
		"_", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		".", "",
	)
	return replacer.Replace(normalized)
}

func (s *ProjectMihomoService) selectProxyGroup(ctx context.Context, client *http.Client, settings *ProjectMihomoSettings, groupName, nodeName string) error {
	body, err := json.Marshal(map[string]string{
		"name": nodeName,
	})
	if err != nil {
		return fmt.Errorf("marshal proxy group selection: %w", err)
	}
	req, err := s.controllerRequest(ctx, http.MethodPut, settings, "/proxies/"+url.PathEscape(groupName), body)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("select mihomo proxy group: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("select mihomo proxy group: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (s *ProjectMihomoService) controllerRequest(ctx context.Context, method string, settings *ProjectMihomoSettings, path string, body []byte) (*http.Request, error) {
	target := strings.TrimRight(settings.ControllerURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, target, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build mihomo controller request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if settings.ControllerSecret != "" {
		req.Header.Set("Authorization", "Bearer "+settings.ControllerSecret)
	}
	return req, nil
}
