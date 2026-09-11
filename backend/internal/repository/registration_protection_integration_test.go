//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func registrationIntegrationSource() service.RegistrationSource {
	suffix := fmt.Sprint(time.Now().UnixNano())
	return service.RegistrationSource{IPHash: "ip-" + suffix, IdentityHash: "identity-" + suffix,
		IPAddress: "203.0.113.9", UserAgent: "integration-test", Path: "/api/v1/auth/register", Policy: service.DefaultRegistrationProtectionSettings()}
}

func TestRegistrationPostgresConcurrentCreationHonorsQuota(t *testing.T) {
	for _, rotated := range []bool{false, true} {
		t.Run(fmt.Sprint("rotated=", rotated), func(t *testing.T) {
			source := registrationIntegrationSource()
			source.Policy.IPSuccessLimit = 3
			repo := NewUserRepository(integrationEntClient, integrationDB)
			results := make(chan error, 16)
			var wg sync.WaitGroup
			for i := 0; i < 16; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					current := source
					if rotated {
						current.IdentityHash += fmt.Sprint(i)
					}
					user := &service.User{Email: fmt.Sprintf("reg-%s-%d@example.com", source.IPHash, i), PasswordHash: "test", Role: service.RoleUser, Status: service.StatusActive, Concurrency: 3}
					results <- repo.Create(service.WithRegistrationSource(context.Background(), current), user)
				}(i)
			}
			wg.Wait()
			close(results)
			allowed := 0
			for err := range results {
				if err == nil {
					allowed++
					continue
				}
				var quota *service.RegistrationQuotaExceeded
				require.True(t, errors.As(err, &quota), "unexpected create error: %v", err)
				require.Positive(t, quota.RetryAfter)
			}
			expected := 2
			if rotated {
				expected = 3
			}
			require.Equal(t, expected, allowed)
			var count int
			require.NoError(t, integrationDB.QueryRow(`SELECT COUNT(*) FROM registration_source_accounts WHERE ip_hash=$1`, source.IPHash).Scan(&count))
			require.Equal(t, expected, count)
		})
	}
}

func TestRegistrationPostgresLoweredQuotaRetryAfter(t *testing.T) {
	ctx := context.Background()
	source := registrationIntegrationSource()
	source.Policy.IPSuccessLimit = 2
	source.Policy.IdentitySuccessLimit = 2
	for _, age := range []int{20, 10, 1} {
		_, err := integrationDB.Exec(`INSERT INTO registration_source_accounts (ip_hash,identity_hash,ip_address,user_agent,trigger_path,created_at) VALUES ($1,$2,$3,$4,$5,CURRENT_TIMESTAMP - ($6 * INTERVAL '1 hour'))`, source.IPHash, source.IdentityHash, source.IPAddress, source.UserAgent, source.Path, age)
		require.NoError(t, err)
	}
	err := checkRegistrationSourceQuota(ctx, integrationEntClient, source)
	var quota *service.RegistrationQuotaExceeded
	require.ErrorAs(t, err, &quota)
	// Three successes under the old policy: one expiry is insufficient for a
	// new limit of two. The second-oldest record expires in about 14 hours.
	require.InDelta(t, (14 * time.Hour).Seconds(), quota.RetryAfter.Seconds(), 10)
}

func TestRegistrationPostgresOuterRollbackAndFailedSignupReleaseQuota(t *testing.T) {
	ctx := context.Background()
	source := registrationIntegrationSource()
	repo := NewUserRepository(integrationEntClient, integrationDB)
	tx, err := integrationEntClient.Tx(ctx)
	require.NoError(t, err)
	user := &service.User{Email: source.IPHash + "@example.com", PasswordHash: "test", Role: service.RoleUser, Status: service.StatusActive, Concurrency: 3}
	err = repo.Create(service.WithRegistrationSource(dbent.NewTxContext(ctx, tx), source), user)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	var count int
	require.NoError(t, integrationDB.QueryRow(`SELECT COUNT(*) FROM registration_source_accounts WHERE ip_hash=$1`, source.IPHash).Scan(&count))
	require.Zero(t, count)
	require.NoError(t, repo.Create(service.WithRegistrationSource(ctx, source), user))
	require.NoError(t, repo.Delete(service.WithRegistrationRollback(ctx), user.ID))
	require.NoError(t, integrationDB.QueryRow(`SELECT COUNT(*) FROM registration_source_accounts WHERE ip_hash=$1`, source.IPHash).Scan(&count))
	require.Zero(t, count)
}

