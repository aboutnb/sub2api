//go:build integration

package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestImageStudioKeyConcurrentCreation(t *testing.T) {
	client := testEntClient(t)
	ctx := context.Background()
	user, err := client.User.Create().SetEmail(fmt.Sprintf("studio-%d@test.com", time.Now().UnixNano())).SetPasswordHash("hash").Save(ctx)
	require.NoError(t, err)
	group, err := client.Group.Create().SetName(fmt.Sprintf("studio-%d", user.ID)).Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = client.APIKey.Delete().Where(apikey.UserIDEQ(user.ID)).Exec(ctx)
		_ = client.Group.DeleteOneID(group.ID).Exec(ctx)
		_ = client.User.DeleteOneID(user.ID).Exec(ctx)
	})
	repo := newAPIKeyRepositoryWithSQL(client, integrationDB)
	var wg sync.WaitGroup
	results := make(chan *service.APIKey, 12)
	failures := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key, err := repo.ImageStudioKey(ctx, user.ID, group.ID, fmt.Sprintf("studio-internal-%d-%d", user.ID, i), "AI 绘图", true)
			results <- key
			failures <- err
		}(i)
	}
	wg.Wait()
	close(results)
	close(failures)
	for err := range failures {
		require.NoError(t, err)
	}
	var id int64
	for key := range results {
		if id == 0 {
			id = key.ID
		}
		require.Equal(t, id, key.ID)
	}
	count, err := client.APIKey.Query().Where(apikey.UserIDEQ(user.ID), apikey.PurposeEQ(service.ImageStudioKeyPurpose)).Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func (s *APIKeyRepoSuite) TestImageStudioKeyIdempotentAndHidden() {
	user := s.mustCreateUser("studio@test.com")
	group := s.mustCreateGroup("studio")
	first, err := s.repo.ImageStudioKey(s.ctx, user.ID, group.ID, "studio-internal-first", "AI 绘图", true)
	s.Require().NoError(err)
	second, err := s.repo.ImageStudioKey(s.ctx, user.ID, group.ID, "studio-internal-second", "AI 绘图", true)
	s.Require().NoError(err)
	s.Equal(first.ID, second.ID)
	s.Equal(first.Key, second.Key)
	s.Equal(service.ImageStudioKeyPurpose, second.Purpose)
	keys, _, err := s.repo.ListByUserID(s.ctx, user.ID, pagination.PaginationParams{Page: 1, PageSize: 10}, service.APIKeyListFilters{})
	s.Require().NoError(err)
	s.Empty(keys)
	_, err = s.repo.ImageStudioKey(s.ctx, user.ID+1000, group.ID, "", "AI 绘图", false)
	s.Require().Error(err)
}

func studioRouteFixture(t *testing.T) (*smartRouteRepository, *service.SmartRouteConfig, *service.SmartRouteConfig) {
	t.Helper()
	ctx := context.Background()
	client := testEntClient(t)
	user, err := client.User.Create().SetEmail(fmt.Sprintf("studio-route-%d@test.com", time.Now().UnixNano())).SetPasswordHash("hash").Save(ctx)
	require.NoError(t, err)
	first, err := client.Group.Create().SetName(fmt.Sprintf("studio-route-first-%d", user.ID)).Save(ctx)
	require.NoError(t, err)
	second, err := client.Group.Create().SetName(fmt.Sprintf("studio-route-second-%d", user.ID)).Save(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = client.APIKey.Delete().Where(apikey.UserIDEQ(user.ID)).Exec(ctx)
		_ = client.Group.DeleteOneID(first.ID).Exec(ctx)
		_ = client.Group.DeleteOneID(second.ID).Exec(ctx)
		_ = client.User.DeleteOneID(user.ID).Exec(ctx)
	})
	key, err := newAPIKeyRepositoryWithSQL(client, integrationDB).ImageStudioPromptRouteKey(ctx, user.ID, fmt.Sprintf("studio-internal-route-%d", user.ID), "AI 绘图")
	require.NoError(t, err)
	config := &service.SmartRouteConfig{
		APIKeyID: key.ID, Mode: service.SmartRouteModeSmart, Platform: service.PlatformOpenAI,
		SubscriptionType: service.SubscriptionTypeStandard, CrossPlatform: true, PreferPlatform: service.PlatformOpenAI,
		CandidateGroupIDs: []int64{first.ID, second.ID}, Strategy: service.SmartRouteStrategyAuto,
		Weights: service.SmartRouteWeights{Price: 40, Speed: 30, Success: 30},
	}
	edited := cloneSmartRouteConfig(config)
	edited.CandidateGroupIDs = []int64{second.ID}
	edited.Strategy = service.SmartRouteStrategyPrice
	return NewSmartRouteRepository(client, nil).(*smartRouteRepository), config, edited
}

