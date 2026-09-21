package handler

import (
	"context"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (h *GatewayHandler) ImageStudioModels(ctx context.Context, group service.Group) ([]service.ImageStudioModel, error) {
	return h.gatewayService.ImageStudioModels(ctx, group)
}

func (h *AsyncImageHandler) ImageStudioAsyncEnabled() bool { return h.enabled() }
