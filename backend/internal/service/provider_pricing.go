package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	providerPricingSchemaVersion = "1.1"
	providerPricingCurrency      = "CNY"
	providerPricingDefaultUnit   = "per_1m_tokens"
	providerPricingSiteName      = "aivoza"
	providerPricingSiteDomain    = "aivoza.com"
	ProviderPricingCacheTTL      = 30 * time.Minute
)

// ProviderPricingSnapshot is the cached, protocol-neutral price document.
type ProviderPricingSnapshot struct {
	SchemaVersion string
	Success       bool
	Message       string
	Currency      string
	PriceUnit     string
	SiteName      string
	SiteDomain    string
	UpdatedAt     time.Time
	Models        []ProviderPricingModel
}

type ProviderPricingModel struct {
	ModelName          string                       `json:"model_name"`
	GroupName          string                       `json:"group_name"`
	PriceUnit          string                       `json:"price_unit,omitempty"`
	InputPrice         *float64                     `json:"input_price,omitempty"`
	OutputPrice        *float64                     `json:"output_price,omitempty"`
	CacheInputPrice    *float64                     `json:"cache_input_price,omitempty"`
	CacheCreatePrice   *float64                     `json:"cache_create_price,omitempty"`
	CacheCreatePrice1h *float64                     `json:"cache_create_price_1h,omitempty"`
	UnitPrice          *float64                     `json:"unit_price,omitempty"`
	BillingBasis       *ProviderPricingBillingBasis `json:"billing_basis,omitempty"`
	Enabled            bool                         `json:"enabled"`
	Note               string                       `json:"note,omitempty"`
}

type ProviderPricingBillingBasis struct {
	Resolution string `json:"resolution"`
}

type providerPricingCandidate struct {
	name     string
	platform string
	pricing  *ChannelModelPricing
	enabled  bool
}

// ProviderPricingService builds the Hvoy-compatible snapshot and caches it for
// the configured TTL. The cache is deliberately process-local: the payload is
// cheap to rebuild and the endpoint is read-only.
type ProviderPricingService struct {
	channelRepo     ChannelRepository
	groupRepo       GroupRepository
	pricingService  *PricingService
	billingService  *BillingService
	resolver        *ModelPricingResolver
	cacheMu         sync.Mutex
	cached          *ProviderPricingSnapshot
	cachedExpiresAt time.Time
}

func NewProviderPricingService(
	channelRepo ChannelRepository,
	groupRepo GroupRepository,
	pricingService *PricingService,
	billingService *BillingService,
	resolver *ModelPricingResolver,
) *ProviderPricingService {
	return &ProviderPricingService{
		channelRepo:    channelRepo,
		groupRepo:      groupRepo,
		pricingService: pricingService,
		billingService: billingService,
		resolver:       resolver,
	}
}

func (s *ProviderPricingService) Snapshot(ctx context.Context) (*ProviderPricingSnapshot, error) {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	now := time.Now().UTC()
	if s.cached != nil && now.Before(s.cachedExpiresAt) {
		return s.cached, nil
	}

	snapshot, err := s.buildSnapshot(ctx, now)
	if err != nil {
		return nil, err
	}
	s.cached = snapshot
	s.cachedExpiresAt = now.Add(ProviderPricingCacheTTL)
	return snapshot, nil
}

