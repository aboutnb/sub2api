package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	SmartRouteModeSingle = "single"
	SmartRouteModeSmart  = "smart"

	SmartRouteStrategyAuto    = "auto"
	SmartRouteStrategyPrice   = "price"
	SmartRouteStrategySpeed   = "speed"
	SmartRouteStrategySuccess = "success"
	SmartRouteStrategyCustom  = "custom"

	SmartRouteMaxCandidates  = 20
	SmartRouteMinSamples     = 20
	SmartRouteWindow         = 30 * time.Minute
	SmartRouteSnapshotWindow = 5 * time.Minute
	SmartRouteMaxSnapshots   = 4096
)

var (
	ErrSmartRoutingDisabled  = infraerrors.Forbidden("SMART_ROUTING_DISABLED", "Smart routing is disabled")
	ErrSmartRouteInvalid     = infraerrors.BadRequest("SMART_ROUTE_INVALID", "Invalid smart routing configuration")
	ErrSmartRouteUnavailable = infraerrors.ServiceUnavailable(
		"SMART_ROUTE_NO_AVAILABLE_GROUP",
		"No smart routing candidate is currently available",
	)
	ErrSmartRouteModelUnsupported = infraerrors.NotFound(
		"MODEL_NOT_FOUND",
		"The requested model is not available in any smart routing candidate group",
	)
	ErrSmartRouteEndpointUnsupported = infraerrors.BadRequest(
		"SMART_ROUTE_ENDPOINT_UNSUPPORTED",
		"This endpoint is not supported by smart routing",
	)
)

type SmartRouteWeights struct {
	Price   int `json:"price"`
	Speed   int `json:"speed"`
	Success int `json:"success"`
}

type SmartRouteRateGuard struct {
	Enabled                bool     `json:"enabled"`
	MaxRateMultiplier      *float64 `json:"max_rate_multiplier"`
	MaxImageRateMultiplier *float64 `json:"max_image_rate_multiplier"`
}

type SmartRouteInput struct {
	Mode              string               `json:"mode"`
	GroupID           *int64               `json:"group_id,omitempty"`
	CandidateGroupIDs []int64              `json:"candidate_group_ids,omitempty"`
	Strategy          string               `json:"strategy,omitempty"`
	Weights           *SmartRouteWeights   `json:"weights,omitempty"`
	RateGuard         *SmartRouteRateGuard `json:"rate_guard,omitempty"`
}

type SmartRouteConfig struct {
	APIKeyID          int64               `json:"-"`
	Mode              string              `json:"mode"`
	Platform          string              `json:"platform,omitempty"`
	SubscriptionType  string              `json:"subscription_type,omitempty"`
	CandidateGroupIDs []int64             `json:"candidate_group_ids,omitempty"`
	Strategy          string              `json:"strategy"`
	Weights           SmartRouteWeights   `json:"weights"`
	RateGuard         SmartRouteRateGuard `json:"rate_guard"`
	UpdatedAt         time.Time           `json:"updated_at,omitempty"`
	RuntimeGroups     []*Group            `json:"-"`
}

type SmartRouteMetric struct {
	GroupID      int64
	Samples      int64
	SuccessRate  float64
	LatencyP50MS float64
}

type SmartRouteRequest struct {
	Method     string
	Path       string
	Model      string
	Kind       string
	Streaming  bool
	ClaudeCode bool
}

func (r SmartRouteRequest) IsImage() bool { return r.Kind == "image" || r.Kind == "batch_image" }

func (r SmartRouteRequest) IsBatchImage() bool { return r.Kind == "batch_image" }

// IsHistoricalImageTaskRead identifies owner-scoped async task polling. The
// handler validates task ownership, so polling must not depend on a live route.
func (r SmartRouteRequest) IsHistoricalImageTaskRead() bool {
	if r.Method != "GET" {
		return false
	}
	path := strings.TrimRight(strings.ToLower(r.Path), "/")
	return strings.HasPrefix(path, "/v1/images/tasks/") || strings.HasPrefix(path, "/images/tasks/")
}

