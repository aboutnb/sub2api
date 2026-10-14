package middleware

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestTicketWriteLimitsIndependentOfGlobalAndIP(t *testing.T) {
	fake := &fakePanelAllower{}
	p := &PanelRateLimiter{limiter: fake}
	a := newPanelTestRouter(p.TicketWrite(true), &panelTestIdentity{userID: 1, role: service.RoleAdmin})
	b := newPanelTestRouter(p.TicketWrite(true), &panelTestIdentity{userID: 2, role: service.RoleUser})
	for i := 0; i < 5; i++ {
		require.Equal(t, 200, performPanelRequest(a, "127.0.0.1:1").Code)
	}
	limited := performPanelRequest(a, "8.8.8.8:1")
	require.Equal(t, http.StatusTooManyRequests, limited.Code)
	require.NotEmpty(t, limited.Header().Get("Retry-After"))
	require.Equal(t, 200, performPanelRequest(b, "127.0.0.1:1").Code)
	// The hour bucket must still reject when the minute bucket has expired.
	fake.mu.Lock()
	fake.counts["ticket:create:minute:user:1"] = 0
	fake.counts["ticket:create:hour:user:1"] = 20
	fake.mu.Unlock()
	require.Equal(t, 429, performPanelRequest(a, "127.0.0.1:1").Code)
}

func TestTicketWriteFailsClosed(t *testing.T) {
	p := &PanelRateLimiter{limiter: &fakePanelAllower{err: errors.New("redis offline")}}
	router := newPanelTestRouter(p.TicketWrite(false), &panelTestIdentity{userID: 1, role: service.RoleUser})
	require.Equal(t, 503, performPanelRequest(router, "127.0.0.1:1").Code)
	router = newPanelTestRouter(p.TicketWrite(false), nil)
	require.Equal(t, 401, performPanelRequest(router, "127.0.0.1:1").Code)
}
