package routes

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type studioSettingsRepo struct {
	service.SettingRepository
	enabled bool
}

func (r *studioSettingsRepo) GetValue(context.Context, string) (string, error) { return "false", nil }
func (r *studioSettingsRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	value := "false"
	if r.enabled {
		value = "true"
	}
	return map[string]string{"image_studio_enabled": value}, nil
}

type studioKeysRepo struct {
	service.APIKeyRepository
	create bool
	owner  int64
}

func (r *studioKeysRepo) ImageStudioKey(_ context.Context, user, group int64, _, name string, create bool) (*service.APIKey, error) {
	r.create = create
	r.owner = user
	if name == "" {
		name = "AI 绘图"
	}
	return &service.APIKey{ID: 10, UserID: user, GroupID: &group, Key: "studio-internal-test", Name: name, Purpose: service.ImageStudioKeyPurpose}, nil
}

func (r *studioKeysRepo) ImageStudioPromptRouteKey(_ context.Context, user int64, _, name string) (*service.APIKey, error) {
	if name == "" {
		name = service.ImageStudioPromptRouteName
	}
	return &service.APIKey{ID: 10, UserID: user, Key: "studio-internal-test", Name: name, Purpose: service.ImageStudioKeyPurpose}, nil
}

func (studioGroupsRepo) Create(context.Context, *service.Group) error { return nil }
func (studioGroupsRepo) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	return nil, nil
}
func (studioGroupsRepo) BindAccountsToGroup(context.Context, int64, []int64) error { return nil }
func (studioGroupsRepo) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	return 0, nil
}

type studioUsersRepo struct{ service.UserRepository }

func (studioUsersRepo) GetByID(_ context.Context, id int64) (*service.User, error) {
	return &service.User{ID: id}, nil
}

type studioGroupsRepo struct{ service.GroupRepository }

func (studioGroupsRepo) ListActive(context.Context) ([]service.Group, error) {
	return []service.Group{{ID: 99, IsExclusive: true, Platform: service.PlatformOpenAI, AllowImageGeneration: true}}, nil
}

type studioSubscriptionsRepo struct {
	service.UserSubscriptionRepository
}

func (studioSubscriptionsRepo) ListActiveByUserID(context.Context, int64) ([]service.UserSubscription, error) {
	return nil, nil
}

func TestImageStudioRouteAuthenticationGateAndPolling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, method, path string
		login, enabled     bool
		want               int
	}{
		{"bootstrap requires identity", "GET", "/bootstrap", false, false, 401},
		{"submission requires identity", "POST", "/groups/7/images/generations", false, true, 401},
		{"disabled blocks new work", "POST", "/groups/7/images/generations", true, false, 403},
		{"exclusive group denied", "POST", "/groups/99/images/generations", true, true, 403},
		{"disabled still dispatches task reads", "GET", "/groups/7/images/tasks/imgtask_test", true, false, 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.Gateway.MaxBodySize = 1 << 20
			keyRepo := &studioKeysRepo{}
			keys := service.NewAPIKeyService(keyRepo, studioUsersRepo{}, studioGroupsRepo{}, studioSubscriptionsRepo{}, nil, nil, cfg)
			settings := service.NewSettingService(&studioSettingsRepo{enabled: tc.enabled}, cfg)
			router := gin.New()
			auth := func(c *gin.Context) {
				if tc.login {
					c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
				}
				c.Next()
			}
			next := func(c *gin.Context) { c.Next() }
			gatewayAuth := func(c *gin.Context) {
				require.Equal(t, "/v1/images/tasks/imgtask_test", c.Request.URL.Path)
				require.Equal(t, "Bearer studio-internal-test", c.GetHeader("Authorization"))
				require.Empty(t, c.GetHeader("X-API-Key"))
				require.Equal(t, int64(42), keyRepo.owner)
				require.False(t, keyRepo.create)
				c.AbortWithStatus(http.StatusNoContent)
			}
			RegisterImageStudioRoutes(router.Group("/api/v1"), &handler.Handlers{}, auth, next, next, nil, gatewayAuth, nil, keys, nil, nil, settings, nil, nil, cfg)
			req := httptest.NewRequest(tc.method, "/api/v1/image-studio"+tc.path, bytes.NewBufferString(`{"model":"gpt-image-2"}`))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "must-not-forward")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, tc.want, w.Code, w.Body.String())
		})
	}
}