// IsHistoricalBatchImageRead identifies owner-scoped batch endpoints that only
// read or mutate an already-created job. They must not require a currently
// available candidate group; the job is keyed by user and API key ownership.
func (r SmartRouteRequest) IsHistoricalBatchImageRead() bool {
	path := strings.TrimRight(strings.ToLower(r.Path), "/")
	if !strings.Contains(path, "/images/batches") || strings.HasSuffix(path, "/models") {
		return false
	}
	if r.Method == "GET" {
		return true
	}
	// Cancel/delete operate on an existing owner-scoped batch. The submit
	// endpoint has no trailing id and remains a live smart-routing request.
	return (r.Method == "POST" || r.Method == "DELETE") && strings.Contains(path, "/images/batches/")
}

type SmartRouteCandidate struct {
	Group               *Group
	Position            int
	EffectiveMultiplier float64
	Metric              SmartRouteMetric
	PriceScore          float64
	SpeedScore          float64
	SuccessScore        float64
	TotalScore          float64
}

type SmartRouteRepository interface {
	WithinTransaction(ctx context.Context, fn func(context.Context) error) error
	Get(ctx context.Context, apiKeyID int64) (*SmartRouteConfig, error)
	GetMany(ctx context.Context, apiKeyIDs []int64) (map[int64]*SmartRouteConfig, error)
	Replace(ctx context.Context, config *SmartRouteConfig) error
	Delete(ctx context.Context, apiKeyID int64) error
	GetMetrics(ctx context.Context, groupIDs []int64, model, metric string, start, end time.Time) (map[int64]SmartRouteMetric, error)
	Invalidate(ctx context.Context, apiKeyID int64)
}

type smartRouteRankingSnapshot struct {
	groupIDs  []int64
	expiresAt time.Time
}

type SmartRouteService struct {
	repo          SmartRouteRepository
	apiKeys       *APIKeyService
	groups        GroupRepository
	userRates     UserGroupRateRepository
	scheduler     *SchedulerSnapshotService
	settings      *SettingService
	subscriptions *SubscriptionService
	rankingMu     sync.Mutex
	rankings      map[string]smartRouteRankingSnapshot
}

func NewSmartRouteService(
	repo SmartRouteRepository,
	apiKeys *APIKeyService,
	groups GroupRepository,
	userRates UserGroupRateRepository,
	scheduler *SchedulerSnapshotService,
	settings *SettingService,
	subscriptions *SubscriptionService,
) *SmartRouteService {
	return &SmartRouteService{
		repo: repo, apiKeys: apiKeys, groups: groups, userRates: userRates,
		scheduler: scheduler, settings: settings, subscriptions: subscriptions, rankings: make(map[string]smartRouteRankingSnapshot),
	}
}

func SmartRoutePresetWeights(strategy string) (SmartRouteWeights, bool) {
	switch strategy {
	case SmartRouteStrategyAuto, "":
		return SmartRouteWeights{Price: 40, Speed: 30, Success: 30}, true
	case SmartRouteStrategyPrice:
		return SmartRouteWeights{Price: 100}, true
	case SmartRouteStrategySpeed:
		return SmartRouteWeights{Price: 15, Speed: 70, Success: 15}, true
	case SmartRouteStrategySuccess:
		return SmartRouteWeights{Price: 15, Speed: 15, Success: 70}, true
	default:
		return SmartRouteWeights{}, false
	}
}

