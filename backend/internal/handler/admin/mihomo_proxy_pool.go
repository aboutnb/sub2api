package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const proxyProviderProjectMihomo = "project_mihomo"

type projectMihomoSettingsSnapshot struct {
	Protocol        string   `json:"protocol"`
	TargetHost      string   `json:"target_host"`
	ListenerCount   int      `json:"listener_count"`
	ListenerPorts   []int    `json:"listener_ports"`
	ListenerNames   []string `json:"listener_names"`
	ProxyNamePrefix string   `json:"proxy_name_prefix"`
}

type projectMihomoProxyCandidate struct {
	Proxy           service.Proxy
	AccountCount    int64
	LatencyMs       *int64
	QualityScore    *int
	QualityStatus   string
	BatchAssigned   int
	OriginalOrdinal int
}

type projectMihomoProxyAllocator struct {
	candidates []projectMihomoProxyCandidate
}

func (h *AccountHandler) resolveProjectMihomoProxyID(
	ctx context.Context,
	explicitProxyID *int64,
	proxyProvider string,
	allocator *projectMihomoProxyAllocator,
) (*int64, error) {
	if explicitProxyID != nil {
		return explicitProxyID, nil
	}
	if !isProjectMihomoProxyProvider(proxyProvider) {
		return explicitProxyID, nil
	}
	if allocator == nil {
		return nil, fmt.Errorf("project mihomo proxy allocator is not initialized")
	}
	return allocator.Next()
}

func (h *AccountHandler) newProjectMihomoProxyAllocator(ctx context.Context, requested bool) (*projectMihomoProxyAllocator, error) {
	if !requested {
		return nil, nil
	}
	if h.adminService == nil {
		return nil, fmt.Errorf("admin service unavailable")
	}
	if h.settingService == nil {
		return nil, fmt.Errorf("setting service unavailable")
	}

	settings, err := h.loadProjectMihomoSettings(ctx)
	if err != nil {
		return nil, err
	}
	if settings.ListenerCount <= 0 {
		return nil, fmt.Errorf("project mihomo has no configured listener ports")
	}

	proxies, err := h.adminService.GetAllProxiesWithAccountCount(ctx)
	if err != nil {
		return nil, err
	}

	candidates := make([]projectMihomoProxyCandidate, 0, len(proxies))
	for i := range proxies {
		item := proxies[i]
		if !isProjectMihomoManagedProxyRow(settings, &item.Proxy) {
			continue
		}
		if !item.Proxy.IsActive() {
			continue
		}
		candidates = append(candidates, projectMihomoProxyCandidate{
			Proxy:           item.Proxy,
			AccountCount:    item.AccountCount,
			LatencyMs:       item.LatencyMs,
			QualityScore:    item.QualityScore,
			QualityStatus:   strings.TrimSpace(strings.ToLower(item.QualityStatus)),
			OriginalOrdinal: len(candidates),
		})
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("project mihomo proxy pool has no active synced proxies")
	}

	return &projectMihomoProxyAllocator{candidates: candidates}, nil
}

func (a *projectMihomoProxyAllocator) Next() (*int64, error) {
	if a == nil || len(a.candidates) == 0 {
		return nil, fmt.Errorf("project mihomo proxy pool is empty")
	}

	best := 0
	for i := 1; i < len(a.candidates); i++ {
		if compareProjectMihomoCandidates(a.candidates[i], a.candidates[best]) < 0 {
			best = i
		}
	}

	a.candidates[best].BatchAssigned++
	id := a.candidates[best].Proxy.ID
	return &id, nil
}

func compareProjectMihomoCandidates(a, b projectMihomoProxyCandidate) int {
	loadA := a.AccountCount + int64(a.BatchAssigned)
	loadB := b.AccountCount + int64(b.BatchAssigned)
	if loadA != loadB {
		if loadA < loadB {
			return -1
		}
		return 1
	}

	scoreA := normalizedProjectMihomoQualityScore(a.QualityStatus, a.QualityScore)
	scoreB := normalizedProjectMihomoQualityScore(b.QualityStatus, b.QualityScore)
	if scoreA != scoreB {
		if scoreA > scoreB {
			return -1
		}
		return 1
	}

	latA := normalizedProjectMihomoLatency(a.LatencyMs)
	latB := normalizedProjectMihomoLatency(b.LatencyMs)
	if latA != latB {
		if latA < latB {
			return -1
		}
		return 1
	}

	if a.Proxy.Port != b.Proxy.Port {
		if a.Proxy.Port < b.Proxy.Port {
			return -1
		}
		return 1
	}

	if a.Proxy.ID != b.Proxy.ID {
		if a.Proxy.ID < b.Proxy.ID {
			return -1
		}
		return 1
	}

	if a.OriginalOrdinal < b.OriginalOrdinal {
		return -1
	}
	if a.OriginalOrdinal > b.OriginalOrdinal {
		return 1
	}
	return 0
}