func TestRegistrationPostgresListAndReviewPreserveBalance(t *testing.T) {
	ctx := context.Background()
	source := registrationIntegrationSource()
	source.Policy.ObserveIdentitySuccessLimit = 1
	user := &service.User{Email: source.IPHash + "@example.com", PasswordHash: "test", Role: service.RoleUser, Status: service.StatusActive, Concurrency: 3, Balance: 12.5}
	require.NoError(t, NewUserRepository(integrationEntClient, integrationDB).Create(service.WithRegistrationSource(ctx, source), user))
	repo := NewRegistrationProtectionRepository(integrationDB)
	for _, kind := range []string{"sources", "events", "blocks", "accounts"} {
		page, err := repo.List(ctx, service.RegistrationRiskFilter{Kind: kind, Page: 1, PageSize: 20, Query: source.IPAddress})
		require.NoError(t, err, kind)
		require.NotEmpty(t, page.Items)
	}
	var id int64
	require.NoError(t, integrationDB.QueryRow(`SELECT id FROM registration_risk_accounts WHERE user_id=$1`, user.ID).Scan(&id))
	_, err := repo.ReviewAccount(ctx, id, user.ID, "restrict", "integration restriction")
	require.NoError(t, err)
	_, err = repo.ReviewAccount(ctx, id, user.ID, "release", "integration release")
	require.NoError(t, err)
	var concurrency int
	var balance float64
	require.NoError(t, integrationDB.QueryRow(`SELECT concurrency,balance FROM users WHERE id=$1`, user.ID).Scan(&concurrency, &balance))
	require.Equal(t, 3, concurrency)
	require.Equal(t, 12.5, balance)
}

func TestRegistrationPostgresBlocksAndDuplicateGrant(t *testing.T) {
	ctx := context.Background()
	source := registrationIntegrationSource()
	repo := NewRegistrationProtectionRepository(integrationDB)
	require.NoError(t, repo.ActivateSourceBlock(ctx, source, "ip_ua", 20))
	require.NoError(t, repo.ActivateSourceBlock(ctx, source, "ip_ua", 21))
	retry, err := repo.CheckSourceBlock(ctx, source)
	require.NoError(t, err)
	require.Positive(t, retry)
	var count int
	var id int64
	require.NoError(t, integrationDB.QueryRow(`SELECT COUNT(*),MIN(id) FROM registration_source_blocks WHERE source_hash=$1`, source.IdentityHash).Scan(&count, &id))
	require.Equal(t, 1, count)
	actor := mustCreateUser(t, integrationEntClient, &service.User{Email: source.IPHash + "-actor@example.com", Concurrency: 3})
	require.NoError(t, repo.ReleaseSourceBlock(ctx, id, actor.ID, "verified", func(context.Context, string) error { return nil }))
	retry, err = repo.CheckSourceBlock(ctx, source)
	require.NoError(t, err)
	require.Zero(t, retry)
	grantRepo := NewSignupRiskGrantRepository(integrationDB)
	for i := 0; i < 2; i++ {
		user := &service.User{Email: fmt.Sprintf("%s-%d@example.com", source.IPHash, i), PasswordHash: "test", Role: service.RoleUser, Status: service.StatusActive, Concurrency: -1}
		require.NoError(t, NewUserRepository(integrationEntClient, integrationDB).Create(service.WithRegistrationSource(ctx, source), user))
		allowed, err := grantRepo.ClaimSignupGrant(ctx, user.ID, source.IdentityHash)
		require.NoError(t, err)
		require.Equal(t, i == 0, allowed)
		if i == 1 {
			var reason, status string
			require.NoError(t, integrationDB.QueryRow(`SELECT reason,status FROM registration_risk_accounts WHERE user_id=$1`, user.ID).Scan(&reason, &status))
			require.Equal(t, "duplicate_signup_identity", reason)
			require.Equal(t, "restricted", status)
		}
	}
}

func TestRegistrationPostgresRetentionPreservesUnresolvedEvidence(t *testing.T) {
	ctx := context.Background()
	source := registrationIntegrationSource()
	source.Policy.ObserveIdentitySuccessLimit = 1
	user := &service.User{Email: source.IPHash + "@example.com", PasswordHash: "test", Role: service.RoleUser, Status: service.StatusActive, Concurrency: 3}
	require.NoError(t, NewUserRepository(integrationEntClient, integrationDB).Create(service.WithRegistrationSource(ctx, source), user))
	_, err := integrationDB.Exec(`UPDATE registration_source_accounts SET created_at=NOW()-INTERVAL '100 days' WHERE user_id=$1`, user.ID)
	require.NoError(t, err)
	_, err = integrationDB.Exec(`UPDATE registration_risk_events SET created_at=NOW()-INTERVAL '100 days' WHERE user_id=$1`, user.ID)
	require.NoError(t, err)
	repo := NewRegistrationProtectionRepository(integrationDB)
	_, err = repo.CleanupBefore(ctx, time.Now().AddDate(0, 0, -90), 1000)
	require.NoError(t, err)
	var count int
	require.NoError(t, integrationDB.QueryRow(`SELECT COUNT(*) FROM registration_source_accounts WHERE user_id=$1`, user.ID).Scan(&count))
	require.Equal(t, 1, count)
	_, err = integrationDB.Exec(`UPDATE registration_risk_accounts SET status='released' WHERE user_id=$1`, user.ID)
	require.NoError(t, err)
	_, err = repo.CleanupBefore(ctx, time.Now().AddDate(0, 0, -90), 1000)
	require.NoError(t, err)
	require.NoError(t, integrationDB.QueryRow(`SELECT COUNT(*) FROM registration_source_accounts WHERE user_id=$1`, user.ID).Scan(&count))
	require.Zero(t, count)
}
