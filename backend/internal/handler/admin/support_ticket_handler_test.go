package admin

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/Wei-Shaw/sub2api/internal/service/tickettest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func adminSupportTicketRequest(t *testing.T, fn gin.HandlerFunc, method, url, body string) (int, map[string]any) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, url, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 9})
	fn(c)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	return recorder.Code, payload
}

func TestSupportTicketAdminHandlerQueuesAndReopen(t *testing.T) {
	r := tickettest.New()
	r.AddUser(1, "owner@example.test")
	r.AddUser(9, "admin@example.test")
	s := service.NewSupportTicketService(r)
	_, err := s.Create(context.Background(), 1, service.CreateTicketInput{Subject: "Help", Body: "Body", Category: "billing"})
	require.NoError(t, err)
	h := NewSupportTicketHandler(s)
	code, payload := adminSupportTicketRequest(t, h.List, "GET", "/admin/tickets?queue=unanswered&search=OWNER", "")
	require.Equal(t, 200, code)
	require.EqualValues(t, 0, payload["code"])
	list, ok := payload["data"].(map[string]any)
	require.True(t, ok)
	require.EqualValues(t, 1, list["total"])
	items, ok := list["items"].([]any)
	require.True(t, ok)
	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "owner@example.test", item["user_email"])
	for _, queue := range []string{"all", "open"} {
		code, payload = adminSupportTicketRequest(t, h.List, "GET", "/admin/tickets?queue="+queue, "")
		require.Equal(t, 400, code)
		require.Equal(t, "TICKET_STATUS_INVALID", payload["reason"])
	}
	code, payload = adminSupportTicketRequest(t, h.UpdateStatus, "PUT", "/admin/tickets/1/status", `{"status":"closed"}`)
	require.Equal(t, 200, code)
	closed, ok := payload["data"].(map[string]any)
	require.True(t, ok)
	require.NotNil(t, closed["closed_at"])
	code, payload = adminSupportTicketRequest(t, h.Reply, "POST", "/admin/tickets/1/messages", `{"body":"Resolved"}`)
	require.Equal(t, 200, code)
	ticket := payload["data"].(map[string]any)
	require.Equal(t, "pending_user", ticket["status"])
	require.Nil(t, ticket["closed_at"])
	require.Len(t, ticket["messages"], 2)
	code, payload = adminSupportTicketRequest(t, h.Get, "GET", "/admin/tickets/1", "")
	require.Equal(t, 200, code)
	require.Equal(t, "owner@example.test", payload["data"].(map[string]any)["user_email"])
	code, payload = adminSupportTicketRequest(t, h.UpdateStatus, "PUT", "/admin/tickets/1/status", `{"status":"all"}`)
	require.Equal(t, 400, code)
	require.Equal(t, "TICKET_STATUS_INVALID", payload["reason"])
}