func (s *ProviderPricingService) buildSnapshot(ctx context.Context, updatedAt time.Time) (*ProviderPricingSnapshot, error) {
	if s.channelRepo == nil || s.groupRepo == nil {
		return nil, fmt.Errorf("provider pricing repositories are not configured")
	}
	channels, err := s.channelRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list channels for provider pricing: %w", err)
	}
	groups, err := s.listAllGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("list groups for provider pricing: %w", err)
	}

	sort.SliceStable(channels, func(i, j int) bool {
		if channels[i].Status != channels[j].Status {
			return channels[i].Status == StatusActive
		}
		return strings.ToLower(channels[i].Name) < strings.ToLower(channels[j].Name)
	})

	type candidateMap map[string]*providerPricingCandidate
	candidates := make(map[int64]candidateMap, len(groups))
	groupByID := make(map[int64]*Group, len(groups))
	for i := range groups {
		groupByID[groups[i].ID] = &groups[i]
		candidates[groups[i].ID] = make(candidateMap)
		for _, configured := range groups[i].ModelPricing {
			for _, model := range configured.Models {
				if strings.HasSuffix(strings.TrimSpace(model), "*") || strings.TrimSpace(model) == "" {
					continue
				}
				platform := configured.Platform
				if platform == "" {
					platform = groups[i].Platform
				}
				if !providerPricingPlatformMatches(groups[i].Platform, platform) {
					continue
				}
				configuredCopy := configured.Clone()
				providerPricingAddCandidate(candidates[groups[i].ID], model, platform, &configuredCopy, false)
			}
		}
	}

	for i := range channels {
		channel := &channels[i]
		supported := channel.SupportedModels()
		fillGlobalPricingFallback(s.pricingService, supported)
		for _, groupID := range channel.GroupIDs {
			group, ok := groupByID[groupID]
			if !ok {
				continue
			}
			for _, model := range supported {
				if !providerPricingPlatformMatches(group.Platform, model.Platform) {
					continue
				}
				providerPricingAddCandidate(candidates[groupID], model.Name, model.Platform, model.Pricing,
					group.Status == StatusActive && channel.Status == StatusActive)
			}
		}
	}

	modelRows := make([]ProviderPricingModel, 0)
	rowIndex := make(map[string]int)
	for _, group := range groups {
		groupCandidates := candidates[group.ID]
		keys := make([]string, 0, len(groupCandidates))
		for key := range groupCandidates {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			candidate := groupCandidates[key]
			row, ok := s.buildModelRow(ctx, &group, candidate)
			if ok {
				key := strings.ToLower(row.GroupName) + "\x00" + strings.ToLower(row.ModelName)
				if index, exists := rowIndex[key]; exists {
					if !modelRows[index].Enabled && row.Enabled {
						modelRows[index] = row
					}
					continue
				}
				rowIndex[key] = len(modelRows)
				modelRows = append(modelRows, row)
			}
		}
	}

	return &ProviderPricingSnapshot{
		SchemaVersion: providerPricingSchemaVersion,
		Success:       true,
		Currency:      providerPricingCurrency,
		PriceUnit:     providerPricingDefaultUnit,
		SiteName:      providerPricingSiteName,
		SiteDomain:    providerPricingSiteDomain,
		UpdatedAt:     updatedAt,
		Models:        modelRows,
	}, nil
}

func (s *ProviderPricingService) listAllGroups(ctx context.Context) ([]Group, error) {
	const pageSize = 1000
	params := pagination.PaginationParams{Page: 1, PageSize: pageSize, SortBy: "id", SortOrder: pagination.SortOrderAsc}
	groups := make([]Group, 0)
	for {
		page, result, err := s.groupRepo.List(ctx, params)
		if err != nil {
			return nil, err
		}
		for _, group := range page {
			if !group.IsExclusive {
				groups = append(groups, group)
			}
		}
		if result == nil || params.Page >= result.Pages || len(page) == 0 {
			break
		}
		params.Page++
	}
	return groups, nil
}

func providerPricingPlatformMatches(groupPlatform, modelPlatform string) bool {
	if groupPlatform == PlatformComposite {
		return isConcreteRequestPlatform(modelPlatform)
	}
	return groupPlatform == modelPlatform
}

func providerPricingAddCandidate(dst map[string]*providerPricingCandidate, name, platform string, pricing *ChannelModelPricing, enabled bool) {
	name = strings.TrimSpace(name)
	platform = strings.TrimSpace(platform)
	if name == "" || platform == "" {
		return
	}
	key := strings.ToLower(platform) + "\x00" + strings.ToLower(name)
	if existing, ok := dst[key]; ok {
		existing.enabled = existing.enabled || enabled
		if existing.pricing == nil && pricing != nil {
			existing.pricing = pricing
		}
		return
	}
	dst[key] = &providerPricingCandidate{name: name, platform: platform, pricing: pricing, enabled: enabled}
}

