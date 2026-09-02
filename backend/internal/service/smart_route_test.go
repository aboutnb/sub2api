package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeSmartRouteInput(t *testing.T) {
	t.Run("preset overrides caller weights", func(t *testing.T) {
		input, err := NormalizeSmartRouteInput(SmartRouteInput{
			Mode: SmartRouteModeSmart, CandidateGroupIDs: []int64{3, 8}, Strategy: SmartRouteStrategySpeed,
			Weights: &SmartRouteWeights{Price: 100},
		})
		require.NoError(t, err)
		require.Equal(t, SmartRouteWeights{Price: 15, Speed: 70, Success: 15}, *input.Weights)
		require.True(t, input.RateGuard.Enabled)
		require.Equal(t, 1.0, *input.RateGuard.MaxRateMultiplier)
	})

	t.Run("rejects duplicate candidates", func(t *testing.T) {
		_, err := NormalizeSmartRouteInput(SmartRouteInput{
			Mode: SmartRouteModeSmart, CandidateGroupIDs: []int64{3, 3}, Strategy: SmartRouteStrategyAuto,
		})
		require.Error(t, err)
	})

	t.Run("rejects invalid custom weights", func(t *testing.T) {
		_, err := NormalizeSmartRouteInput(SmartRouteInput{
			Mode: SmartRouteModeSmart, CandidateGroupIDs: []int64{3}, Strategy: SmartRouteStrategyCustom,
			Weights: &SmartRouteWeights{Price: 40, Speed: 30, Success: 29},
		})
		require.Error(t, err)
	})

	t.Run("single requires only a group", func(t *testing.T) {
		groupID := int64(4)
		input, err := NormalizeSmartRouteInput(SmartRouteInput{Mode: SmartRouteModeSingle, GroupID: &groupID})
		require.NoError(t, err)
		require.Equal(t, groupID, *input.GroupID)
	})
}

func TestScoreSmartRouteCandidatesDeterministic(t *testing.T) {
	candidates := []SmartRouteCandidate{
		{Group: &Group{ID: 9}, Position: 1, EffectiveMultiplier: 1, Metric: SmartRouteMetric{Samples: 0}},
		{Group: &Group{ID: 3}, Position: 0, EffectiveMultiplier: 1, Metric: SmartRouteMetric{Samples: 0}},
	}
	ranked := ScoreSmartRouteCandidates(candidates, SmartRouteWeights{Price: 40, Speed: 30, Success: 30})
	require.Equal(t, []int64{3, 9}, []int64{ranked[0].Group.ID, ranked[1].Group.ID})
	require.Equal(t, float64(50), ranked[0].SpeedScore)
	require.Equal(t, float64(50), ranked[0].SuccessScore)
}

func TestScoreSmartRouteCandidatesStrategies(t *testing.T) {
	candidates := []SmartRouteCandidate{
		{Group: &Group{ID: 1}, Position: 0, EffectiveMultiplier: 0.5, Metric: SmartRouteMetric{Samples: 30, SuccessRate: 0.8, LatencyP50MS: 400}},
		{Group: &Group{ID: 2}, Position: 1, EffectiveMultiplier: 1, Metric: SmartRouteMetric{Samples: 30, SuccessRate: 0.99, LatencyP50MS: 100}},
	}
	tests := []struct {
		strategy string
		wantID   int64
	}{
		{strategy: SmartRouteStrategyAuto, wantID: 2},
		{strategy: SmartRouteStrategyPrice, wantID: 1},
		{strategy: SmartRouteStrategySpeed, wantID: 2},
		{strategy: SmartRouteStrategySuccess, wantID: 2},
	}
	for _, test := range tests {
		t.Run(test.strategy, func(t *testing.T) {
			weights, ok := SmartRoutePresetWeights(test.strategy)
			require.True(t, ok)
			ranked := ScoreSmartRouteCandidates(append([]SmartRouteCandidate(nil), candidates...), weights)
			require.Equal(t, test.wantID, ranked[0].Group.ID)
		})
	}
}

