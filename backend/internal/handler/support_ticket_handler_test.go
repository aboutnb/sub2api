package handler

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

func supportTicketRequest(t *testing.T, fn gin.HandlerFunc, userID int64, method, url, body, id string) (int, map[string]any) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, url, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: id}}
	if userID > 0 {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
	}
	fn(c)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	return recorder.Code, payload
}

func TestSupportTicketHandlerUserWorkflow(t *testing.T) {
	r := tickettest.New()
	r.AddUser(1, "user@example.test")
	r.AddUser(2, "other@example.test")
	h := NewSupportTicketHandler(service.NewSupportTicketService(r))
	code, payload := supportTicketRequest(t, h.Create, 1, "POST", "/tickets", `{"subject":"Help","body":"Message","category":"billing"}`, "")
	require.Equal(t, 200, code)
	require.EqualValues(t, 0, payload["code"])
	ticket, ok := payload["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "open", ticket["status"])
	require.Contains(t, ticket, "closed_at")
	require.Nil(t, ticket["closed_at"])
	require.Len(t, ticket["messages"], 1)
	code, payload = supportTicketRequest(t, h.List, 1, "GET", "/tickets?page=1&page_size=5&category=all", "", "")
	require.Equal(t, 200, code)
	list, ok := payload["data"].(map[string]any)
	require.True(t, ok)
	require.EqualValues(t, 1, list["total"])
	items, ok := list["items"].([]any)
	require.True(t, ok)
	require.NotContains(t, items[0], "messages")
	code, payload = supportTicketRequest(t, h.Reply, 1, "POST", "/tickets/1/messages", `{"body":"More info"}`, "1")
	require.Equal(t, 200, code)
	reply, ok := payload["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "pending_admin", reply["status"])
	code, _ = supportTicketRequest(t, h.Close, 1, "POST", "/tickets/1/close", "", "1")
	require.Equal(t, 200, code)
	code, payload = supportTicketRequest(t, h.Reply, 1, "POST", "/tickets/1/messages", `{"body":"More"}`, "1")
	require.Equal(t, 409, code)
	require.Equal(t, "TICKET_CLOSED", payload["reason"])
	for _, fn := range []gin.HandlerFunc{h.Get, h.Reply, h.Close} {
		code, payload = supportTicketRequest(t, fn, 2, "POST", "/tickets/1", `{"body":"More"}`, "1")
		require.Equal(t, 404, code)
		require.Equal(t, "TICKET_NOT_FOUND", payload["reason"])
	}
}

func TestSupportTicketHandlerValidationAndMetadata(t *testing.T) {
	r := tickettest.New()
	r.AddUser(1, "user@example.test")
	s := service.NewSupportTicketService(r)
	h := NewSupportTicketHandler(s)
	code, _ := supportTicketRequest(t, h.List, 0, "GET", "/tickets", "", "")
	require.Equal(t, 401, code)
	for _, tc := range []struct {
		fn                    gin.HandlerFunc
		body, url, id, reason string
	}{
		{h.Create, `{}`, "/tickets", "", "TICKET_INPUT_REQUIRED"},
		{h.Create, `{"subject":"Help","body":"Body","category":"all"}`, "/tickets", "", "TICKET_CATEGORY_INVALID"},
		{h.Reply, `{"body":" "}`, "/tickets/1/messages", "1", "TICKET_INPUT_REQUIRED"},
		{h.List, "", "/tickets?status=all", "", "TICKET_STATUS_INVALID"},
		{h.List, "", "/tickets?category=BAD", "", "TICKET_CATEGORY_INVALID"},
	} {
		code, payload := supportTicketRequest(t, tc.fn, 1, "POST", tc.url, tc.body, tc.id)
		require.Equal(t, 400, code)
		require.Equal(t, tc.reason, payload["reason"])
	}
	code, _ = supportTicketRequest(t, h.Get, 1, "GET", "/tickets/bad", "", "bad")
	require.Equal(t, 400, code)
	code, _ = supportTicketRequest(t, h.Create, 1, "POST", "/tickets", "{", "")
	require.Equal(t, 400, code)
	for i := 0; i < 10; i++ {
		_, err := s.Create(context.Background(), 1, service.CreateTicketInput{Subject: "Help", Body: "Body", Category: "other"})
		require.NoError(t, err)
	}
	code, payload := supportTicketRequest(t, h.Create, 1, "POST", "/tickets", `{"subject":"Help","body":"Body","category":"other"}`, "")
	require.Equal(t, 409, code)
	require.Equal(t, "TICKET_OPEN_LIMIT", payload["reason"])
	require.Equal(t, "10", payload["metadata"].(map[string]any)["max"])
}