func TestImageStudioRequestModelPreservesBody(t *testing.T) {
	for _, body := range []string{`{"model":"drawing","prompt":"test"}`, `{"model":"drawing","n":2}`} {
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		model, err := imageStudioRequestModel(req)
		require.NoError(t, err)
		require.Equal(t, "drawing", model)
		result, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		require.Equal(t, body, string(result))
	}
}

func TestImageStudioRequestMultipartAndStreaming(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "drawing"))
	file, err := writer.CreateFormFile("image[]", "test.png")
	require.NoError(t, err)
	_, err = file.Write([]byte("image"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	original := append([]byte(nil), body.Bytes()...)
	req := httptest.NewRequest("POST", "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	model, err := imageStudioRequestModel(req)
	require.NoError(t, err)
	require.Equal(t, "drawing", model)
	result, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Equal(t, original, result)
	req = httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"model":"gpt-image-2","stream":true}`))
	_, err = imageStudioRequestModel(req)
	require.Error(t, err)
}

type promptGroupsRepo struct{ studioGroupsRepo }

func (promptGroupsRepo) ListActive(context.Context) ([]service.Group, error) {
	return []service.Group{{ID: 7, Platform: service.PlatformOpenAI, AllowImageGeneration: true}}, nil
}

func TestImageStudioPromptOptimizationUsesTextModel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	require.Equal(t, "gpt-5.4-mini", imageStudioPromptModel([]string{"gpt-image-2", "gpt-5.4", "gpt-5.4-mini"}))
	imageStudioTextModelsOverride = func(group service.Group) []string {
		require.Equal(t, int64(7), group.ID)
		return []string{"gpt-image-2", "gpt-5.4-mini"}
	}
	t.Cleanup(func() { imageStudioTextModelsOverride = nil })
	cfg := &config.Config{}
	cfg.Gateway.MaxBodySize = 1 << 20
	keys := service.NewAPIKeyService(&studioKeysRepo{}, studioUsersRepo{}, promptGroupsRepo{}, studioSubscriptionsRepo{}, nil, nil, cfg)
	settings := service.NewSettingService(&studioSettingsRepo{enabled: true}, cfg)
	router := gin.New()
	auth := func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		c.Next()
	}
	next := func(c *gin.Context) { c.Next() }
	gatewayAuth := func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		require.Equal(t, "/v1/chat/completions", c.Request.URL.Path)
		require.Contains(t, string(body), `"model":"gpt-5.4-mini"`)
		require.Contains(t, string(body), "一只猫")
		require.NotContains(t, string(body), "gpt-image-2")
		require.Equal(t, "Bearer studio-internal-test", c.GetHeader("Authorization"))
		c.Abort()
		c.Data(http.StatusOK, "application/json", []byte(`{"choices":[{"message":{"content":"一只橘色猫咪坐在窗边，柔和晨光，浅景深"}}]}`))
	}
	RegisterImageStudioRoutes(router.Group("/api/v1"), &handler.Handlers{}, auth, next, next, nil, gatewayAuth, nil, keys, nil, nil, settings, nil, nil, cfg)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-studio/groups/7/prompt", bytes.NewBufferString(`{"prompt":"一只猫"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "一只橘色猫咪坐在窗边")
}

type promptPoolGroupRepo struct {
	service.GroupRepository
	groups  []service.Group
	sources []int64
	bound   []int64
	nextID  int64
}

func (r *promptPoolGroupRepo) ListActive(context.Context) ([]service.Group, error) {
	return append([]service.Group(nil), r.groups...), nil
}

func (r *promptPoolGroupRepo) Create(_ context.Context, group *service.Group) error {
	r.nextID++
	group.ID = 100 + r.nextID
	r.groups = append(r.groups, *group)
	return nil
}

func (r *promptPoolGroupRepo) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	return 0, nil
}

func (r *promptPoolGroupRepo) GetAccountIDsByGroupIDs(_ context.Context, groupIDs []int64) ([]int64, error) {
	r.sources = append([]int64(nil), groupIDs...)
	return []int64{11, 12}, nil
}

func (r *promptPoolGroupRepo) BindAccountsToGroup(_ context.Context, _ int64, accountIDs []int64) error {
	r.bound = append([]int64(nil), accountIDs...)
	return nil
}

type promptPoolKeyRepo struct {
	studioKeysRepo
	groupID int64
	name    string
	calls   int
}

func (r *promptPoolKeyRepo) ImageStudioKey(ctx context.Context, user, group int64, credential, name string, create bool) (*service.APIKey, error) {
	r.calls++
	r.groupID = group
	r.name = name
	return r.studioKeysRepo.ImageStudioKey(ctx, user, group, credential, name, create)
}