func (s *ProviderPricingService) buildModelRow(ctx context.Context, group *Group, candidate *providerPricingCandidate) (ProviderPricingModel, bool) {
	row := ProviderPricingModel{
		ModelName: candidate.name,
		GroupName: group.Name,
		Enabled:   candidate.enabled && group.Status == StatusActive,
	}
	pricing := candidate.pricing
	if groupPricing := matchGroupModelPricing(group, candidate.name); groupPricing != nil {
		pricing = groupPricing
	}
	if pricing != nil {
		switch pricing.BillingMode {
		case BillingModeImage, BillingModePerRequest:
			resolved := &ResolvedPricing{Mode: pricing.BillingMode, RequestTiers: pricing.Intervals}
			if pricing.PerRequestPrice != nil {
				resolved.DefaultPerRequestPrice = *pricing.PerRequestPrice
			}
			return s.buildPerCallRow(row, group, resolved)
		case BillingModeVideo:
			return ProviderPricingModel{}, false
		}
	}

	var resolved *ResolvedPricing
	if s.resolver != nil {
		ctx = WithResolvedTargetPlatform(ctx, candidate.platform)
		groupID := group.ID
		resolved = s.resolver.Resolve(ctx, PricingInput{Model: candidate.name, GroupID: &groupID, Group: group})
	}
	if resolved == nil {
		resolved = &ResolvedPricing{Mode: BillingModeToken}
	}

	mode := resolved.Mode
	if mode == "" {
		mode = BillingModeToken
	}
	switch mode {
	case BillingModeToken:
		if tokenRow, ok := s.buildTokenRow(ctx, row, group, candidate.name, candidate.platform, resolved); ok {
			return tokenRow, true
		}
		return providerPricingRowFromChannelPricing(row, pricing, group)
	case BillingModeImage, BillingModePerRequest:
		return s.buildPerCallRow(row, group, resolved)
	case BillingModeVideo:
		return ProviderPricingModel{}, false
	default:
		return ProviderPricingModel{}, false
	}
}

func (s *ProviderPricingService) buildTokenRow(ctx context.Context, row ProviderPricingModel, group *Group, model, platform string, resolved *ResolvedPricing) (ProviderPricingModel, bool) {
	var input, output, cacheWrite, cacheWrite1h, cacheRead *float64
	base := resolved.BasePricing
	official := false
	if s.billingService != nil {
		if catalogPrice, err := s.billingService.GetModelPricing(model); err == nil && catalogPrice != nil && catalogPrice.InputPricePerToken > 0 {
			base = catalogPrice
			official = true
		}
	}
	// Official catalog rates are base prices; group/channel overrides and time
	// policies are only fallback sources for models absent from the catalog.
	if !official && s.billingService != nil && s.resolver != nil {
		schedule, err := s.billingService.ResolveContextPricingSchedule(ctx, s.resolver, ContextPricingScheduleInput{Model: model, Group: group, Platform: platform})
		if err == nil && schedule != nil {
			for i := range schedule.Tiers {
				tier := schedule.Tiers[i]
				if tier.Input != nil && *tier.Input > 0 {
					input, output, cacheWrite, cacheWrite1h, cacheRead = tier.Input, tier.Output, tier.CacheWrite, tier.CacheWrite1h, tier.CacheRead
					break
				}
			}
		}
	}
	if input == nil && base != nil {
		if base.InputPricePerToken > 0 {
			input = &base.InputPricePerToken
		}
		if base.OutputPricePerToken > 0 {
			output = &base.OutputPricePerToken
		}
		cacheWriteValue := base.CacheCreation5mPrice
		if cacheWriteValue <= 0 {
			cacheWriteValue = base.CacheCreationPricePerToken
		}
		if cacheWriteValue > 0 {
			cacheWrite = &cacheWriteValue
		}
		if base.CacheCreation1hPrice > 0 {
			cacheWrite1h = &base.CacheCreation1hPrice
		}
		if base.CacheReadPricePerToken > 0 {
			cacheRead = &base.CacheReadPricePerToken
		}
	}
	if input == nil {
		return ProviderPricingModel{}, false
	}
	row.PriceUnit = providerPricingDefaultUnit
	multiplier := providerPricingMultiplier(group, BillingModeToken)
	row.InputPrice = providerPricingTokenPrice(input, multiplier)
	if row.InputPrice == nil {
		return ProviderPricingModel{}, false
	}
	row.OutputPrice = providerPricingTokenPrice(output, multiplier)
	row.CacheInputPrice = providerPricingTokenPrice(cacheRead, multiplier)
	row.CacheCreatePrice = providerPricingTokenPrice(cacheWrite, multiplier)
	row.CacheCreatePrice1h = providerPricingTokenPrice(cacheWrite1h, multiplier)
	return row, true
}

