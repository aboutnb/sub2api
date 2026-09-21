package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestImageStudioCompositeAsyncAndTaskOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &asyncImageMemoryStore{tasks: map[string]*service.ImageTaskRecord{}}
	tasks := service.NewImageTaskServiceWithUploader(store, nil, time.Hour, time.Minute)
	executed := make(chan string, 1)
	h := &AsyncImageHandler{tasks: tasks, execute: func(platform string, c *gin.Context) {
		executed <- platform
		c.JSON(200, gin.H{"data": []gin.H{{"url": "https://example.test/image.png"}}})
	}}
	groupID := int64(7)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 9, UserID: 42, GroupID: &groupID, Group: &service.Group{ID: groupID, Platform: service.PlatformComposite, AllowImageGeneration: true}})
		c.Request = c.Request.WithContext(service.WithCompositeRouteDecision(c.Request.Context(), service.CompositeRouteDecision{Matched: true, TargetPlatform: service.PlatformOpenAI, UpstreamModel: "gpt-image-2"}))
		c.Next()
	})
	router.POST("/v1/images/generations/async", h.Submit)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v1/images/generations/async", strings.NewReader(`{"model":"gpt-image-2","prompt":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code, w.Body.String())
	select {
	case platform := <-executed:
		require.Equal(t, service.PlatformOpenAI, platform)
	case <-time.After(time.Second):
		t.Fatal("gateway did not run")
	}

	task, err := tasks.Create(context.Background(), service.ImageTaskOwner{UserID: 42, APIKeyID: 9, GroupID: 7})
	require.NoError(t, err)
	for _, owner := range []service.ImageTaskOwner{{UserID: 43, APIKeyID: 9}, {UserID: 42, APIKeyID: 10}} {
		_, err := tasks.Get(context.Background(), owner, task.TaskID)
		require.ErrorIs(t, err, service.ErrImageTaskNotFound)
	}
}