func (r *promptPoolKeyRepo) ImageStudioPromptRouteKey(ctx context.Context, user int64, credential, name string) (*service.APIKey, error) {
	r.calls++
	r.groupID = 0
	r.name = name
	return r.studioKeysRepo.ImageStudioPromptRouteKey(ctx, user, credential, name)
}

type promptRouteSaver struct {
	ids   []int64
	calls int
}

func (s *promptRouteSaver) EnsureImageStudioRoute(ctx context.Context, id int64, groups []service.Group) error {
	return s.SaveImageStudioRoute(ctx, id, groups)
}

func (s *promptRouteSaver) ImageStudioRouteGroupIDs(context.Context, int64) ([]int64, error) {
	return append([]int64(nil), s.ids...), nil
}

func (s *promptRouteSaver) SaveImageStudioRoute(_ context.Context, _ int64, groups []service.Group) error {
	s.calls++
	s.ids = s.ids[:0]
	for _, group := range groups {
		s.ids = append(s.ids, group.ID)
	}
	return nil
}

type restrictedPromptUsers struct{ studioUsersRepo }

func (restrictedPromptUsers) GetByID(_ context.Context, id int64) (*service.User, error) {
	return &service.User{ID: id, RestrictPublicGroups: true, AllowedGroups: []int64{7, 8}}, nil
}

func TestImageStudioPromptOptimizationCreatesSmartGroupAcrossPublicChannels(t *testing.T) {
	gin.SetMode(gin.TestMode)
	imageStudioTextModelsOverride = func(group service.Group) []string {
		switch group.ID {
		case 3:
			return []string{"gpt-image-2", "gpt-4o-mini"}
		case 9:
			return []string{"claude-3-5-haiku"}
		default:
			return []string{"gpt-image-2"}
		}
	}
	t.Cleanup(func() { imageStudioTextModelsOverride = nil })
	repo := &promptPoolGroupRepo{groups: []service.Group{
		{ID: 7, Name: "绘图", Status: service.StatusActive, Platform: service.PlatformOpenAI, SubscriptionType: service.SubscriptionTypeStandard, AllowImageGeneration: true, RateMultiplier: 1},
		{ID: 3, Name: "公开文本", Status: service.StatusActive, Platform: service.PlatformOpenAI, SubscriptionType: service.SubscriptionTypeStandard, RateMultiplier: 1},
		{ID: 9, Name: "专属便宜", Status: service.StatusActive, Platform: service.PlatformAnthropic, SubscriptionType: service.SubscriptionTypeStandard, IsExclusive: true, RateMultiplier: 0.1},
	}}
	cfg := &config.Config{}
	cfg.Gateway.MaxBodySize = 1 << 20
	keysRepo := &promptPoolKeyRepo{}
	keys := service.NewAPIKeyService(keysRepo, studioUsersRepo{}, repo, studioSubscriptionsRepo{}, nil, nil, cfg)
	settings := service.NewSettingService(&studioSettingsRepo{enabled: true}, cfg)
	routes := &promptRouteSaver{}
	router := gin.New()
	auth := func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		c.Next()
	}
	next := func(c *gin.Context) { c.Next() }
	gatewayAuth := func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), `"model":"gpt-4o-mini"`)
		require.NotContains(t, string(body), "claude-3-5-haiku")
		c.Abort()
		c.Data(http.StatusOK, "application/json", []byte(`{"choices":[{"message":{"content":"窗边的橘猫，晨光，浅景深"}}]}`))
	}
	RegisterImageStudioRoutes(router.Group("/api/v1"), &handler.Handlers{}, auth, next, next, nil, gatewayAuth, nil, keys, nil, nil, settings, nil, routes, cfg)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-studio/groups/7/prompt", bytes.NewBufferString(`{"prompt":"一只猫"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "窗边的橘猫")
	require.Contains(t, w.Body.String(), "公开文本")
	require.NotContains(t, w.Body.String(), service.ImageStudioSmartGroupName)
	require.Empty(t, repo.sources)
	require.Empty(t, repo.bound)
	require.Equal(t, int64(0), repo.nextID)
	require.Equal(t, []int64{3}, routes.ids)
	require.Equal(t, service.ImageStudioPromptRouteName, keysRepo.name)
	require.Equal(t, int64(0), keysRepo.groupID)
}