func (s *ProviderPricingService) buildPerCallRow(row ProviderPricingModel, group *Group, resolved *ResolvedPricing) (ProviderPricingModel, bool) {
	price := resolved.DefaultPerRequestPrice
	if price <= 0 {
		for _, tier := range resolved.RequestTiers {
			if tier.PerRequestPrice != nil && *tier.PerRequestPrice > 0 {
				price = *tier.PerRequestPrice
				break
			}
		}
	}
	if resolved.Mode == BillingModeImage {
		if groupPrice := providerPricingImagePrice(group); groupPrice != nil {
			price = *groupPrice
		}
	}
	if price <= 0 {
		return ProviderPricingModel{}, false
	}
	row.PriceUnit = "per_call"
	row.UnitPrice = providerPricingUnitPrice(price, providerPricingMultiplier(group, resolved.Mode))
	if row.UnitPrice == nil {
		return ProviderPricingModel{}, false
	}
	return row, true
}

func providerPricingRowFromChannelPricing(row ProviderPricingModel, pricing *ChannelModelPricing, group *Group) (ProviderPricingModel, bool) {
	if pricing == nil {
		return ProviderPricingModel{}, false
	}
	mode := pricing.BillingMode
	if mode == BillingModeImage || mode == BillingModePerRequest {
		price := pricing.PerRequestPrice
		if mode == BillingModeImage {
			if groupPrice := providerPricingImagePrice(group); groupPrice != nil {
				price = groupPrice
			}
		}
		if price == nil || *price <= 0 {
			return ProviderPricingModel{}, false
		}
		row.PriceUnit = "per_call"
		row.UnitPrice = providerPricingUnitPrice(*price, providerPricingMultiplier(group, mode))
		if row.UnitPrice == nil {
			return ProviderPricingModel{}, false
		}
		return row, true
	}
	if mode == BillingModeVideo {
		return ProviderPricingModel{}, false
	}
	if pricing.InputPrice == nil || *pricing.InputPrice <= 0 {
		return ProviderPricingModel{}, false
	}
	row.PriceUnit = providerPricingDefaultUnit
	multiplier := providerPricingMultiplier(group, BillingModeToken)
	row.InputPrice = providerPricingTokenPrice(pricing.InputPrice, multiplier)
	if row.InputPrice == nil {
		return ProviderPricingModel{}, false
	}
	row.OutputPrice = providerPricingTokenPrice(pricing.OutputPrice, multiplier)
	row.CacheInputPrice = providerPricingTokenPrice(pricing.CacheReadPrice, multiplier)
	row.CacheCreatePrice = providerPricingTokenPrice(pricing.CacheWritePrice, multiplier)
	row.CacheCreatePrice1h = providerPricingTokenPrice(pricing.CacheWrite1hPrice, multiplier)
	return row, true
}

func providerPricingImagePrice(group *Group) *float64 {
	if group == nil {
		return nil
	}
	for _, price := range []*float64{group.ImagePrice2K, group.ImagePrice1K, group.ImagePrice4K} {
		if price != nil && *price > 0 {
			return price
		}
	}
	return nil
}

func providerPricingMultiplier(group *Group, mode BillingMode) float64 {
	if group == nil {
		return 1
	}
	if mode == BillingModeImage {
		return resolveImageRateMultiplier(&APIKey{Group: group}, group.RateMultiplier)
	}
	return group.RateMultiplier
}

func providerPricingTokenPrice(usdPerToken *float64, multiplier float64) *float64 {
	if usdPerToken == nil || *usdPerToken <= 0 {
		return nil
	}
	value := roundProviderPricing(*usdPerToken*1_000_000*multiplier, 1)
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return nil
	}
	return &value
}

// Balance units map 1:1 to CNY; apply the public group rate once, then round
// per-call prices to cents so low-cost image requests such as 0.06 remain exact.
func providerPricingUnitPrice(basePrice float64, multiplier float64) *float64 {
	if basePrice <= 0 || multiplier <= 0 {
		return nil
	}
	value := roundProviderPricing(basePrice*multiplier, 2)
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return nil
	}
	return &value
}

func roundProviderPricing(value float64, digits int) float64 {
	scale := math.Pow10(digits)
	rounded := math.Round(value*scale) / scale
	if rounded == 0 {
		return 0
	}
	return rounded
}
