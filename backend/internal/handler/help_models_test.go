package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestHelpModelsGroupAuthorization(t *testing.T) {
	for _, test := range []struct {
		name, id      string
		authenticated bool
		status        int
	}{
		{"unauthorized", "7", false, 401}, {"forbidden", "8", true, 404}, {"invalid", "bad", true, 400}, {"available", "7", true, 200},
	} {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/groups/"+test.id+"/help-models", nil)
			c.Params = gin.Params{{Key: "id", Value: test.id}}
			if test.authenticated {
				c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1})
			}
			authorizer := &channelMonitorV2GroupAuthorizerStub{groups: []service.Group{{ID: 7, Status: service.StatusActive, Platform: service.PlatformOpenAI}}}
			called := false
			helpModelsHandler(authorizer, func(c *gin.Context) {
				called = true
				key, ok := middleware.GetAPIKeyFromContext(c)
				require.True(t, ok)
				require.Equal(t, int64(7), key.Group.ID)
				require.Empty(t, key.Key)
				c.JSON(200, gin.H{"data": []any{}})
			})(c)
			require.Equal(t, test.status, w.Code)
			require.Equal(t, test.status == 200, called)
			require.Equal(t, "private, no-store", w.Header().Get("Cache-Control"))
		})
	}
}
