package handler

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9" //nolint:depguard // Tests the registration route's Redis primitive.
	"github.com/stretchr/testify/require"
)

func TestRegistrationLimitsRejectWithoutConsumingOtherDimensions(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		ttl, err := registrationRiskLimitsScript.Run(ctx, rdb, []string{"email", "ip"}, 3, 60000, 10, 60000).Int64()
		require.NoError(t, err)
		require.Zero(t, ttl)
	}
	for i := 0; i < 5; i++ {
		ttl, err := registrationRiskLimitsScript.Run(ctx, rdb, []string{"email", "ip"}, 3, 60000, 10, 60000).Int64()
		require.NoError(t, err)
		require.Positive(t, ttl)
	}
	require.Equal(t, "3", rdb.Get(ctx, "ip").Val())
	server.FastForward(time.Minute)
	ttl, err := registrationRiskLimitsScript.Run(ctx, rdb, []string{"email", "ip"}, 3, 60000, 10, 60000).Int64()
	require.NoError(t, err)
	require.Zero(t, ttl)
}

func TestRegistrationLimitsConcurrentRequestsCannotExceedSourceBudget(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	var wg sync.WaitGroup
	results := make(chan int64, 40)
	errors := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ttl, err := registrationRiskLimitsScript.Run(context.Background(), rdb, []string{fmt.Sprintf("email:%d", i), "ip"}, 3, 60000, 10, 60000).Int64()
			results <- ttl
			errors <- err
		}(i)
	}
	wg.Wait()
	close(results)
	close(errors)
	allowed := 0
	for err := range errors {
		require.NoError(t, err)
	}
	for ttl := range results {
		if ttl == 0 {
			allowed++
		}
	}
	require.Equal(t, 10, allowed)
	require.Equal(t, "10", rdb.Get(context.Background(), "ip").Val())
}

func TestRegistrationLimitsRepairCounterMissingExpiry(t *testing.T) {
	server := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	require.NoError(t, rdb.Set(ctx, "ip", 10, 0).Err())
	ttl, err := registrationRiskLimitsScript.Run(ctx, rdb, []string{"ip"}, 10, 60000).Int64()
	require.NoError(t, err)
	require.EqualValues(t, 60000, ttl)
	require.Equal(t, time.Minute, server.TTL("ip"))
}
