package handler

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (h *GatewayHandler) ImageStudioModels(ctx context.Context, group service.Group) ([]service.ImageStudioModel, error) {
	return h.gatewayService.ImageStudioModels(ctx, group)
}

func (h *AsyncImageHandler) ImageStudioAsyncEnabled() bool { return h.enabled() }

func (h *GatewayHandler) ImageStudioTextModels(ctx context.Context, group service.Group) []string {
	if h == nil || h.gatewayService == nil {
		return nil
	}
	groupID := group.ID
	platforms := h.gatewayService.GetSchedulablePlatforms(ctx, &groupID)
	if len(platforms) == 0 {
		return nil
	}
	if group.Platform == service.PlatformComposite {
		models := h.compositeAvailableModels(ctx, &groupID, "", false)
		if group.ModelAllowlistEnabled() {
			models = group.ModelAllowlist.FilterForListing(models)
		}
		return models
	}
	if _, ok := platforms[group.Platform]; !ok {
		return nil
	}
	models := h.gatewayService.GetAvailableModels(ctx, &groupID, group.Platform)
	if len(models) == 0 && !service.IsCNProvider(group.Platform) {
		models = defaultModelIDsForPlatform(group.Platform)
	}
	if group.ModelAllowlistEnabled() {
		models = group.ModelAllowlist.FilterForListing(models)
	}
	return models
}