func normalizedProjectMihomoQualityScore(status string, score *int) int {
	base := 0
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "pass", "success":
		base = 1000
	case "warn":
		base = 700
	case "challenge":
		base = 400
	case "fail", "failed":
		base = 100
	default:
		base = 500
	}
	if score != nil {
		base += *score
	}
	return base
}

func normalizedProjectMihomoLatency(latency *int64) int64 {
	if latency == nil || *latency <= 0 {
		return 1<<62 - 1
	}
	return *latency
}

func (h *AccountHandler) loadProjectMihomoSettings(ctx context.Context) (*projectMihomoSettingsSnapshot, error) {
	raw, err := h.settingService.GetRawValue(ctx, service.SettingKeyProjectMihomoSettings)
	if err != nil {
		return nil, err
	}

	settings := projectMihomoSettingsSnapshot{}
	if strings.TrimSpace(raw) == "" {
		return &settings, nil
	}
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return nil, fmt.Errorf("decode project mihomo settings: %w", err)
	}

	normalizeProjectMihomoSettingsSnapshot(&settings)
	return &settings, nil
}

func normalizeProjectMihomoSettingsSnapshot(settings *projectMihomoSettingsSnapshot) {
	if settings == nil {
		return
	}
	if settings.ListenerCount < 0 {
		settings.ListenerCount = 0
	}
	settings.Protocol = strings.TrimSpace(strings.ToLower(settings.Protocol))
	settings.TargetHost = strings.TrimSpace(settings.TargetHost)
	settings.ProxyNamePrefix = strings.TrimSpace(settings.ProxyNamePrefix)
	if settings.ProxyNamePrefix == "" {
		settings.ProxyNamePrefix = "project-mihomo"
	}

	ports := make([]int, 0, settings.ListenerCount)
	for _, port := range settings.ListenerPorts {
		if port > 0 {
			ports = append(ports, port)
		}
		if settings.ListenerCount > 0 && len(ports) >= settings.ListenerCount {
			break
		}
	}
	settings.ListenerPorts = ports

	names := make([]string, 0, settings.ListenerCount)
	for _, name := range settings.ListenerNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		names = append(names, name)
		if settings.ListenerCount > 0 && len(names) >= settings.ListenerCount {
			break
		}
	}
	settings.ListenerNames = names
}

func isProjectMihomoManagedProxyRow(settings *projectMihomoSettingsSnapshot, proxy *service.Proxy) bool {
	if settings == nil || proxy == nil {
		return false
	}

	name := strings.TrimSpace(strings.ToLower(proxy.Name))
	if name == "" {
		return false
	}

	expectedNames := make(map[string]struct{}, len(settings.ListenerNames))
	for _, item := range settings.ListenerNames {
		key := strings.TrimSpace(strings.ToLower(item))
		if key == "" {
			continue
		}
		expectedNames[key] = struct{}{}
	}
	if _, ok := expectedNames[name]; ok {
		return true
	}

	expectedPorts := make([]int, 0, len(settings.ListenerPorts))
	for _, port := range settings.ListenerPorts {
		if port > 0 {
			expectedPorts = append(expectedPorts, port)
		}
	}
	sort.Ints(expectedPorts)
	if len(expectedPorts) == 0 {
		return false
	}
	if proxy.Port <= 0 || !containsInt(expectedPorts, proxy.Port) {
		return false
	}

	if settings.TargetHost != "" && !strings.EqualFold(strings.TrimSpace(proxy.Host), settings.TargetHost) {
		return false
	}
	if settings.Protocol != "" && !strings.EqualFold(strings.TrimSpace(proxy.Protocol), settings.Protocol) {
		return false
	}

	prefix := strings.TrimSpace(strings.ToLower(settings.ProxyNamePrefix))
	if prefix == "" {
		return false
	}
	return strings.HasPrefix(name, prefix+"-")
}

func containsInt(items []int, target int) bool {
	index := sort.SearchInts(items, target)
	return index < len(items) && items[index] == target
}

func isProjectMihomoProxyProvider(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), proxyProviderProjectMihomo)
}