func TestSmartRouteSnapshotRankingStableWithinEpoch(t *testing.T) {
	service := &SmartRouteService{rankings: make(map[string]smartRouteRankingSnapshot)}
	request := SmartRouteRequest{Model: "model-a", Kind: "text"}
	epoch := time.Now().UTC().Truncate(SmartRouteSnapshotWindow)
	weights := SmartRouteWeights{Price: 100}

	first := service.snapshotRanking(7, request, epoch, []SmartRouteCandidate{
		{Group: &Group{ID: 1}, Position: 0, EffectiveMultiplier: 0.5},
		{Group: &Group{ID: 2}, Position: 1, EffectiveMultiplier: 1},
	}, weights)
	require.Equal(t, int64(1), first[0].Group.ID)

	stable := service.snapshotRanking(7, request, epoch, []SmartRouteCandidate{
		{Group: &Group{ID: 1}, Position: 0, EffectiveMultiplier: 2},
		{Group: &Group{ID: 2}, Position: 1, EffectiveMultiplier: 0.25},
	}, weights)
	require.Equal(t, int64(1), stable[0].Group.ID)

	refreshed := service.snapshotRanking(7, request, epoch.Add(SmartRouteSnapshotWindow), []SmartRouteCandidate{
		{Group: &Group{ID: 1}, Position: 0, EffectiveMultiplier: 2},
		{Group: &Group{ID: 2}, Position: 1, EffectiveMultiplier: 0.25},
	}, weights)
	require.Equal(t, int64(2), refreshed[0].Group.ID)
}

func TestSmartRouteRequestSupported(t *testing.T) {
	require.True(t, (SmartRouteRequest{Method: "POST", Path: "/v1/responses"}).Supported())
	require.True(t, (SmartRouteRequest{Method: "POST", Path: "/v1/images/generations"}).Supported())
	require.True(t, (SmartRouteRequest{Method: "GET", Path: "/v1/models"}).Supported())
	require.False(t, (SmartRouteRequest{Method: "POST", Path: "/v1/audio/speech"}).Supported())
	require.False(t, (SmartRouteRequest{Method: "POST", Path: "/v1/videos"}).Supported())
	require.False(t, (SmartRouteRequest{Method: "POST", Path: "/v1/realtime"}).Supported())
	require.False(t, (SmartRouteRequest{Method: "POST", Path: "/v1/search"}).Supported())
	require.False(t, (SmartRouteRequest{Method: "GET", Path: "/v1/responses"}).Supported())
	require.False(t, (SmartRouteRequest{Method: "GET", Path: "/backend-api/codex/responses"}).Supported())
	require.False(t, (SmartRouteRequest{Method: "POST", Path: "/v1/web_search"}).Supported())
}

func TestSmartRouteReadOnlyImageRequestsDoNotNeedAccount(t *testing.T) {
	require.False(t, (SmartRouteRequest{Method: "GET", Path: "/v1/images/batches"}).NeedsAccount())
	require.False(t, (SmartRouteRequest{Method: "GET", Path: "/v1/images/batches/job-1"}).NeedsAccount())
	require.False(t, (SmartRouteRequest{Method: "GET", Path: "/v1/images/tasks/task-1"}).NeedsAccount())
	require.True(t, (SmartRouteRequest{Method: "POST", Path: "/v1/images/batches"}).NeedsAccount())
}

func TestSmartRouteHistoricalImageTaskRead(t *testing.T) {
	require.True(t, (SmartRouteRequest{Method: "GET", Path: "/v1/images/tasks/imgtask_1"}).IsHistoricalImageTaskRead())
	require.False(t, (SmartRouteRequest{Method: "POST", Path: "/v1/images/tasks/imgtask_1"}).IsHistoricalImageTaskRead())
	require.False(t, (SmartRouteRequest{Method: "GET", Path: "/v1/images/generations"}).IsHistoricalImageTaskRead())
}

func TestSmartRouteRequestGroupDisablesCrossGroupFallback(t *testing.T) {
	fallbackGroupID := int64(12)
	group := &Group{
		ID:                              7,
		FallbackGroupID:                 &fallbackGroupID,
		FallbackGroupIDOnInvalidRequest: &fallbackGroupID,
	}

	requestGroup := smartRouteRequestGroup(group)
	require.Nil(t, requestGroup.FallbackGroupID)
	require.Nil(t, requestGroup.FallbackGroupIDOnInvalidRequest)
	require.Equal(t, &fallbackGroupID, group.FallbackGroupID)
	require.Equal(t, &fallbackGroupID, group.FallbackGroupIDOnInvalidRequest)
}