func NormalizeSmartRouteInput(input SmartRouteInput) (SmartRouteInput, error) {
	input.Mode = strings.ToLower(strings.TrimSpace(input.Mode))
	if input.Mode == "" {
		input.Mode = SmartRouteModeSingle
	}
	if input.Mode == SmartRouteModeSingle {
		if input.GroupID == nil || *input.GroupID <= 0 || len(input.CandidateGroupIDs) != 0 {
			return input, ErrSmartRouteInvalid.WithCause(errors.New("single mode requires group_id and forbids candidate_group_ids"))
		}
		return input, nil
	}
	if input.Mode != SmartRouteModeSmart || input.GroupID != nil {
		return input, ErrSmartRouteInvalid.WithCause(errors.New("smart mode requires candidate_group_ids and forbids group_id"))
	}
	if len(input.CandidateGroupIDs) == 0 || len(input.CandidateGroupIDs) > SmartRouteMaxCandidates {
		return input, ErrSmartRouteInvalid.WithCause(fmt.Errorf("candidate_group_ids must contain 1-%d groups", SmartRouteMaxCandidates))
	}
	seen := make(map[int64]struct{}, len(input.CandidateGroupIDs))
	for _, id := range input.CandidateGroupIDs {
		if id <= 0 {
			return input, ErrSmartRouteInvalid.WithCause(errors.New("candidate group ids must be positive"))
		}
		if _, exists := seen[id]; exists {
			return input, ErrSmartRouteInvalid.WithCause(errors.New("candidate group ids must be unique"))
		}
		seen[id] = struct{}{}
	}
	input.Strategy = strings.ToLower(strings.TrimSpace(input.Strategy))
	if input.Strategy == "" {
		input.Strategy = SmartRouteStrategyAuto
	}
	if weights, ok := SmartRoutePresetWeights(input.Strategy); ok {
		input.Weights = &weights
	} else if input.Strategy == SmartRouteStrategyCustom {
		if input.Weights == nil || input.Weights.Price < 0 || input.Weights.Price > 100 || input.Weights.Speed < 0 || input.Weights.Speed > 100 || input.Weights.Success < 0 || input.Weights.Success > 100 || input.Weights.Price+input.Weights.Speed+input.Weights.Success != 100 {
			return input, ErrSmartRouteInvalid.WithCause(errors.New("custom weights must be integers from 0 to 100 and sum to 100"))
		}
	} else {
		return input, ErrSmartRouteInvalid.WithCause(errors.New("unknown routing strategy"))
	}
	if input.RateGuard == nil {
		defaultMaxRate := 1.0
		input.RateGuard = &SmartRouteRateGuard{Enabled: true, MaxRateMultiplier: &defaultMaxRate}
	}
	for _, limit := range []*float64{input.RateGuard.MaxRateMultiplier, input.RateGuard.MaxImageRateMultiplier} {
		if limit != nil && (math.IsNaN(*limit) || math.IsInf(*limit, 0) || *limit < 0) {
			return input, ErrSmartRouteInvalid.WithCause(errors.New("rate guard limits must be finite and non-negative"))
		}
	}
	return input, nil
}

func ScoreSmartRouteCandidates(candidates []SmartRouteCandidate, weights SmartRouteWeights) []SmartRouteCandidate {
	if len(candidates) == 0 {
		return candidates
	}
	minRate, minLatency := math.Inf(1), math.Inf(1)
	for i := range candidates {
		if candidates[i].EffectiveMultiplier < minRate {
			minRate = candidates[i].EffectiveMultiplier
		}
		if candidates[i].Metric.LatencyP50MS > 0 && candidates[i].Metric.LatencyP50MS < minLatency {
			minLatency = candidates[i].Metric.LatencyP50MS
		}
	}
	for i := range candidates {
		if minRate == 0 {
			if candidates[i].EffectiveMultiplier == 0 {
				candidates[i].PriceScore = 100
			}
		} else if candidates[i].EffectiveMultiplier > 0 {
			candidates[i].PriceScore = 100 * minRate / candidates[i].EffectiveMultiplier
		}
		if candidates[i].Metric.Samples < SmartRouteMinSamples || candidates[i].Metric.LatencyP50MS <= 0 || math.IsInf(minLatency, 1) {
			candidates[i].SpeedScore = 50
		} else {
			candidates[i].SpeedScore = 100 * minLatency / candidates[i].Metric.LatencyP50MS
		}
		if candidates[i].Metric.Samples < SmartRouteMinSamples {
			candidates[i].SuccessScore = 50
		} else {
			candidates[i].SuccessScore = 100 * candidates[i].Metric.SuccessRate
		}
		candidates[i].TotalScore = (candidates[i].PriceScore*float64(weights.Price) + candidates[i].SpeedScore*float64(weights.Speed) + candidates[i].SuccessScore*float64(weights.Success)) / 100
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.TotalScore != b.TotalScore {
			return a.TotalScore > b.TotalScore
		}
		if a.EffectiveMultiplier != b.EffectiveMultiplier {
			return a.EffectiveMultiplier < b.EffectiveMultiplier
		}
		if a.Metric.SuccessRate != b.Metric.SuccessRate {
			return a.Metric.SuccessRate > b.Metric.SuccessRate
		}
		if a.Metric.LatencyP50MS != b.Metric.LatencyP50MS {
			return a.Metric.LatencyP50MS < b.Metric.LatencyP50MS
		}
		if a.Position != b.Position {
			return a.Position < b.Position
		}
		return a.Group.ID < b.Group.ID
	})
	return candidates
}

