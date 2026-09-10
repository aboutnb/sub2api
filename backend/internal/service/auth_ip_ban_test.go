package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type authIPBanRepoStub struct {
	active      *AuthIPBan
	activations []*AuthIPBanActivation
	released    *AuthIPBan
}

type authIPBanCounterStub struct {
	mu     sync.Mutex
	counts map[string]int64
}

func (c *authIPBanCounterStub) Increment(_ context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts == nil {
		c.counts = make(map[string]int64)
	}
	c.counts[key]++
	return c.counts[key], window, nil
}

func (c *authIPBanCounterStub) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.counts, key)
	return nil
}

func (r *authIPBanRepoStub) FindActive(context.Context, string, string, time.Time) (*AuthIPBan, error) {
	return r.active, nil
}

func (r *authIPBanRepoStub) Activate(_ context.Context, input *AuthIPBanActivation) (*AuthIPBan, error) {
	copyInput := *input
	r.activations = append(r.activations, &copyInput)
	return &AuthIPBan{
		ID:               int64(len(r.activations)),
		IPAddress:        input.IPAddress,
		BanScope:         input.BanScope,
		UAHash:           input.UAHash,
		UserAgent:        input.UserAgent,
		UACategory:       input.UACategory,
		TargetIdentifier: input.TargetIdentifier,
		FailureCount:     input.FailureCount,
		ExpiresAt:        input.ExpiresAt,
		Status:           "active",
	}, nil
}

func (r *authIPBanRepoStub) List(context.Context, *AuthIPBanFilter) (*AuthIPBanList, error) {
	return &AuthIPBanList{Items: []*AuthIPBan{}, Page: 1, PageSize: 20}, nil
}

func (r *authIPBanRepoStub) GetByID(context.Context, int64) (*AuthIPBan, error) {
	return nil, ErrAuthIPBanNotFound
}

func (r *authIPBanRepoStub) Release(context.Context, int64, int64, string, time.Time) (*AuthIPBan, error) {
	if r.released == nil {
		return nil, ErrAuthIPBanNotFound
	}
	return r.released, nil
}

func newAuthIPBanTestService(t *testing.T) (*AuthIPBanService, *authIPBanRepoStub) {
	t.Helper()
	repo := &authIPBanRepoStub{}
	service := NewAuthIPBanService(repo, &authIPBanCounterStub{})
	fixedNow := time.Date(2026, 7, 22, 4, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return fixedNow }
	return service, repo
}

func TestClassifyAuthUserAgent(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
	}{
		{name: "empty", ua: "", want: AuthUserAgentEmpty},
		{name: "go client", ua: "Go-http-client/1.1", want: AuthUserAgentAutomation},
		{name: "curl", ua: "curl/8.7.1", want: AuthUserAgentAutomation},
		{name: "headless browser", ua: "Mozilla/5.0 HeadlessChrome/126.0", want: AuthUserAgentAutomation},
		{name: "chrome", ua: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/126.0 Safari/537.36", want: AuthUserAgentBrowser},
		{name: "other", ua: "MyMobileClient/1.0", want: AuthUserAgentOther},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, ClassifyAuthUserAgent(test.ua))
		})
	}
}

func TestAuthIPBanServiceAutomationThresholdBansWholeIP(t *testing.T) {
	service, repo := newAuthIPBanTestService(t)
	ctx := context.Background()
	for attempt := 1; attempt < 8; attempt++ {
		ban, err := service.RecordFailure(ctx, "120.48.133.121", "Go-http-client/1.1", "admin@sub2api.local", "/api/v1/auth/login", "turnstile_token_missing")
		require.NoError(t, err)
		require.Nil(t, ban)
	}
	ban, err := service.RecordFailure(ctx, "120.48.133.121", "Go-http-client/1.1", "admin@sub2api.local", "/api/v1/auth/login", "turnstile_token_missing")
	require.NoError(t, err)
	require.NotNil(t, ban)
	require.Len(t, repo.activations, 1)
	require.Equal(t, AuthIPBanScopeIP, repo.activations[0].BanScope)
	require.Equal(t, AuthUserAgentAutomation, repo.activations[0].UACategory)
	require.Empty(t, repo.activations[0].UAHash, "IP-wide bans must have one stable identity regardless of UA changes")
	require.Equal(t, 8, repo.activations[0].FailureCount)
	require.Equal(t, 6*time.Hour, repo.activations[0].ExpiresAt.Sub(repo.activations[0].BannedAt))
}

func TestAuthIPBanServiceBrowserCountersAreTargetScoped(t *testing.T) {
	service, repo := newAuthIPBanTestService(t)
	ctx := context.Background()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/126.0 Safari/537.36"
	for attempt := 0; attempt < 19; attempt++ {
		ban, err := service.RecordFailure(ctx, "203.0.113.9", ua, "first@example.com", "/api/v1/auth/login", "credentials_rejected")
		require.NoError(t, err)
		require.Nil(t, ban)
	}
	for attempt := 0; attempt < 19; attempt++ {
		ban, err := service.RecordFailure(ctx, "203.0.113.9", ua, "second@example.com", "/api/v1/auth/login", "credentials_rejected")
		require.NoError(t, err)
		require.Nil(t, ban)
	}
	require.Empty(t, repo.activations)

	ban, err := service.RecordFailure(ctx, "203.0.113.9", ua, "first@example.com", "/api/v1/auth/login", "credentials_rejected")
	require.NoError(t, err)
	require.NotNil(t, ban)
	require.Len(t, repo.activations, 1)
	require.Equal(t, AuthIPBanScopeIPUA, repo.activations[0].BanScope)
	require.NotEmpty(t, repo.activations[0].UAHash)
}

func TestAuthIPBanServiceSuccessfulLoginClearsFingerprintWindow(t *testing.T) {
	service, repo := newAuthIPBanTestService(t)
	ctx := context.Background()
	ua := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/126.0 Safari/537.36"
	for attempt := 0; attempt < 19; attempt++ {
		_, err := service.RecordFailure(ctx, "203.0.113.10", ua, "user@example.com", "/api/v1/auth/login", "credentials_rejected")
		require.NoError(t, err)
	}
	service.ClearFailures(ctx, "203.0.113.10", ua, "user@example.com")
	for attempt := 0; attempt < 19; attempt++ {
		ban, err := service.RecordFailure(ctx, "203.0.113.10", ua, "user@example.com", "/api/v1/auth/login", "credentials_rejected")
		require.NoError(t, err)
		require.Nil(t, ban)
	}
	require.Empty(t, repo.activations)
}

func TestAuthIPBanServiceSkipsNonPublicAddresses(t *testing.T) {
	service, repo := newAuthIPBanTestService(t)
	for attempt := 0; attempt < 20; attempt++ {
		ban, err := service.RecordFailure(context.Background(), "127.0.0.1", "Go-http-client/1.1", "admin@example.com", "/api/v1/auth/login", "credentials_rejected")
		require.NoError(t, err)
		require.Nil(t, ban)
	}
	require.Empty(t, repo.activations)
}
