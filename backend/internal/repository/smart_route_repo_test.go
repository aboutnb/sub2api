package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestSmartRouteRepositoryPositiveAndNegativeCache(t *testing.T) {
	ctx := context.Background()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := &smartRouteRepository{rdb: rdb, l1: make(map[int64]smartRouteCacheEntry)}

	config := &service.SmartRouteConfig{
		APIKeyID: 4, Mode: service.SmartRouteModeSmart, Strategy: service.SmartRouteStrategyAuto,
		CandidateGroupIDs: []int64{8, 9}, Weights: service.SmartRouteWeights{Price: 40, Speed: 30, Success: 30},
	}
	repo.setCached(ctx, 4, config)
	config.CandidateGroupIDs[0] = 99
	loaded, err := repo.GetMany(ctx, []int64{4})
	require.NoError(t, err)
	require.Equal(t, []int64{8, 9}, loaded[4].CandidateGroupIDs)

	repo.setCached(ctx, 5, nil)
	missing, err := repo.GetMany(ctx, []int64{5})
	require.NoError(t, err)
	require.Empty(t, missing)

	repo.Invalidate(ctx, 4)
	require.False(t, mr.Exists(smartRouteCacheKey(4)))
}