func (s *SmartRouteService) Enabled(ctx context.Context) bool {
	return s != nil && s.settings != nil && s.settings.IsSmartRoutingEnabled(ctx)
}

func (s *SmartRouteService) GetConfig(ctx context.Context, apiKeyID int64) (*SmartRouteConfig, error) {
	if s == nil || s.repo == nil || apiKeyID <= 0 {
		return nil, nil
	}
	return s.repo.Get(ctx, apiKeyID)
}

func (s *SmartRouteService) GetConfigs(ctx context.Context, apiKeyIDs []int64) (map[int64]*SmartRouteConfig, error) {
	if s == nil || s.repo == nil || len(apiKeyIDs) == 0 {
		return map[int64]*SmartRouteConfig{}, nil
	}
	return s.repo.GetMany(ctx, apiKeyIDs)
}

func (s *SmartRouteService) validateSmartGroups(ctx context.Context, userID int64, input SmartRouteInput) (*SmartRouteConfig, error) {
	available, err := s.apiKeys.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, err
	}
	allowed := make(map[int64]*Group, len(available))
	for i := range available {
		allowed[available[i].ID] = &available[i]
	}
	var platform, subscriptionType string
	for _, id := range input.CandidateGroupIDs {
		group := allowed[id]
		if group == nil || !group.IsActive() {
			return nil, ErrGroupNotAllowed
		}
		if group.Platform == PlatformComposite {
			return nil, ErrSmartRouteInvalid.WithCause(errors.New("composite groups are not supported"))
		}
		if platform == "" {
			platform, subscriptionType = group.Platform, group.SubscriptionType
		}
		if group.Platform != platform || group.SubscriptionType != subscriptionType {
			return nil, ErrSmartRouteInvalid.WithCause(errors.New("candidate groups must use the same platform and billing type"))
		}
	}
	return &SmartRouteConfig{Mode: SmartRouteModeSmart, Platform: platform, SubscriptionType: subscriptionType, CandidateGroupIDs: append([]int64(nil), input.CandidateGroupIDs...), Strategy: input.Strategy, Weights: *input.Weights, RateGuard: *input.RateGuard}, nil
}

func (s *SmartRouteService) CreateAPIKey(ctx context.Context, userID int64, req CreateAPIKeyRequest, routing *SmartRouteInput) (*APIKey, *SmartRouteConfig, error) {
	if routing == nil {
		key, err := s.apiKeys.Create(ctx, userID, req)
		return key, nil, err
	}
	normalized, err := NormalizeSmartRouteInput(*routing)
	if err != nil {
		return nil, nil, err
	}
	if normalized.Mode == SmartRouteModeSingle {
		req.GroupID = normalized.GroupID
		key, err := s.apiKeys.Create(ctx, userID, req)
		return key, nil, err
	}
	if !s.Enabled(ctx) {
		return nil, nil, ErrSmartRoutingDisabled
	}
	config, err := s.validateSmartGroups(ctx, userID, normalized)
	if err != nil {
		return nil, nil, err
	}
	var key *APIKey
	err = s.repo.WithinTransaction(ctx, func(txCtx context.Context) error {
		req.GroupID = nil
		created, createErr := s.apiKeys.Create(txCtx, userID, req)
		if createErr != nil {
			return createErr
		}
		key = created
		config.APIKeyID = key.ID
		return s.repo.Replace(txCtx, config)
	})
	if err != nil {
		return nil, nil, err
	}
	s.apiKeys.InvalidateAuthCacheByKey(ctx, key.Key)
	s.repo.Invalidate(ctx, key.ID)
	s.invalidateRankings(key.ID)
	return key, config, nil
}

