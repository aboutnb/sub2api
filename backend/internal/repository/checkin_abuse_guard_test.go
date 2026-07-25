package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestCheckinAbuseGuardDeduplicatesUsersAndLimitsNewUsers(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	guard := newCheckinAbuseGuardForTest(client)
	ctx := context.Background()
	window := 10 * time.Minute

	allowed, count, ttl, err := guard.CheckAndRecord(ctx, "192.0.2.10", 101, window, 2)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, int64(1), count)
	require.Greater(t, ttl, time.Duration(0))

	allowed, count, _, err = guard.CheckAndRecord(ctx, "192.0.2.10", 101, window, 2)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, int64(1), count)

	allowed, count, _, err = guard.CheckAndRecord(ctx, "192.0.2.10", 102, window, 2)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, int64(2), count)

	allowed, count, _, err = guard.CheckAndRecord(ctx, "192.0.2.10", 103, window, 2)
	require.NoError(t, err)
	require.False(t, allowed)
	require.Equal(t, int64(3), count)

	keys := mr.Keys()
	require.Len(t, keys, 1)
	require.True(t, strings.HasPrefix(keys[0], "checkin:risk:source:"))
	require.NotContains(t, keys[0], "192.0.2.10")
	members, err := mr.SMembers(keys[0])
	require.NoError(t, err)
	require.Len(t, members, 2)
	require.Greater(t, mr.TTL(keys[0]), time.Duration(0))
}

func TestCheckinAbuseGuardExpiresWindow(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	guard := newCheckinAbuseGuardForTest(client)

	allowed, _, _, err := guard.CheckAndRecord(context.Background(), "192.0.2.10", 101, time.Minute, 1)
	require.NoError(t, err)
	require.True(t, allowed)
	mr.FastForward(time.Minute + time.Second)

	allowed, count, _, err := guard.CheckAndRecord(context.Background(), "192.0.2.10", 102, time.Minute, 1)
	require.NoError(t, err)
	require.True(t, allowed)
	require.Equal(t, int64(1), count)
}

func TestCheckinAbuseGuardReturnsRedisError(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	guard := newCheckinAbuseGuardForTest(client)
	require.NoError(t, client.Close())

	_, _, _, err := guard.CheckAndRecord(context.Background(), "192.0.2.10", 101, time.Minute, 1)
	require.Error(t, err)
}

func TestCheckinAbuseGuardFailsClosedWithoutHashKey(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	guard := NewCheckinAbuseGuard(client, &config.Config{})

	_, _, err := guard.CheckRequest(context.Background(), "192.0.2.10", 101, time.Minute, 1, 1)
	require.Error(t, err)
}

func TestCheckinRequestGuardLimitsUserWithoutPoisoningSource(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	guard := newCheckinAbuseGuardForTest(client)
	ctx := context.Background()

	allowed, _, err := guard.CheckRequest(ctx, "192.0.2.10", 101, time.Minute, 1, 2)
	require.NoError(t, err)
	require.True(t, allowed)
	allowed, retryAfter, err := guard.CheckRequest(ctx, "192.0.2.10", 101, time.Minute, 1, 2)
	require.NoError(t, err)
	require.False(t, allowed)
	require.Greater(t, retryAfter, time.Duration(0))

	allowed, _, err = guard.CheckRequest(ctx, "192.0.2.10", 102, time.Minute, 1, 2)
	require.NoError(t, err)
	require.True(t, allowed, "the rejected user request must not consume the shared source quota")
}

func TestCheckinRequestGuardLimitsSourceAndHashesAddress(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	guard := newCheckinAbuseGuardForTest(client)
	ctx := context.Background()

	for userID := int64(101); userID <= 102; userID++ {
		allowed, _, err := guard.CheckRequest(ctx, "192.0.2.10", userID, time.Minute, 5, 2)
		require.NoError(t, err)
		require.True(t, allowed)
	}
	allowed, _, err := guard.CheckRequest(ctx, "192.0.2.10", 103, time.Minute, 5, 2)
	require.NoError(t, err)
	require.False(t, allowed)

	for _, key := range mr.Keys() {
		require.NotContains(t, key, "192.0.2.10")
		require.NotContains(t, key, "101")
		require.NotContains(t, key, "102")
		require.NotContains(t, key, "103")
		require.Greater(t, mr.TTL(key), time.Duration(0))
	}
}

func newCheckinAbuseGuardForTest(client *redis.Client) service.CheckinAbuseGuard {
	return NewCheckinAbuseGuard(client, &config.Config{JWT: config.JWTConfig{
		Secret: "0123456789abcdef0123456789abcdef",
	}})
}
