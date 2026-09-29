package service

import (
	"context"
	"math"
	"sort"
	"strings"
)

const (
	ImageStudioSmartGroupName        = "智能分组"
	ImageStudioSmartGroupDescription = "AI 绘图提示词优化自动创建，聚合多个公开渠道，按本分组倍率计费。image-studio-prompt"
)

// IsImageStudioSmartGroup reports the hidden prompt-optimization pool.
// It is not a normal API-key group and must not be bound by users.
func IsImageStudioSmartGroup(group *Group) bool {
	return group != nil && group.Name == ImageStudioSmartGroupName && group.Description == ImageStudioSmartGroupDescription
}

const imageStudioDefaultGPTModel = "gpt-5.6-sol"

// ImageStudioPromptModel picks a chat model suitable for rewriting an image prompt.
// gpt-5.6-sol is the default when the group lists it. Otherwise smaller general
// models win. Image, audio, and embedding models are ignored.
func ImageStudioPromptModel(models []string) string {
	for _, model := range models {
		trimmed := strings.TrimSpace(model)
		if strings.EqualFold(trimmed, imageStudioDefaultGPTModel) {
			return trimmed
		}
	}
	best, score := "", 0
	for _, model := range models {
		current := imageStudioPromptScore(model)
		if current > score {
			best, score = strings.TrimSpace(model), current
		}
	}
	return best
}

// ImageStudioPromptModels returns the group's chat models in list order.
// A non-positive limit keeps the full list. Image, audio, and embedding names are omitted.
func ImageStudioPromptModels(models []string, limit int) []string {
	out := make([]string, 0, len(models))
	seen := map[string]struct{}{}
	for _, model := range models {
		trimmed := strings.TrimSpace(model)
		if imageStudioPromptScore(trimmed) == 0 {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func imageStudioPromptScore(model string) int {
	lower := strings.ToLower(strings.TrimSpace(model))
	if lower == "" {
		return 0
	}
	for _, blocked := range []string{"image", "embed", "whisper", "tts", "dall", "moderation", "realtime", "audio", "transcribe"} {
		if strings.Contains(lower, blocked) {
			return 0
		}
	}
	if lower == imageStudioDefaultGPTModel {
		return 10
	}
	for _, hint := range []string{"mini", "flash", "haiku", "nano"} {
		if strings.Contains(lower, hint) {
			return 3
		}
	}
	return 1
}

// BestImageStudioPromptTarget chooses one permitted group and its text model.
// The smart pool itself is never selected here; callers use it separately.
func BestImageStudioPromptTarget(groups []Group, models func(Group) []string) (Group, string) {
	var best Group
	bestModel := ""
	bestScore := 0
	found := false
	for i := range groups {
		group := groups[i]
		if IsImageStudioSmartGroup(&group) || !imageStudioPromptRateOK(group.RateMultiplier) || models == nil {
			continue
		}
		model := ImageStudioPromptModel(models(group))
		score := imageStudioPromptScore(model)
		if score == 0 {
			continue
		}
		if !found || score > bestScore || (score == bestScore && imageStudioPromptPreferred(group, best)) {
			best, bestModel, bestScore, found = group, model, score, true
		}
	}
	if !found {
		return Group{}, ""
	}
	return best, bestModel
}

func imageStudioPromptRateOK(rate float64) bool {
	return !math.IsNaN(rate) && !math.IsInf(rate, 0) && rate >= 0
}

func imageStudioPromptPreferred(candidate, current Group) bool {
	if candidate.RateMultiplier != current.RateMultiplier {
		return candidate.RateMultiplier < current.RateMultiplier
	}
	if candidate.SortOrder != current.SortOrder {
		return candidate.SortOrder < current.SortOrder
	}
	return candidate.ID < current.ID
}

type ImageStudioPromptRoute struct {
	Groups []Group
	Model  string
	Label  string
}

type imageStudioPromptSource struct {
	group Group
	model string
	score int
}

// AllowsImageStudioPublicPromptPool is true only for users who may use every
// public text channel. Restricted users stay on the groups already granted to them.
func (s *APIKeyService) AllowsImageStudioPublicPromptPool(ctx context.Context, userID int64) bool {
	if s == nil || s.userRepo == nil || userID <= 0 {
		return false
	}
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return false
	}
	return !user.RestrictPublicGroups
}

// PlanImageStudioPromptRoute lists public text channels for one prompt key.
// OpenAI is ordered first and supplies the model when it has a usable text model.
// No admin group is created.
func (s *APIKeyService) PlanImageStudioPromptRoute(ctx context.Context, models func(Group) []string) (*ImageStudioPromptRoute, error) {
	if s == nil || s.groupRepo == nil || models == nil {
		return nil, nil
	}
	groups, err := s.groupRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	sources := imageStudioPromptSources(groups, models)
	if len(sources) == 0 {
		return nil, nil
	}
	if len(sources) > SmartRouteMaxCandidates {
		sources = sources[:SmartRouteMaxCandidates]
	}
	chosen := sources[0]
	for _, source := range sources {
		if source.group.Platform == PlatformOpenAI {
			chosen = source
			break
		}
	}
	out := make([]Group, len(sources))
	for i := range sources {
		out[i] = sources[i].group
	}
	return &ImageStudioPromptRoute{Groups: out, Model: chosen.model, Label: chosen.group.Name}, nil
}

func imageStudioPromptSources(groups []Group, models func(Group) []string) []imageStudioPromptSource {
	sources := make([]imageStudioPromptSource, 0, len(groups))
	for i := range groups {
		group := groups[i]
		if !imageStudioPublicPromptSource(&group) || !imageStudioPromptRateOK(group.RateMultiplier) || models == nil {
			continue
		}
		model := ImageStudioPromptModel(models(group))
		score := imageStudioPromptScore(model)
		if score == 0 {
			continue
		}
		sources = append(sources, imageStudioPromptSource{group: group, model: model, score: score})
	}
	sort.SliceStable(sources, func(i, j int) bool {
		leftOpen := sources[i].group.Platform == PlatformOpenAI
		rightOpen := sources[j].group.Platform == PlatformOpenAI
		if leftOpen != rightOpen {
			return leftOpen
		}
		if sources[i].score != sources[j].score {
			return sources[i].score > sources[j].score
		}
		return imageStudioPromptPreferred(sources[i].group, sources[j].group)
	})
	return sources
}

func imageStudioPublicPromptSource(group *Group) bool {
	return group != nil && group.IsActive() && !group.IsExclusive && !group.IsSubscriptionType() && group.Platform != PlatformComposite && !IsImageStudioSmartGroup(group)
}