func TestImageStudioRouteConcurrentInitialization(t *testing.T) {
	repo, initial, alternate := studioRouteFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	start := make(chan struct{})
	failures := make(chan error, 12)
	for i := 0; i < cap(failures); i++ {
		go func(i int) {
			<-start
			config := initial
			if i%2 == 1 {
				config = alternate
			}
			failures <- repo.CreateImageStudioRouteIfAbsent(ctx, config)
		}(i)
	}
	close(start)
	for i := 0; i < cap(failures); i++ {
		require.NoError(t, <-failures)
	}
	stored, err := repo.Get(ctx, initial.APIKeyID)
	require.NoError(t, err)
	require.NotNil(t, stored)
	if stored.Strategy == initial.Strategy {
		require.Equal(t, initial.CandidateGroupIDs, stored.CandidateGroupIDs)
	} else {
		require.Equal(t, alternate.Strategy, stored.Strategy)
		require.Equal(t, alternate.CandidateGroupIDs, stored.CandidateGroupIDs)
	}
	// A stale negative cache must never make a later bootstrap overwrite a route.
	repo.setCached(ctx, initial.APIKeyID, nil)
	require.NoError(t, repo.CreateImageStudioRouteIfAbsent(ctx, initial))
	repo.Invalidate(ctx, initial.APIKeyID)
	after, err := repo.Get(ctx, initial.APIKeyID)
	require.NoError(t, err)
	require.Equal(t, stored.CandidateGroupIDs, after.CandidateGroupIDs)
	require.Equal(t, stored.Strategy, after.Strategy)
}

func TestImageStudioRouteInitializationWaitsForUserEdit(t *testing.T) {
	repo, initial, edited := studioRouteFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	editWritten := make(chan error, 1)
	releaseEdit := make(chan struct{})
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseEdit) }) }
	defer release()
	editDone := make(chan error, 1)
	go func() {
		editDone <- repo.WithinTransaction(ctx, func(txCtx context.Context) error {
			err := repo.Replace(txCtx, edited)
			editWritten <- err
			if err != nil {
				return err
			}
			select {
			case <-releaseEdit:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	require.NoError(t, <-editWritten)
	repo.setCached(ctx, initial.APIKeyID, nil)
	initDone := make(chan error, 1)
	go func() { initDone <- repo.CreateImageStudioRouteIfAbsent(ctx, initial) }()
	select {
	case err := <-initDone:
		t.Fatalf("initialization completed before the edit released its key lock: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	release()
	require.NoError(t, <-editDone)
	require.NoError(t, <-initDone)
	repo.Invalidate(ctx, initial.APIKeyID)
	stored, err := repo.Get(ctx, initial.APIKeyID)
	require.NoError(t, err)
	require.Equal(t, edited.CandidateGroupIDs, stored.CandidateGroupIDs)
	require.Equal(t, edited.Strategy, stored.Strategy)
}

func TestImageStudioRouteInitializationRollsBackPartialCandidates(t *testing.T) {
	repo, config, _ := studioRouteFixture(t)
	ctx := context.Background()
	config.CandidateGroupIDs = append(config.CandidateGroupIDs, -1)
	require.Error(t, repo.CreateImageStudioRouteIfAbsent(ctx, config))
	stored, err := repo.Get(ctx, config.APIKeyID)
	require.NoError(t, err)
	require.Nil(t, stored)
}
