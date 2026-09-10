package service

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"sync"
	"time"
)

type httpUpstreamTraceContextKey struct{}

// HTTPUpstreamTrace records low-level timings from net/http without logging per request.
type HTTPUpstreamTrace struct {
	mu sync.Mutex

	startedAt time.Time

	getConnAt             time.Time
	dnsStartAt            time.Time
	connectStartAt        time.Time
	tlsHandshakeStartAt   time.Time
	wroteRequestAt        time.Time
	gotFirstResponseAt    time.Time
	connWaitDuration      time.Duration
	dnsDuration           time.Duration
	connectDuration       time.Duration
	tlsHandshakeDuration  time.Duration
	requestWriteDuration  time.Duration
	headerWaitDuration    time.Duration
	firstResponseDuration time.Duration
	connIdleDuration      time.Duration

	gotConn          bool
	hasDNS           bool
	hasConnect       bool
	hasTLSHandshake  bool
	wroteRequest     bool
	gotFirstResponse bool
	connReused       bool
	connWasIdle      bool
}

type HTTPUpstreamTraceSnapshot struct {
	ConnWaitDuration      time.Duration
	DNSDuration           time.Duration
	ConnectDuration       time.Duration
	TLSHandshakeDuration  time.Duration
	RequestWriteDuration  time.Duration
	HeaderWaitDuration    time.Duration
	FirstResponseDuration time.Duration
	ConnIdleDuration      time.Duration

	GotConn          bool
	HasDNS           bool
	HasConnect       bool
	HasTLSHandshake  bool
	WroteRequest     bool
	GotFirstResponse bool
	ConnReused       bool
	ConnWasIdle      bool
}

func NewHTTPUpstreamTrace(startedAt time.Time) *HTTPUpstreamTrace {
	if startedAt.IsZero() {
		startedAt = time.Now()
	}
	return &HTTPUpstreamTrace{startedAt: startedAt}
}

func WithHTTPUpstreamTrace(ctx context.Context, trace *HTTPUpstreamTrace) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if trace == nil {
		return ctx
	}
	return context.WithValue(ctx, httpUpstreamTraceContextKey{}, trace)
}

func HTTPUpstreamTraceFromContext(ctx context.Context) (*HTTPUpstreamTrace, bool) {
	if ctx == nil {
		return nil, false
	}
	trace, ok := ctx.Value(httpUpstreamTraceContextKey{}).(*HTTPUpstreamTrace)
	if !ok || trace == nil {
		return nil, false
	}
	return trace, true
}

func (t *HTTPUpstreamTrace) ClientTrace() *httptrace.ClientTrace {
	if t == nil {
		return nil
	}
	return &httptrace.ClientTrace{
		GetConn: func(string) {
			t.markGetConn(time.Now())
		},
		GotConn: func(info httptrace.GotConnInfo) {
			t.markGotConn(time.Now(), info)
		},
		DNSStart: func(httptrace.DNSStartInfo) {
			t.markDNSStart(time.Now())
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			t.markDNSDone(time.Now())
		},
		ConnectStart: func(_, _ string) {
			t.markConnectStart(time.Now())
		},
		ConnectDone: func(_, _ string, _ error) {
			t.markConnectDone(time.Now())
		},
		TLSHandshakeStart: func() {
			t.markTLSHandshakeStart(time.Now())
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			t.markTLSHandshakeDone(time.Now())
		},
		WroteRequest: func(httptrace.WroteRequestInfo) {
			t.markWroteRequest(time.Now())
		},
		GotFirstResponseByte: func() {
			t.markGotFirstResponseByte(time.Now())
		},
	}
}

func (t *HTTPUpstreamTrace) Snapshot() HTTPUpstreamTraceSnapshot {
	if t == nil {
		return HTTPUpstreamTraceSnapshot{}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	return HTTPUpstreamTraceSnapshot{
		ConnWaitDuration:      t.connWaitDuration,
		DNSDuration:           t.dnsDuration,
		ConnectDuration:       t.connectDuration,
		TLSHandshakeDuration:  t.tlsHandshakeDuration,
		RequestWriteDuration:  t.requestWriteDuration,
		HeaderWaitDuration:    t.headerWaitDuration,
		FirstResponseDuration: t.firstResponseDuration,
		ConnIdleDuration:      t.connIdleDuration,
		GotConn:               t.gotConn,
		HasDNS:                t.hasDNS,
		HasConnect:            t.hasConnect,
		HasTLSHandshake:       t.hasTLSHandshake,
		WroteRequest:          t.wroteRequest,
		GotFirstResponse:      t.gotFirstResponse,
		ConnReused:            t.connReused,
		ConnWasIdle:           t.connWasIdle,
	}
}

func (t *HTTPUpstreamTrace) markGetConn(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.startedAt.IsZero() {
		t.startedAt = now
	}
	t.getConnAt = now
}

func (t *HTTPUpstreamTrace) markGotConn(now time.Time, info httptrace.GotConnInfo) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.gotConn = true
	t.connReused = info.Reused
	t.connWasIdle = info.WasIdle
	if info.WasIdle {
		t.connIdleDuration = info.IdleTime
	}
	if !t.getConnAt.IsZero() {
		t.connWaitDuration = nonNegativeDuration(now.Sub(t.getConnAt))
	}
}

func (t *HTTPUpstreamTrace) markDNSStart(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dnsStartAt = now
}

func (t *HTTPUpstreamTrace) markDNSDone(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.hasDNS = true
	if !t.dnsStartAt.IsZero() {
		t.dnsDuration = nonNegativeDuration(now.Sub(t.dnsStartAt))
	}
}

func (t *HTTPUpstreamTrace) markConnectStart(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.connectStartAt = now
}

func (t *HTTPUpstreamTrace) markConnectDone(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.hasConnect = true
	if !t.connectStartAt.IsZero() {
		t.connectDuration = nonNegativeDuration(now.Sub(t.connectStartAt))
	}
}

func (t *HTTPUpstreamTrace) markTLSHandshakeStart(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tlsHandshakeStartAt = now
}

func (t *HTTPUpstreamTrace) markTLSHandshakeDone(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.hasTLSHandshake = true
	if !t.tlsHandshakeStartAt.IsZero() {
		t.tlsHandshakeDuration = nonNegativeDuration(now.Sub(t.tlsHandshakeStartAt))
	}
}

func (t *HTTPUpstreamTrace) markWroteRequest(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.wroteRequest = true
	t.wroteRequestAt = now
	if !t.startedAt.IsZero() {
		t.requestWriteDuration = nonNegativeDuration(now.Sub(t.startedAt))
	}
}

func (t *HTTPUpstreamTrace) markGotFirstResponseByte(now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.gotFirstResponse = true
	t.gotFirstResponseAt = now
	if !t.startedAt.IsZero() {
		t.firstResponseDuration = nonNegativeDuration(now.Sub(t.startedAt))
	}
	if !t.wroteRequestAt.IsZero() {
		t.headerWaitDuration = nonNegativeDuration(now.Sub(t.wroteRequestAt))
	}
}

func nonNegativeDuration(d time.Duration) time.Duration {
	if d < 0 {
		return 0
	}
	return d
}
