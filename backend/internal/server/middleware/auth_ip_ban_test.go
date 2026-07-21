package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

type authIPBanMiddlewareRepo struct {
	active      *service.AuthIPBan
	activations int
}

func (r *authIPBanMiddlewareRepo) FindActive(context.Context, string, string, time.Time) (*service.AuthIPBan, error) {
	return r.active, nil
}

func (r *authIPBanMiddlewareRepo) Activate(_ context.Context, input *service.AuthIPBanActivation) (*service.AuthIPBan, error) {
	r.activations++
	r.active = &service.AuthIPBan{
		ID:         int64(r.activations),
		IPAddress:  input.IPAddress,
		BanScope:   input.BanScope,
		UserAgent:  input.UserAgent,
		UACategory: input.UACategory,
		ExpiresAt:  input.ExpiresAt,
		Status:     "active",
	}
	return r.active, nil
}

func (r *authIPBanMiddlewareRepo) List(context.Context, *service.AuthIPBanFilter) (*service.AuthIPBanList, error) {
	return &service.AuthIPBanList{}, nil
}

func (r *authIPBanMiddlewareRepo) GetByID(context.Context, int64) (*service.AuthIPBan, error) {
	return nil, service.ErrAuthIPBanNotFound
}

func (r *authIPBanMiddlewareRepo) Release(context.Context, int64, int64, string, time.Time) (*service.AuthIPBan, error) {
	return nil, service.ErrAuthIPBanNotFound
}

func newAuthIPBanMiddlewareRouter(t *testing.T, status int) (*gin.Engine, *authIPBanMiddlewareRepo) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	repo := &authIPBanMiddlewareRepo{}
	banService := service.NewAuthIPBanService(repo, rdb)
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.POST("/api/v1/auth/login", AuthIPBan(banService), func(c *gin.Context) {
		SetAuthAttemptTarget(c, "admin@example.com")
		SetAuthAttemptFailureReason(c, "turnstile_token_missing")
		c.JSON(status, gin.H{"message": "result"})
	})
	return router, repo
}

func performAuthIPBanRequest(router *gin.Engine) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Go-http-client/1.1")
	request.RemoteAddr = "120.48.133.121:4567"
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestAuthIPBanMiddlewareBlocksAfterAutomationThreshold(t *testing.T) {
	router, repo := newAuthIPBanMiddlewareRouter(t, http.StatusBadRequest)
	for attempt := 0; attempt < 8; attempt++ {
		require.Equal(t, http.StatusBadRequest, performAuthIPBanRequest(router).Code)
	}
	require.Equal(t, 1, repo.activations)

	recorder := performAuthIPBanRequest(router)
	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	require.NotEmpty(t, recorder.Header().Get("Retry-After"))
	require.Contains(t, recorder.Body.String(), "AUTH_IP_TEMPORARILY_BANNED")
	require.Contains(t, recorder.Body.String(), "临时限制")
}

func TestAuthIPBanMiddlewareDoesNotCountServerErrors(t *testing.T) {
	router, repo := newAuthIPBanMiddlewareRouter(t, http.StatusInternalServerError)
	for attempt := 0; attempt < 12; attempt++ {
		require.Equal(t, http.StatusInternalServerError, performAuthIPBanRequest(router).Code)
	}
	require.Zero(t, repo.activations)
}
