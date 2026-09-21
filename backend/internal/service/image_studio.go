package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
)

const ImageStudioKeyPurpose = "image_studio"
const imageStudioKeyPrefix = "studio-internal-"
const imageStudioSettingKey = "image_studio_enabled"

type imageStudioCredentialContextKey struct{}

// WithImageStudioCredential is used only after JWT authentication and ownership checks.
func WithImageStudioCredential(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, imageStudioCredentialContextKey{}, key)
}

func IsImageStudioCredential(key string) bool { return strings.HasPrefix(key, imageStudioKeyPrefix) }

func isImageStudioRequest(ctx context.Context) bool {
	key, ok := ctx.Value(imageStudioCredentialContextKey{}).(string)
	return ok && IsImageStudioCredential(key)
}

type imageStudioKeyRepository interface {
	ImageStudioKey(context.Context, int64, int64, string, bool) (*APIKey, error)
}

func (s *APIKeyService) ImageStudioKey(ctx context.Context, userID, groupID int64, create bool) (*APIKey, error) {
	repo, ok := s.apiKeyRepo.(imageStudioKeyRepository)
	if !ok {
		return nil, ErrAPIKeyNotFound
	}
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return repo.ImageStudioKey(ctx, userID, groupID, imageStudioKeyPrefix+hex.EncodeToString(bytes), create)
}

func (s *SettingService) ImageStudioEnabled(ctx context.Context) bool {
	values, err := s.settingRepo.GetMultiple(ctx, []string{imageStudioSettingKey})
	return err == nil && values[imageStudioSettingKey] == "true"
}

func (s *SettingService) SetImageStudioEnabled(ctx context.Context, enabled bool) error {
	return s.settingRepo.SetMultiple(ctx, map[string]string{imageStudioSettingKey: fmt.Sprint(enabled)})
}

type ImageStudioModel struct {
	ID           string                  `json:"id"`
	Platform     string                  `json:"platform"`
	Capabilities ImageStudioCapabilities `json:"capabilities"`
}

type ImageStudioCapabilities struct {
	Edits          bool     `json:"edits"`
	Mask           bool     `json:"mask"`
	Multiple       bool     `json:"multiple"`
	MaxInputImages int      `json:"max_input_images"`
	MaxUploadBytes int      `json:"max_upload_bytes"`
	InputTypes     []string `json:"input_types"`
}

func imageStudioCapabilities(platform string) ImageStudioCapabilities {
	maxImages := 16
	if platform == PlatformGrok {
		maxImages = grokMediaMaxEditSourceImages
	}
	return ImageStudioCapabilities{true, true, true, maxImages, openAIImageMaxUploadPartSize, []string{"image/png", "image/jpeg", "image/webp"}}
}

func studioAccountSupportsImages(account *Account, model string) bool {
	if !account.IsModelSupported(model) {
		return false
	}
	return studioAccountImageTargetSupported(account, model)
}

func studioAccountImageTargetSupported(account *Account, model string) bool {
	target := account.GetMappedModel(model)
	switch account.Platform {
	case PlatformOpenAI:
		return account.SupportsOpenAIImageCapability(OpenAIImagesCapabilityBasic) && IsGPTImageGenerationModel(target)
	case PlatformGrok:
		eligible, _ := account.GrokMediaGenerationEligibility()
		return eligible && isGrokImageGenerationModel(target)
	}
	return false
}

