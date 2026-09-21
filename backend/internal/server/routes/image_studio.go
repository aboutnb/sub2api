package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type imageStudioGroup struct {
	ID               int64                      `json:"id"`
	Name             string                     `json:"name"`
	Models           []service.ImageStudioModel `json:"models"`
	SubscriptionType string                     `json:"subscription_type"`
	RateMultiplier   float64                    `json:"rate_multiplier"`
}

func RegisterImageStudioRoutes(v1 *gin.RouterGroup, h *handler.Handlers, jwt middleware.JWTAuthMiddleware, admin middleware.AdminAuthMiddleware, audit middleware.AuditLogMiddleware, limiter *middleware.PanelRateLimiter, apiAuth middleware.APIKeyAuthMiddleware, resolver middleware.APIKeyGroupResolver, keys *service.APIKeyService, subscriptions *service.SubscriptionService, ops *service.OpsService, settings *service.SettingService, composite *service.CompositeRouteResolver, cfg *config.Config) {
	// Private in-process router preserves the gateway middleware/handler chain without network hops.
	gateway := gin.New()
	var trustedProxies []string
	if cfg.Server.TrustedProxiesConfigured {
		trustedProxies = cfg.Server.TrustedProxies
	}
	if err := gateway.SetTrustedProxies(trustedProxies); err != nil {
		_ = gateway.SetTrustedProxies(nil)
	}
	RegisterGatewayRoutes(gateway, h, apiAuth, resolver, keys, subscriptions, ops, settings, composite, cfg)
	routes := v1.Group("/image-studio", gin.HandlerFunc(jwt), middleware.BackendModeUserGuard(settings), limiter.Global())
	routes.GET("/bootstrap", func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		subject, ok := middleware.GetAuthSubjectFromContext(c)
		if !ok {
			response.Unauthorized(c, "User not authenticated")
			return
		}
		groups, err := keys.GetAvailableGroups(c.Request.Context(), subject.UserID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		sort.SliceStable(groups, func(i, j int) bool {
			if groups[i].SortOrder == groups[j].SortOrder {
				return groups[i].ID < groups[j].ID
			}
			return groups[i].SortOrder < groups[j].SortOrder
		})
		result := []imageStudioGroup{}
		rates, err := keys.GetUserGroupRates(c.Request.Context(), subject.UserID)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		for _, group := range groups {
			if !service.GroupAllowsImageGeneration(&group) {
				continue
			}
			models, err := h.Gateway.ImageStudioModels(c.Request.Context(), group)
			if err != nil {
				response.ErrorFrom(c, err)
				return
			}
			if len(models) > 0 {
				if rate, ok := rates[group.ID]; ok {
					group.RateMultiplier = rate
				}
				result = append(result, imageStudioGroup{group.ID, group.Name, models, group.SubscriptionType, group.RateMultiplier})
			}
		}
		response.Success(c, gin.H{"enabled": settings.ImageStudioEnabled(c.Request.Context()), "storage_namespace": "image-studio-user-" + strconv.FormatInt(subject.UserID, 10), "groups": result, "async_enabled": h.AsyncImage.ImageStudioAsyncEnabled(), "max_body_bytes": cfg.Gateway.MaxBodySize})
	})
	dispatch := func(c *gin.Context) {
		subject, ok := middleware.GetAuthSubjectFromContext(c)
		if !ok {
			response.Unauthorized(c, "User not authenticated")
			return
		}
		groupID, err := strconv.ParseInt(c.Param("group_id"), 10, 64)
		if err != nil || groupID <= 0 {
			response.BadRequest(c, "Invalid group")
			return
		}
		read := c.Request.Method == http.MethodGet
		if !read {
			if !settings.ImageStudioEnabled(c.Request.Context()) {
				response.Forbidden(c, "Image studio is disabled")
				return
			}
			groups, err := keys.GetAvailableGroups(c.Request.Context(), subject.UserID)
			if err != nil {
				response.ErrorFrom(c, err)
				return
			}
			var group *service.Group
			for i := range groups {
				if groups[i].ID == groupID && service.GroupAllowsImageGeneration(&groups[i]) {
					group = &groups[i]
					break
				}
			}
			if group == nil {
				response.Forbidden(c, "Image generation is not allowed for this group")
				return
			}
			model, err := imageStudioRequestModel(c.Request)
			if err != nil {
				response.BadRequest(c, "Invalid image request")
				return
			}
			models, err := h.Gateway.ImageStudioModels(c.Request.Context(), *group)
			if err != nil {
				response.ErrorFrom(c, err)
				return
			}
			allowed := false
			for _, item := range models {
				if item.ID == model {
					allowed = true
					break
				}
			}
			if !allowed {
				response.Forbidden(c, "Image model is no longer available")
				return
			}
		}
		key, err := keys.ImageStudioKey(c.Request.Context(), subject.UserID, groupID, !read)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		if key.UserID != subject.UserID || key.GroupID == nil || *key.GroupID != groupID {
			response.Forbidden(c, "Invalid task owner")
			return
		}
		req := c.Request.Clone(service.WithImageStudioCredential(c.Request.Context(), key.Key))
		suffix := strings.TrimPrefix(c.FullPath(), "/api/v1/image-studio/groups/:group_id")
		if read {
			suffix = "/images/tasks/" + c.Param("task_id")
		}
		req.URL.Path = "/v1" + suffix
		req.URL.RawPath = ""
		req.URL.RawQuery = ""
		req.Header.Del("X-API-Key")
		req.Header.Del("X-Goog-API-Key")
		req.Header.Set("Authorization", "Bearer "+key.Key)
		gateway.ServeHTTP(c.Writer, req)
	}
	for _, path := range []string{"/generations", "/edits", "/generations/async", "/edits/async"} {
		routes.POST("/groups/:group_id/images"+path, middleware.RequestBodyLimit(cfg.Gateway.MaxBodySize), dispatch)
	}
	routes.GET("/groups/:group_id/images/tasks/:task_id", dispatch)
	adminRoutes := v1.Group("/admin/image-studio", gin.HandlerFunc(admin), gin.HandlerFunc(audit), limiter.Global())
	adminRoutes.GET("/settings", func(c *gin.Context) {
		response.Success(c, gin.H{"enabled": settings.ImageStudioEnabled(c.Request.Context())})
	})
	adminRoutes.PUT("/settings", func(c *gin.Context) {
		var input struct {
			Enabled *bool `json:"enabled"`
		}
		if c.ShouldBindJSON(&input) != nil || input.Enabled == nil {
			response.BadRequest(c, "enabled is required")
			return
		}
		if err := settings.SetImageStudioEnabled(c.Request.Context(), *input.Enabled); err != nil {
			response.ErrorFrom(c, err)
			return
		}
		response.Success(c, gin.H{"enabled": *input.Enabled})
	})
}

func imageStudioRequestModel(req *http.Request) (string, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	copy := req.Clone(req.Context())
	copy.Body = io.NopCloser(bytes.NewReader(body))
	if strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data") {
		if err := copy.ParseMultipartForm(8 << 20); err != nil {
			return "", err
		}
		defer func() { _ = copy.MultipartForm.RemoveAll() }()
		if stream := copy.FormValue("stream"); stream != "" {
			value, err := strconv.ParseBool(stream)
			if err != nil || value {
				return "", io.ErrUnexpectedEOF
			}
		}
		return strings.TrimSpace(copy.FormValue("model")), nil
	}
	var data struct {
		Model  string `json:"model"`
		Stream bool   `json:"stream"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	if data.Stream {
		return "", io.ErrUnexpectedEOF
	}
	return strings.TrimSpace(data.Model), nil
}
