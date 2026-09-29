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
	models := h.gatewayService.GetAvailableModels(ctx, &groupID, group.Platform)
	if len(models) == 0 {
		models = defaultModelIDsForPlatform(group.Platform)
	}
	if group.ModelAllowlistEnabled() {
		models = group.ModelAllowlist.FilterForListing(models)
	}
	return models
}
