package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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

type imageStudioPromptRouter interface {
	EnsureImageStudioRoute(context.Context, int64, []service.Group) error
	ImageStudioRouteGroupIDs(context.Context, int64) ([]int64, error)
}

func RegisterImageStudioRoutes(v1 *gin.RouterGroup, h *handler.Handlers, jwt middleware.JWTAuthMiddleware, admin middleware.AdminAuthMiddleware, audit middleware.AuditLogMiddleware, limiter *middleware.PanelRateLimiter, apiAuth middleware.APIKeyAuthMiddleware, resolver middleware.APIKeyGroupResolver, keys *service.APIKeyService, subscriptions *service.SubscriptionService, ops *service.OpsService, settings *service.SettingService, composite *service.CompositeRouteResolver, promptRoutes imageStudioPromptRouter, cfg *config.Config) {
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
		payload := gin.H{"enabled": settings.ImageStudioEnabled(c.Request.Context()), "storage_namespace": "image-studio-user-" + strconv.FormatInt(subject.UserID, 10), "groups": result, "async_enabled": h.AsyncImage.ImageStudioAsyncEnabled(), "max_body_bytes": cfg.Gateway.MaxBodySize}
		promptModels, promptRoute, err := imageStudioPromptOptions(c, h, keys, settings, promptRoutes, subject.UserID, groups)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		payload["prompt_models"] = promptModels
		if promptRoute != nil {
			payload["prompt_route"] = promptRoute
		}
		response.Success(c, payload)
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
	routes.POST("/groups/:group_id/prompt", middleware.RequestBodyLimit(1<<20), func(c *gin.Context) {
		optimizeImageStudioPrompt(c, h, keys, settings, promptRoutes, gateway)
	})
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

func imageStudioSelectedRouteAllows(c *gin.Context, keys *service.APIKeyService, promptRoutes imageStudioPromptRouter, userID, groupID int64) bool {
	if promptRoutes == nil || keys == nil || !keys.AllowsImageStudioPublicPromptPool(c.Request.Context(), userID) {
		return true
	}
	key, err := keys.ImageStudioPromptRouteKey(c.Request.Context(), userID)
	if err != nil || key == nil {
		return false
	}
	ids, err := promptRoutes.ImageStudioRouteGroupIDs(c.Request.Context(), key.ID)
	if err != nil {
		return false
	}
	if len(ids) == 0 {
		return true
	}
	for _, id := range ids {
		if id == groupID {
			return true
		}
	}
	return false
}

func imageStudioPromptOptions(c *gin.Context, h *handler.Handlers, keys *service.APIKeyService, settings *service.SettingService, promptRoutes imageStudioPromptRouter, userID int64, available []service.Group) ([]gin.H, gin.H, error) {
	modelsFor := func(candidate service.Group) []string { return imageStudioGroupModels(c, h, candidate) }
	choicesFor := func(source []service.Group) []gin.H {
		out := make([]gin.H, 0)
		for _, group := range source {
			for _, model := range service.ImageStudioPromptModels(modelsFor(group), 0) {
				out = append(out, gin.H{"group_id": group.ID, "group_name": group.Name, "platform": group.Platform, "model": model})
			}
		}
		return out
	}
	if promptRoutes == nil || !settings.ImageStudioEnabled(c.Request.Context()) || !keys.AllowsImageStudioPublicPromptPool(c.Request.Context(), userID) {
		return choicesFor(available), nil, nil
	}
	plan, err := keys.PlanImageStudioPromptRoute(c.Request.Context(), modelsFor)
	if err != nil || plan == nil || plan.Model == "" || len(plan.Groups) == 0 {
		return choicesFor(available), nil, err
	}
	key, err := keys.ImageStudioPromptRouteKey(c.Request.Context(), userID)
	if err != nil {
		return nil, nil, err
	}
	if err := promptRoutes.EnsureImageStudioRoute(c.Request.Context(), key.ID, plan.Groups); err != nil {
		return nil, nil, err
	}
	ids, err := promptRoutes.ImageStudioRouteGroupIDs(c.Request.Context(), key.ID)
	if err != nil {
		return nil, nil, err
	}
	source := imageStudioRouteSource(plan, ids)
	model, label := imageStudioPromptChoice(source, modelsFor)
	if model == "" {
		model, label = plan.Model, plan.Label
	}
	channels := len(source)
	if len(ids) > 0 {
		channels = len(ids)
	}
	return choicesFor(source), gin.H{"key_id": key.ID, "name": key.Name, "channels": channels, "default_model": model, "default_group": label}, nil
}

func imageStudioRouteSource(plan *service.ImageStudioPromptRoute, ids []int64) []service.Group {
	if plan == nil {
		return nil
	}
	if len(ids) == 0 {
		return plan.Groups
	}
	allowed := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		allowed[id] = struct{}{}
	}
	source := make([]service.Group, 0, len(ids))
	for _, group := range plan.Groups {
		if _, ok := allowed[group.ID]; ok {
			source = append(source, group)
		}
	}
	if len(source) == 0 {
		return plan.Groups
	}
	return source
}

