package handler

import (
	"context"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type helpGroupAuthorizer interface {
	GetAvailableGroups(context.Context, int64) ([]service.Group, error)
}

// Reuse the gateway model listing, including allowlists, without exposing user keys.
func helpModelsHandler(authorizer helpGroupAuthorizer, models gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "private, no-store")
		subject, ok := middleware.GetAuthSubjectFromContext(c)
		if !ok {
			response.Unauthorized(c, "User not authenticated")
			return
		}
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil || id <= 0 {
			response.BadRequest(c, "Invalid group ID")
			return
		}
		groups, err := authorizer.GetAvailableGroups(c.Request.Context(), subject.UserID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for i := range groups {
			if groups[i].ID == id && groups[i].Status == service.StatusActive {
				c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{GroupID: &groups[i].ID, Group: &groups[i]})
				models(c)
				return
			}
		}
		response.NotFound(c, "Group not found")
	}
}

func (h *APIKeyHandler) HelpModels(gateway *GatewayHandler) gin.HandlerFunc {
	return helpModelsHandler(h.apiKeyService, gateway.Models)
}