func TestImageStudioPromptOptimizationRestrictedUserStaysOnGrantedGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	imageStudioTextModelsOverride = func(group service.Group) []string {
		switch group.ID {
		case 8:
			return []string{"gpt-4o-mini"}
		case 3:
			return []string{"claude-3-5-haiku"}
		default:
			return []string{"gpt-image-2"}
		}
	}
	t.Cleanup(func() { imageStudioTextModelsOverride = nil })
	repo := &promptPoolGroupRepo{groups: []service.Group{
		{ID: 7, Name: "绘图", Status: service.StatusActive, Platform: service.PlatformOpenAI, SubscriptionType: service.SubscriptionTypeStandard, AllowImageGeneration: true, RateMultiplier: 1},
		{ID: 8, Name: "已授权", Status: service.StatusActive, Platform: service.PlatformOpenAI, SubscriptionType: service.SubscriptionTypeStandard, RateMultiplier: 1.2},
		{ID: 3, Name: "未授权公开", Status: service.StatusActive, Platform: service.PlatformAnthropic, SubscriptionType: service.SubscriptionTypeStandard, RateMultiplier: 0.1},
	}}
	cfg := &config.Config{}
	cfg.Gateway.MaxBodySize = 1 << 20
	keysRepo := &promptPoolKeyRepo{}
	keys := service.NewAPIKeyService(keysRepo, restrictedPromptUsers{}, repo, studioSubscriptionsRepo{}, nil, nil, cfg)
	settings := service.NewSettingService(&studioSettingsRepo{enabled: true}, cfg)
	router := gin.New()
	auth := func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		c.Next()
	}
	next := func(c *gin.Context) { c.Next() }
	gatewayAuth := func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), `"model":"gpt-4o-mini"`)
		require.NotContains(t, string(body), "haiku")
		c.Abort()
		c.Data(http.StatusOK, "application/json", []byte(`{"choices":[{"message":{"content":"授权渠道改写"}}]}`))
	}
	routes := &promptRouteSaver{}
	RegisterImageStudioRoutes(router.Group("/api/v1"), &handler.Handlers{}, auth, next, next, nil, gatewayAuth, nil, keys, nil, nil, settings, nil, routes, cfg)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-studio/groups/7/prompt", bytes.NewBufferString(`{"prompt":"一只猫"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), "授权渠道改写")
	require.Equal(t, 0, routes.calls)
	require.Equal(t, int64(8), keysRepo.groupID)
	require.Equal(t, "AI 绘图", keysRepo.name)
	require.NotEqual(t, service.ImageStudioSmartGroupName, keysRepo.name)
}

func TestImageStudioPromptPoolFailureDoesNotRetryAnotherGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	imageStudioTextModelsOverride = func(group service.Group) []string {
		if group.ID == 3 {
			return []string{"gpt-4o-mini"}
		}
		return []string{"gpt-5.4-mini"}
	}
	t.Cleanup(func() { imageStudioTextModelsOverride = nil })
	repo := &promptPoolGroupRepo{groups: []service.Group{
		{ID: 7, Name: "绘图", Status: service.StatusActive, Platform: service.PlatformOpenAI, SubscriptionType: service.SubscriptionTypeStandard, AllowImageGeneration: true},
		{ID: 3, Name: "公开文本", Status: service.StatusActive, Platform: service.PlatformOpenAI, SubscriptionType: service.SubscriptionTypeStandard},
	}}
	cfg := &config.Config{}
	cfg.Gateway.MaxBodySize = 1 << 20
	keysRepo := &promptPoolKeyRepo{}
	keys := service.NewAPIKeyService(keysRepo, studioUsersRepo{}, repo, studioSubscriptionsRepo{}, nil, nil, cfg)
	settings := service.NewSettingService(&studioSettingsRepo{enabled: true}, cfg)
	router := gin.New()
	auth := func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 42})
		c.Next()
	}
	next := func(c *gin.Context) { c.Next() }
	gatewayAuth := func(c *gin.Context) { c.Status(http.StatusBadGateway) }
	RegisterImageStudioRoutes(router.Group("/api/v1"), &handler.Handlers{}, auth, next, next, nil, gatewayAuth, nil, keys, nil, nil, settings, nil, &promptRouteSaver{}, cfg)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/image-studio/groups/7/prompt", bytes.NewBufferString(`{"prompt":"一只猫"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.Contains(t, w.Body.String(), `"optimized":false`)
	require.Equal(t, 1, keysRepo.calls)
}
