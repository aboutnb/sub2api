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

func (r *studioKeysRepo) ImageStudioKey(_ context.Context, user, group int64, _ string, create bool) (*service.APIKey, error) {
	r.create = create
	r.owner = user
	return &service.APIKey{ID: 10, UserID: user, GroupID: &group, Key: "studio-internal-test", Purpose: service.ImageStudioKeyPurpose}, nil
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
			RegisterImageStudioRoutes(router.Group("/api/v1"), &handler.Handlers{}, auth, next, next, nil, gatewayAuth, nil, keys, nil, nil, settings, nil, cfg)
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