func imageStudioPromptChoice(source []service.Group, modelsFor func(service.Group) []string) (string, string) {
	var chosen *service.Group
	model := ""
	for i := range source {
		group := &source[i]
		candidate := service.ImageStudioPromptModel(modelsFor(*group))
		if candidate == "" {
			continue
		}
		if chosen == nil || (group.Platform == service.PlatformOpenAI && chosen.Platform != service.PlatformOpenAI) {
			chosen = group
			model = candidate
		}
	}
	if chosen == nil {
		return "", ""
	}
	return model, chosen.Name
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

func optimizeImageStudioPrompt(c *gin.Context, h *handler.Handlers, keys *service.APIKeyService, settings *service.SettingService, promptRoutes imageStudioPromptRouter, gateway http.Handler) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if !settings.ImageStudioEnabled(c.Request.Context()) {
		response.Forbidden(c, "Image studio is disabled")
		return
	}
	groupID, err := strconv.ParseInt(c.Param("group_id"), 10, 64)
	if err != nil || groupID <= 0 {
		response.BadRequest(c, "Invalid group")
		return
	}
	groups, err := keys.GetAvailableGroups(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	allowed := false
	for i := range groups {
		if groups[i].ID == groupID && service.GroupAllowsImageGeneration(&groups[i]) {
			allowed = true
			break
		}
	}
	if !allowed {
		response.Forbidden(c, "Image generation is not allowed for this group")
		return
	}
	var input struct {
		Prompt  string `json:"prompt"`
		Model   string `json:"model"`
		GroupID int64  `json:"group_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid prompt")
		return
	}
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" || len([]rune(prompt)) > 4000 {
		response.BadRequest(c, "Invalid prompt")
		return
	}
	modelsFor := func(candidate service.Group) []string {
		return imageStudioGroupModels(c, h, candidate)
	}
	if requested := strings.TrimSpace(input.Model); requested != "" && input.GroupID > 0 {
		var selected *service.Group
		for i := range groups {
			if groups[i].ID == input.GroupID {
				selected = &groups[i]
				break
			}
		}
		allowedModel := false
		if selected != nil {
			for _, model := range service.ImageStudioPromptModels(modelsFor(*selected), 0) {
				if model == requested {
					allowedModel = true
					break
				}
			}
		}
		if selected == nil || !allowedModel || !imageStudioSelectedRouteAllows(c, keys, promptRoutes, subject.UserID, selected.ID) {
			response.Forbidden(c, "Prompt model is no longer available")
			return
		}
		key, err := keys.ImageStudioKey(c.Request.Context(), subject.UserID, selected.ID, true)
		if err != nil {
			response.ErrorFrom(c, err)
			return
		}
		optimized, ok := completeImageStudioPrompt(c, gateway, key, requested, prompt)
		if !ok {
			response.Success(c, gin.H{"prompt": prompt, "optimized": false})
			return
		}
		respondImageStudioPrompt(c, prompt, optimized, requested, selected.Name)
		return
	}
	if promptRoutes != nil && keys.AllowsImageStudioPublicPromptPool(c.Request.Context(), subject.UserID) {
		plan, planErr := keys.PlanImageStudioPromptRoute(c.Request.Context(), modelsFor)
		if planErr != nil {
			response.ErrorFrom(c, planErr)
			return
		}
		if plan != nil && plan.Model != "" && len(plan.Groups) > 0 {
			key, keyErr := keys.ImageStudioPromptRouteKey(c.Request.Context(), subject.UserID)
			if keyErr != nil {
				response.ErrorFrom(c, keyErr)
				return
			}
			if err := promptRoutes.EnsureImageStudioRoute(c.Request.Context(), key.ID, plan.Groups); err != nil {
				response.ErrorFrom(c, err)
				return
			}
			ids, idsErr := promptRoutes.ImageStudioRouteGroupIDs(c.Request.Context(), key.ID)
			if idsErr != nil {
				response.ErrorFrom(c, idsErr)
				return
			}
			model, label := imageStudioPromptChoice(imageStudioRouteSource(plan, ids), modelsFor)
			if model == "" {
				response.Success(c, gin.H{"prompt": prompt, "optimized": false})
				return
			}
			// One text call only. A failed route is not retried on another billing group.
			optimized, ok := completeImageStudioPrompt(c, gateway, key, model, prompt)
			if !ok {
				response.Success(c, gin.H{"prompt": prompt, "optimized": false})
				return
			}
			respondImageStudioPrompt(c, prompt, optimized, model, label)
			return
		}
	}
	target, model := service.BestImageStudioPromptTarget(groups, modelsFor)
	if model == "" {
		response.Success(c, gin.H{"prompt": prompt, "optimized": false})
		return
	}
	key, err := keys.ImageStudioKey(c.Request.Context(), subject.UserID, target.ID, true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	optimized, ok := completeImageStudioPrompt(c, gateway, key, model, prompt)
	if !ok {
		response.Success(c, gin.H{"prompt": prompt, "optimized": false})
		return
	}
	respondImageStudioPrompt(c, prompt, optimized, model, target.Name)
}

func imageStudioGroupModels(c *gin.Context, h *handler.Handlers, group service.Group) []string {
	if imageStudioTextModelsOverride != nil {
		return imageStudioTextModelsOverride(group)
	}
	if h != nil && h.Gateway != nil {
		return h.Gateway.ImageStudioTextModels(c.Request.Context(), group)
	}
	return nil
}

func completeImageStudioPrompt(c *gin.Context, gateway http.Handler, key *service.APIKey, model, prompt string) (string, bool) {
	body, _ := json.Marshal(map[string]any{
		"model": model, "stream": false, "max_tokens": 800,
		"messages": []map[string]string{
			{"role": "system", "content": "Rewrite the user request into one image-generation prompt in the same language. Preserve the subject, required text, style, and constraints. Add concrete composition, lighting, color, camera, and material detail only when the user did not specify them. Return only the final prompt."},
			{"role": "user", "content": prompt},
		},
	})
	req, err := http.NewRequestWithContext(service.WithImageStudioCredential(c.Request.Context(), key.Key), http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key.Key)
	recorder := httptest.NewRecorder()
	gateway.ServeHTTP(recorder, req)
	if recorder.Code < 200 || recorder.Code >= 300 {
		return "", false
	}
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if json.Unmarshal(recorder.Body.Bytes(), &result) != nil || len(result.Choices) == 0 {
		return "", false
	}
	optimized := strings.TrimSpace(result.Choices[0].Message.Content)
	if optimized == "" || len([]rune(optimized)) > 8000 {
		return prompt, true
	}
	return optimized, true
}

func respondImageStudioPrompt(c *gin.Context, original, optimized, model, groupName string) {
	payload := gin.H{"prompt": optimized, "optimized": optimized != original}
	if optimized != original {
		payload["model"] = model
		payload["group"] = groupName
	}
	response.Success(c, payload)
}

func imageStudioPromptModel(models []string) string {
	return service.ImageStudioPromptModel(models)
}

var imageStudioTextModelsOverride func(service.Group) []string