func (s *SmartRouteService) UpdateAPIKey(ctx context.Context, id, userID int64, req UpdateAPIKeyRequest, routing *SmartRouteInput, legacyGroupProvided bool) (*APIKey, *SmartRouteConfig, error) {
	var config *SmartRouteConfig
	var normalized SmartRouteInput
	var err error
	if routing != nil {
		normalized, err = NormalizeSmartRouteInput(*routing)
		if err != nil {
			return nil, nil, err
		}
		if normalized.Mode == SmartRouteModeSmart {
			if !s.Enabled(ctx) {
				return nil, nil, ErrSmartRoutingDisabled
			}
			config, err = s.validateSmartGroups(ctx, userID, normalized)
			if err != nil {
				return nil, nil, err
			}
			config.APIKeyID = id
		} else {
			req.GroupID = normalized.GroupID
		}
	}
	err = s.repo.WithinTransaction(ctx, func(txCtx context.Context) error {
		if _, updateErr := s.apiKeys.Update(txCtx, id, userID, req); updateErr != nil {
			return updateErr
		}
		if routing != nil && normalized.Mode == SmartRouteModeSmart {
			return s.repo.Replace(txCtx, config)
		}
		if routing != nil || legacyGroupProvided {
			return s.repo.Delete(txCtx, id)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	key, err := s.apiKeys.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	s.apiKeys.InvalidateAuthCacheByKey(ctx, key.Key)
	s.repo.Invalidate(ctx, key.ID)
	s.invalidateRankings(key.ID)
	if routing == nil && !legacyGroupProvided {
		config, _ = s.repo.Get(ctx, id)
	}
	return key, config, nil
}

func (r SmartRouteRequest) Supported() bool {
	path := strings.ToLower(r.Path)
	if strings.Contains(path, "/video") || strings.Contains(path, "/audio") || strings.Contains(path, "/live") ||
		strings.Contains(path, "/realtime") || strings.Contains(path, "/search") || strings.Contains(path, "/alpha/search") {
		return false
	}
	if r.Method == "GET" && strings.HasSuffix(strings.TrimRight(path, "/"), "/responses") {
		return false
	}
	if r.Method == "GET" && (strings.HasSuffix(path, "/models") || strings.HasSuffix(path, "/usage") || strings.HasSuffix(path, "/sub2api/billing")) {
		return true
	}
	if r.Method == "GET" && strings.Contains(path, "/models/") {
		return true
	}
	return strings.Contains(path, "/messages") ||
		(r.Method == "POST" && strings.Contains(path, "/responses")) ||
		strings.Contains(path, "/chat/completions") ||
		strings.Contains(path, "/embeddings") ||
		strings.Contains(path, "/count_tokens") ||
		strings.Contains(path, ":counttokens") ||
		strings.Contains(path, ":generatecontent") ||
		strings.Contains(path, ":streamgeneratecontent") ||
		strings.Contains(path, "/images")
}

func (r SmartRouteRequest) NeedsAccount() bool {
	path := strings.ToLower(r.Path)
	if r.Method == "GET" && (strings.HasSuffix(path, "/models") || strings.HasSuffix(path, "/usage") || strings.HasSuffix(path, "/sub2api/billing")) {
		return false
	}
	return r.Method != "GET" || (!strings.Contains(path, "/images/batches") && !strings.Contains(path, "/images/tasks/"))
}

func (s *SmartRouteService) Resolve(ctx context.Context, apiKey *APIKey, request SmartRouteRequest) (*APIKey, *SmartRouteConfig, error) {
	if s == nil || apiKey == nil {
		return apiKey, nil, nil
	}
	config, err := s.GetConfig(ctx, apiKey.ID)
	if err != nil || config == nil {
		return apiKey, config, err
	}
	if request.IsHistoricalBatchImageRead() || request.IsHistoricalImageTaskRead() {
		return apiKey, config, nil
	}
	if !s.Enabled(ctx) {
		return nil, config, ErrSmartRoutingDisabled
	}
	if !request.Supported() {
		return nil, config, ErrSmartRouteEndpointUnsupported
	}

	available, err := s.apiKeys.GetAvailableGroups(ctx, apiKey.UserID)
	if err != nil {
		return nil, config, err
	}
	allowed := make(map[int64]*Group, len(available))
	for i := range available {
		allowed[available[i].ID] = &available[i]
	}

	now := time.Now()
	epoch := now.UTC().Truncate(SmartRouteSnapshotWindow)
	metricName := "duration"
	if request.Streaming && !request.IsImage() {
		metricName = "ttft"
	}
	metrics, metricErr := s.repo.GetMetrics(ctx, config.CandidateGroupIDs, request.Model, metricName, epoch.Add(-SmartRouteWindow), epoch)
	if metricErr == nil && request.Model != "" {
		fallback, fallbackErr := s.repo.GetMetrics(ctx, config.CandidateGroupIDs, "", metricName, epoch.Add(-SmartRouteWindow), epoch)
		if fallbackErr == nil {
			for id, value := range metrics {
				if value.Samples < SmartRouteMinSamples {
					if groupValue, ok := fallback[id]; ok {
						metrics[id] = groupValue
					}
				}
			}
			for id, value := range fallback {
				if _, ok := metrics[id]; !ok {
					metrics[id] = value
				}
			}
		}
	}
	if metricErr != nil {
		metrics = map[int64]SmartRouteMetric{}
	}

	candidates := make([]SmartRouteCandidate, 0, len(config.CandidateGroupIDs))
	hadSchedulableAccounts := false
	hadModelSupport := false
	for position, id := range config.CandidateGroupIDs {
		group := allowed[id]
		if group == nil || !group.IsActive() || group.Platform != config.Platform || group.SubscriptionType != config.SubscriptionType || group.Platform == PlatformComposite {
			continue
		}
		if group.ClaudeCodeOnly && !request.ClaudeCode {
			continue
		}
		// Subscription candidates must have an active subscription with remaining
		// quota before ranking. The auth middleware performs the authoritative
		// check again after selection; this pre-filter prevents an exhausted
		// candidate from blocking a healthy alternative group.
		if group.IsSubscriptionType() && s.subscriptions != nil {
			subscription, subErr := s.subscriptions.GetActiveSubscription(ctx, apiKey.UserID, group.ID)
			if subErr != nil || subscription == nil {
				continue
			}
			if _, limitErr := s.subscriptions.ValidateAndCheckLimits(subscription, group); limitErr != nil {
				continue
			}
		}
		if request.IsBatchImage() && !group.AllowBatchImageGeneration {
			continue
		}
		if request.Kind == "image" && !group.AllowImageGeneration {
			continue
		}
		if request.NeedsAccount() && s.scheduler != nil {
			accounts, _, listErr := s.scheduler.ListSchedulableAccounts(ctx, &group.ID, group.Platform, false)
			if listErr != nil {
				continue
			}
			if len(accounts) > 0 {
				hadSchedulableAccounts = true
			}
			eligible := false
			for i := range accounts {
				modelSupported := request.Model == "" || IsModelSupportedForSmartRoute(ctx, &accounts[i], request.Model)
				if modelSupported {
					hadModelSupport = true
				}
				if modelSupported && accounts[i].IsSchedulableForModelWithContext(ctx, request.Model) {
					eligible = true
					break
				}
			}
			if !eligible {
				continue
			}
		}
		baseRate := group.RateMultiplier
		if s.userRates != nil {
			override, rateErr := s.userRates.GetByUserAndGroup(ctx, apiKey.UserID, group.ID)
			if rateErr != nil {
				// Never rank with a guessed rate when the user override cannot be
				// read: that could violate the configured rate guard or pricing.
				continue
			}
			if override != nil {
				baseRate = *override
			}
		}
		if math.IsNaN(baseRate) || math.IsInf(baseRate, 0) || baseRate < 0 {
			// A malformed persisted/user override must never participate in
			// ranking or bypass the rate guard.
			continue
		}
		probeKey := *apiKey
		probeKey.GroupID = &group.ID
		probeKey.Group = group
		textRate, imageRate := computePeakAwareMultipliers(&probeKey, baseRate, now)
		effectiveRate := textRate
		if request.IsImage() {
			effectiveRate = imageRate
		}
		if config.RateGuard.Enabled {
			limit := config.RateGuard.MaxRateMultiplier
			if request.IsImage() {
				limit = config.RateGuard.MaxImageRateMultiplier
			}
			if limit != nil && effectiveRate > *limit {
				continue
			}
		}
		metric := metrics[group.ID]
		metric.GroupID = group.ID
		config.RuntimeGroups = append(config.RuntimeGroups, group)
		candidates = append(candidates, SmartRouteCandidate{Group: group, Position: position, EffectiveMultiplier: effectiveRate, Metric: metric})
	}
	if len(candidates) == 0 {
		if request.Model != "" && hadSchedulableAccounts && !hadModelSupport {
			return nil, config, ErrSmartRouteModelUnsupported
		}
		return nil, config, ErrSmartRouteUnavailable
	}
	ranked := s.snapshotRanking(apiKey.ID, request, epoch, candidates, config.Weights)
	selected := smartRouteRequestGroup(ranked[0].Group)
	clone := *apiKey
	groupID := selected.ID
	clone.GroupID = &groupID
	clone.Group = selected
	return &clone, config, nil
}

func (s *SmartRouteService) snapshotRanking(apiKeyID int64, request SmartRouteRequest, epoch time.Time, candidates []SmartRouteCandidate, weights SmartRouteWeights) []SmartRouteCandidate {
	key := strconv.FormatInt(apiKeyID, 10) + "|" + request.Model + "|" + request.Kind + "|" + strconv.FormatBool(request.Streaming) + "|" + strconv.FormatInt(epoch.Unix(), 10)
	now := time.Now()
	s.rankingMu.Lock()
	snapshot, exists := s.rankings[key]
	if exists && now.Before(snapshot.expiresAt) {
		positions := make(map[int64]int, len(snapshot.groupIDs))
		for i, id := range snapshot.groupIDs {
			positions[id] = i
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			left, leftOK := positions[candidates[i].Group.ID]
			right, rightOK := positions[candidates[j].Group.ID]
			if leftOK != rightOK {
				return leftOK
			}
			if leftOK {
				return left < right
			}
			return candidates[i].Position < candidates[j].Position
		})
		s.rankingMu.Unlock()
		return candidates
	}
	s.rankingMu.Unlock()

	ranked := ScoreSmartRouteCandidates(candidates, weights)
	ids := make([]int64, 0, len(ranked))
	for i := range ranked {
		ids = append(ids, ranked[i].Group.ID)
	}
	s.rankingMu.Lock()
	if s.rankings == nil {
		s.rankings = make(map[string]smartRouteRankingSnapshot)
	}
	s.rankings[key] = smartRouteRankingSnapshot{groupIDs: ids, expiresAt: epoch.Add(SmartRouteSnapshotWindow)}
	for cacheKey, value := range s.rankings {
		if !now.Before(value.expiresAt) {
			delete(s.rankings, cacheKey)
		}
	}
	for len(s.rankings) > SmartRouteMaxSnapshots {
		var oldestKey string
		var oldestExpiry time.Time
		for cacheKey, value := range s.rankings {
			if oldestKey == "" || value.expiresAt.Before(oldestExpiry) {
				oldestKey, oldestExpiry = cacheKey, value.expiresAt
			}
		}
		if oldestKey == "" {
			break
		}
		delete(s.rankings, oldestKey)
	}
	s.rankingMu.Unlock()
	return ranked
}

func (s *SmartRouteService) invalidateRankings(apiKeyID int64) {
	prefix := strconv.FormatInt(apiKeyID, 10) + "|"
	s.rankingMu.Lock()
	defer s.rankingMu.Unlock()
	for key := range s.rankings {
		if strings.HasPrefix(key, prefix) {
			delete(s.rankings, key)
		}
	}
}

func (s *SmartRouteService) ResolveRecordedGroup(ctx context.Context, apiKey *APIKey, groupID int64) (*APIKey, *SmartRouteConfig, error) {
	if apiKey == nil || groupID <= 0 {
		return apiKey, nil, nil
	}
	// Task history is keyed by ownership and the recorded group, so it must
	// remain readable even after a smart-route config is removed or its cache
	// becomes unavailable. A config read error is therefore non-fatal here;
	// normal live requests still fail closed in Resolve.
	config, _ := s.GetConfig(ctx, apiKey.ID)
	if s.groups == nil {
		return apiKey, config, nil
	}
	group, err := s.groups.GetByID(ctx, groupID)
	if err != nil || group == nil {
		// Historical task ownership is already checked by ImageTaskService. A
		// deleted group must not strand a result, so keep the key ungrouped and
		// let the owner-scoped handler read the recorded task.
		return apiKey, config, nil
	}
	selected := smartRouteRequestGroup(group)
	clone := *apiKey
	clone.GroupID = &selected.ID
	clone.Group = selected
	if config != nil {
		config.RuntimeGroups = []*Group{group}
	}
	return &clone, config, nil
}

func smartRouteRequestGroup(group *Group) *Group {
	if group == nil {
		return nil
	}
	clone := *group
	clone.FallbackGroupID = nil
	clone.FallbackGroupIDOnInvalidRequest = nil
	return &clone
}
