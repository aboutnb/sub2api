package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestAuthIPBanCounterIncrementAndDelete(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	counter := NewAuthIPBanCounter(rdb)
	ctx := context.Background()
	window := 30 * time.Minute

	count, ttl, err := counter.Increment(ctx, "auth-ip-ban-test", window)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
	require.Equal(t, window, ttl)

	count, ttl, err = counter.Increment(ctx, "auth-ip-ban-test", window)
	require.NoError(t, err)
	require.Equal(t, int64(2), count)
	require.Equal(t, window, ttl)

	require.NoError(t, counter.Delete(ctx, "auth-ip-ban-test"))
	count, _, err = counter.Increment(ctx, "auth-ip-ban-test", window)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}