// ImageStudioModels uses the same account mappings and group allowlist as the gateway.
// Unknown/wildcard aliases are not advertised as concrete image models.
func (s *GatewayService) ImageStudioModels(ctx context.Context, group Group) ([]ImageStudioModel, error) {
	accounts, err := s.accountRepo.ListSchedulableByGroupID(ctx, group.ID)
	if err != nil {
		return nil, err
	}
	models := map[string]string{}
	var channel *Channel
	if s.channelService != nil {
		channel, err = s.channelService.GetChannelForGroup(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		if channel != nil {
			for platform, mapping := range channel.ModelMapping {
				if platform != PlatformOpenAI && platform != PlatformGrok {
					continue
				}
				for alias := range mapping {
					if !strings.ContainsAny(alias, "*?") {
						models[alias] = platform
					}
				}
			}
		}
	}
	supports := func(account *Account, model string) bool {
		if !account.IsModelSupported(model) {
			return false
		}
		mapped := model
		if channel != nil {
			mapping := s.channelService.ResolveChannelMapping(ctx, group.ID, model)
			mapped = mapping.MappedModel
			billingModel := billingModelForRestriction(mapping.BillingModelSource, model, mapped)
			if billingModel == "" {
				billingModel = resolveOpenAIAccountUpstreamModelForRequest(account, model, false)
			}
			if s.channelService.IsModelRestricted(ctx, group.ID, billingModel) {
				return false
			}
		}
		return studioAccountImageTargetSupported(account, mapped)
	}
	for _, account := range accounts {
		if account.Platform != PlatformOpenAI && account.Platform != PlatformGrok {
			continue
		}
		if group.Platform != PlatformComposite && group.Platform != account.Platform {
			continue
		}
		mapping := account.GetModelMapping()
		for alias, target := range mapping {
			if strings.ContainsAny(alias, "*?") || !isOpenAIImageGenerationModel(target) {
				continue
			}
			models[alias] = account.Platform
		}
		{
			defaults := openai.DefaultModelIDs()
			if account.Platform == PlatformGrok {
				defaults = xai.DefaultModelIDs()
			}
			for _, model := range defaults {
				if isOpenAIImageGenerationModel(model) && studioAccountSupportsImages(&account, model) {
					models[model] = account.Platform
				}
			}
			if group.ModelAllowlistEnabled() {
				for _, model := range group.ModelAllowlist.Models {
					if !strings.ContainsAny(model, "*?") && studioAccountSupportsImages(&account, model) {
						models[model] = account.Platform
					}
				}
			}
		}
	}
	if group.Platform == PlatformComposite && s.compositeResolver != nil {
		if s.compositeResolver.repo != nil {
			routes, err := s.compositeResolver.repo.ListByGroup(ctx, group.ID, false)
			if err != nil {
				return nil, err
			}
			for _, route := range routes {
				if route.Enabled && route.MatchType == CompositeRouteMatchExact && (route.Endpoint == CompositeRouteEndpointImages || route.Endpoint == CompositeRouteEndpointAny) && (route.TargetPlatform == PlatformOpenAI || route.TargetPlatform == PlatformGrok) {
					models[route.PublicModel] = route.TargetPlatform
				}
			}
		}
		for model := range models {
			decision, err := s.compositeResolver.Resolve(ctx, group.ID, model, CompositeRouteEndpointImages)
			if err != nil {
				return nil, err
			}
			supported := false
			for i := range accounts {
				if accounts[i].Platform == decision.TargetPlatform && supports(&accounts[i], decision.UpstreamModel) {
					supported = true
					break
				}
			}
			if !decision.Matched || !supported {
				delete(models, model)
			} else {
				models[model] = decision.TargetPlatform
			}
		}
	} else {
		for model := range models {
			supported := false
			for i := range accounts {
				if accounts[i].Platform == group.Platform && supports(&accounts[i], model) {
					supported = true
					break
				}
			}
			if !supported {
				delete(models, model)
			}
		}
	}
	ids := make([]string, 0, len(models))
	for id := range models {
		if group.ModelAllowlistEnabled() && !group.ModelAllowlist.Allows(id) {
			continue
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]ImageStudioModel, 0, len(ids))
	for _, id := range ids {
		result = append(result, ImageStudioModel{ID: id, Platform: models[id], Capabilities: imageStudioCapabilities(models[id])})
	}
	return result, nil
}
