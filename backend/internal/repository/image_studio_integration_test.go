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
			key, err := repo.ImageStudioKey(ctx, user.ID, group.ID, fmt.Sprintf("studio-internal-%d-%d", user.ID, i), true)
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
	first, err := s.repo.ImageStudioKey(s.ctx, user.ID, group.ID, "studio-internal-first", true)
	s.Require().NoError(err)
	second, err := s.repo.ImageStudioKey(s.ctx, user.ID, group.ID, "studio-internal-second", true)
	s.Require().NoError(err)
	s.Equal(first.ID, second.ID)
	s.Equal(first.Key, second.Key)
	s.Equal(service.ImageStudioKeyPurpose, second.Purpose)
	keys, _, err := s.repo.ListByUserID(s.ctx, user.ID, pagination.PaginationParams{Page: 1, PageSize: 10}, service.APIKeyListFilters{})
	s.Require().NoError(err)
	s.Empty(keys)
	_, err = s.repo.ImageStudioKey(s.ctx, user.ID+1000, group.ID, "", false)
	s.Require().Error(err)
}
