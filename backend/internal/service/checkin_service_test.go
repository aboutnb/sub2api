package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type checkinSettingRepoStub struct {
	values map[string]string
	err    error
}

func (s *checkinSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	return nil, ErrSettingNotFound
}

func (s *checkinSettingRepoStub) GetValue(context.Context, string) (string, error) {
	return "", ErrSettingNotFound
}

func (s *checkinSettingRepoStub) Set(_ context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *checkinSettingRepoStub) GetMultiple(_ context.Context, keys []string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			values[key] = value
		}
	}
	return values, nil
}

func (s *checkinSettingRepoStub) SetMultiple(_ context.Context, updates map[string]string) error {
	for key, value := range updates {
		if err := s.Set(context.Background(), key, value); err != nil {
			return err
		}
	}
	return nil
}
func (s *checkinSettingRepoStub) GetAll(context.Context) (map[string]string, error) { return nil, nil }
func (s *checkinSettingRepoStub) Delete(context.Context, string) error              { return nil }

type checkinRepoStub struct {
	state          *CheckinUserState
	stateCalls     int
	existing       *CheckinRecord
	applyCalls     int
	calculateCalls int
	reward         float64
	randomValue    float64
}

type checkinAbuseGuardStub struct {
	allowed       bool
	err           error
	calls         int
	requestDenied bool
	requestErr    error
	requestCalls  int
	requestRetry  time.Duration
	source        string
	userID        int64
	window        time.Duration
	maxUsers      int
}

func (s *checkinAbuseGuardStub) CheckRequest(_ context.Context, source string, userID int64, window time.Duration, _, _ int) (bool, time.Duration, error) {
	s.requestCalls++
	s.source = source
	s.userID = userID
	if s.requestRetry <= 0 {
		s.requestRetry = window
	}
	return !s.requestDenied, s.requestRetry, s.requestErr
}

func (s *checkinAbuseGuardStub) CheckAndRecord(_ context.Context, source string, userID int64, window time.Duration, maxUsers int) (bool, int64, time.Duration, error) {
	s.calls++
	s.source = source
	s.userID = userID
	s.window = window
	s.maxUsers = maxUsers
	return s.allowed, 1, window, s.err
}

func (s *checkinRepoStub) GetUserState(context.Context, int64) (*CheckinUserState, error) {
	s.stateCalls++
	return s.state, nil
}

func (s *checkinRepoStub) GetByDate(context.Context, int64, string) (*CheckinRecord, error) {
	if s.existing == nil {
		return nil, sql.ErrNoRows
	}
	return s.existing, nil
}

func (s *checkinRepoStub) List(context.Context, int64, int, int) ([]CheckinRecord, int64, error) {
	return nil, 0, nil
}

func (s *checkinRepoStub) Apply(_ context.Context, userID int64, date, mode string, calculate func(float64) (float64, float64, error)) (*CheckinRecord, bool, error) {
	s.applyCalls++
	if s.existing != nil {
		return s.existing, false, nil
	}
	reward, randomValue, err := calculate(s.state.Balance)
	s.calculateCalls++
	if err != nil {
		return nil, false, err
	}
	s.reward = reward
	s.randomValue = randomValue
	return &CheckinRecord{
		UserID:        userID,
		CheckinDate:   time.Date(2026, 7, 25, 0, 0, 0, 0, time.Local),
		Mode:          mode,
		RandomValue:   randomValue,
		RewardAmount:  reward,
		BalanceBefore: s.state.Balance,
		BalanceAfter:  s.state.Balance + reward,
	}, true, nil
}

func checkinSettings(values map[string]string) *checkinSettingRepoStub {
	defaults := map[string]string{
		SettingKeyCheckinEnabled:          "true",
		SettingKeyCheckinNormalMin:        "0.01",
		SettingKeyCheckinNormalMax:        "0.05",
		SettingKeyCheckinLuckyMinMultiply: "-0.05",
		SettingKeyCheckinLuckyMaxMultiply: "0.10",
		SettingKeyCheckinRiskEnabled:      "false",
		SettingKeyCheckinMinAccountAge:    "0",
		SettingKeyCheckinIPWindow:         "10",
		SettingKeyCheckinIPMaxUsers:       "20",
	}
	for key, value := range values {
		defaults[key] = value
	}
	return &checkinSettingRepoStub{values: defaults}
}

func newCheckinServiceForTest(repo CheckinRepository, settings SettingRepository) *CheckinService {
	return NewCheckinService(repo, settings, &config.Config{RunMode: config.RunModeStandard}, nil, &checkinAbuseGuardStub{allowed: true})
}

func TestCheckinServiceNormalRewardUsesConfiguredRange(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
	svc := newCheckinServiceForTest(repo, checkinSettings(nil))

	record, newlyCheckedIn, err := svc.CheckIn(context.Background(), 7, "normal", "127.0.0.1")

	require.NoError(t, err)
	require.True(t, newlyCheckedIn)
	require.Equal(t, "normal", record.Mode)
	require.GreaterOrEqual(t, repo.reward, 0.01)
	require.LessOrEqual(t, repo.reward, 0.05)
	require.Equal(t, repo.reward, repo.randomValue)
}

func TestCheckinServiceLuckyRewardCannotMakeBalanceNegative(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 2}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinLuckyMinMultiply: "-1",
		SettingKeyCheckinLuckyMaxMultiply: "-1",
	})
	svc := newCheckinServiceForTest(repo, settings)

	record, _, err := svc.CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

	require.NoError(t, err)
	require.Equal(t, -2.0, record.RewardAmount)
	require.Equal(t, 0.0, record.BalanceAfter)
	require.Equal(t, -1.0, record.RandomValue)
}

func TestCheckinServiceLuckyRewardHandlesLargeBalance(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 396368513823.92}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinLuckyMinMultiply: "0.1",
		SettingKeyCheckinLuckyMaxMultiply: "0.1",
	})
	svc := newCheckinServiceForTest(repo, settings)

	record, _, err := svc.CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

	require.NoError(t, err)
	require.Greater(t, record.RewardAmount, 0.0)
	require.InDelta(t, 39636851382.392, record.RewardAmount, 0.001)
}

func TestCheckinServiceRejectsIneligibleUserBeforeLoadingConfig(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleAdmin, Status: StatusActive}}
	settings := checkinSettings(map[string]string{SettingKeyCheckinNormalMin: "invalid"})
	svc := newCheckinServiceForTest(repo, settings)

	_, _, err := svc.CheckIn(context.Background(), 7, "normal", "127.0.0.1")

	require.ErrorIs(t, err, ErrCheckinNotEligible)
	require.Zero(t, repo.applyCalls)
}

func TestCheckinServiceRateLimitsNegativeBalanceBeforeSettlement(t *testing.T) {
	guard := &checkinAbuseGuardStub{allowed: true}
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: -100}}
	svc := NewCheckinService(repo, checkinSettings(nil), &config.Config{RunMode: config.RunModeStandard}, nil, guard)

	_, _, err := svc.CheckIn(context.Background(), 7, "normal", "192.0.2.10")

	require.ErrorIs(t, err, ErrCheckinNegativeBalance)
	require.Equal(t, 1, guard.requestCalls)
	require.Zero(t, repo.applyCalls)

	status, err := svc.Status(context.Background(), 7)
	require.NoError(t, err)
	require.False(t, status.Eligible)
	require.Equal(t, "negative_balance", status.UnavailableReason)
}

func TestCheckinServiceDuplicateDoesNotRecalculate(t *testing.T) {
	existing := &CheckinRecord{ID: 11, UserID: 7, Mode: "normal", RewardAmount: 0.03}
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}, existing: existing}
	svc := newCheckinServiceForTest(repo, checkinSettings(map[string]string{
		SettingKeyCheckinEnabled:   "false",
		SettingKeyCheckinNormalMin: "invalid",
	}))

	record, newlyCheckedIn, err := svc.CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

	require.NoError(t, err)
	require.False(t, newlyCheckedIn)
	require.Same(t, existing, record)
	require.Zero(t, repo.applyCalls)
	require.Zero(t, repo.calculateCalls)
}

func TestCheckinStatusPreservesInvalidConfigReason(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
	svc := newCheckinServiceForTest(repo, checkinSettings(map[string]string{SettingKeyCheckinNormalMin: "invalid"}))

	status, err := svc.Status(context.Background(), 7)

	require.NoError(t, err)
	require.True(t, status.Eligible)
	require.False(t, status.Enabled)
	require.False(t, status.CanCheckIn)
	require.Equal(t, "config_invalid", status.UnavailableReason)
}

func TestCheckinDecimalRejectsNonFiniteAndOverPrecisionValues(t *testing.T) {
	for _, value := range []string{"NaN", "+Inf", "-Inf", "0.000000001", "1e-3"} {
		_, err := parseCheckinDecimal(value)
		require.Error(t, err, value)
	}
	require.NoError(t, validateCheckinConfigValues("0.00000001", "1", "-0.05", "0.10", 24, 10, 20))
}

func TestCheckinServiceRejectsNewAccountBeforeRiskGuard(t *testing.T) {
	guard := &checkinAbuseGuardStub{allowed: true}
	repo := &checkinRepoStub{state: &CheckinUserState{
		Role: RoleUser, Status: StatusActive, Balance: 10, CreatedAt: time.Now().Add(-23 * time.Hour),
	}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinRiskEnabled:   "true",
		SettingKeyCheckinMinAccountAge: "24",
	})
	svc := NewCheckinService(repo, settings, &config.Config{RunMode: config.RunModeStandard}, nil, guard)

	_, _, err := svc.CheckIn(context.Background(), 7, "normal", "127.0.0.1")

	require.ErrorIs(t, err, ErrCheckinAccountTooNew)
	require.Zero(t, guard.calls)
	require.Equal(t, 1, guard.requestCalls)
	require.Zero(t, repo.applyCalls)
}

func TestCheckinServiceAllowsOldAccountThroughRiskGuard(t *testing.T) {
	guard := &checkinAbuseGuardStub{allowed: true}
	repo := &checkinRepoStub{state: &CheckinUserState{
		Role: RoleUser, Status: StatusActive, Balance: 10, CreatedAt: time.Now().Add(-48 * time.Hour),
	}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinRiskEnabled:   "true",
		SettingKeyCheckinMinAccountAge: "24",
		SettingKeyCheckinIPWindow:      "12",
		SettingKeyCheckinIPMaxUsers:    "3",
	})
	svc := NewCheckinService(repo, settings, &config.Config{RunMode: config.RunModeStandard}, nil, guard)

	_, newlyCheckedIn, err := svc.CheckIn(context.Background(), 7, "normal", "2001:0db8::1")

	require.NoError(t, err)
	require.True(t, newlyCheckedIn)
	require.Equal(t, 1, guard.calls)
	require.Equal(t, 1, guard.requestCalls)
	require.Equal(t, "2001:db8::1", guard.source)
	require.Equal(t, int64(7), guard.userID)
	require.Equal(t, 12*time.Minute, guard.window)
	require.Equal(t, 3, guard.maxUsers)
}

func TestCheckinServiceRiskDisabledBypassesCampaignChecksButNotSecurityGuard(t *testing.T) {
	guard := &checkinAbuseGuardStub{allowed: false}
	repo := &checkinRepoStub{state: &CheckinUserState{
		Role: RoleUser, Status: StatusActive, Balance: 10, CreatedAt: time.Now(),
	}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinRiskEnabled:   "false",
		SettingKeyCheckinMinAccountAge: "720",
	})
	svc := NewCheckinService(repo, settings, &config.Config{RunMode: config.RunModeStandard}, nil, guard)

	_, newlyCheckedIn, err := svc.CheckIn(context.Background(), 7, "normal", "192.0.2.10")

	require.NoError(t, err)
	require.True(t, newlyCheckedIn)
	require.Zero(t, guard.calls)
	require.Equal(t, 1, guard.requestCalls)
}

func TestCheckinServiceDuplicateBypassesChangedRiskRules(t *testing.T) {
	existing := &CheckinRecord{ID: 11, UserID: 7, Mode: "normal", RewardAmount: 0.03}
	guard := &checkinAbuseGuardStub{allowed: false}
	repo := &checkinRepoStub{
		state:    &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10, CreatedAt: time.Now()},
		existing: existing,
	}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinRiskEnabled:   "true",
		SettingKeyCheckinMinAccountAge: "720",
	})
	svc := NewCheckinService(repo, settings, &config.Config{RunMode: config.RunModeStandard}, nil, guard)

	record, newlyCheckedIn, err := svc.CheckIn(context.Background(), 7, "lucky", "192.0.2.10")

	require.NoError(t, err)
	require.False(t, newlyCheckedIn)
	require.Same(t, existing, record)
	require.Zero(t, guard.calls)
	require.Equal(t, 1, guard.requestCalls)
}

func TestCheckinServiceRejectsLimitedSource(t *testing.T) {
	guard := &checkinAbuseGuardStub{allowed: false}
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10, CreatedAt: time.Now().Add(-48 * time.Hour)}}
	settings := checkinSettings(map[string]string{SettingKeyCheckinRiskEnabled: "true", SettingKeyCheckinMinAccountAge: "24"})
	svc := NewCheckinService(repo, settings, &config.Config{RunMode: config.RunModeStandard}, nil, guard)

	_, _, err := svc.CheckIn(context.Background(), 7, "normal", "192.0.2.10")

	require.ErrorIs(t, err, ErrCheckinSourceLimited)
	require.Equal(t, 1, guard.calls)
	require.Equal(t, 1, guard.requestCalls)
	require.Zero(t, repo.applyCalls)
}

func TestCheckinServiceFailsClosedWhenSecurityGuardUnavailable(t *testing.T) {
	testCases := []struct {
		name   string
		source string
		guard  CheckinAbuseGuard
	}{
		{name: "missing source", source: "", guard: &checkinAbuseGuardStub{allowed: true}},
		{name: "invalid source", source: "not-an-ip", guard: &checkinAbuseGuardStub{allowed: true}},
		{name: "missing guard", source: "192.0.2.10", guard: nil},
		{name: "redis error", source: "192.0.2.10", guard: &checkinAbuseGuardStub{allowed: true, requestErr: errors.New("redis unavailable")}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10, CreatedAt: time.Now().Add(-48 * time.Hour)}}
			settings := checkinSettings(map[string]string{SettingKeyCheckinRiskEnabled: "true", SettingKeyCheckinMinAccountAge: "24"})
			svc := NewCheckinService(repo, settings, &config.Config{RunMode: config.RunModeStandard}, nil, testCase.guard)

			_, _, err := svc.CheckIn(context.Background(), 7, "normal", testCase.source)

			require.ErrorIs(t, err, ErrCheckinSecurityUnavailable)
			require.Zero(t, repo.applyCalls)
		})
	}
}

func TestCheckinServiceFailsClosedWhenCampaignRiskGuardUnavailable(t *testing.T) {
	guard := &checkinAbuseGuardStub{allowed: true, err: errors.New("redis unavailable")}
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10, CreatedAt: time.Now().Add(-48 * time.Hour)}}
	settings := checkinSettings(map[string]string{SettingKeyCheckinRiskEnabled: "true", SettingKeyCheckinMinAccountAge: "24"})
	svc := NewCheckinService(repo, settings, &config.Config{RunMode: config.RunModeStandard}, nil, guard)

	_, _, err := svc.CheckIn(context.Background(), 7, "normal", "192.0.2.10")

	require.ErrorIs(t, err, ErrCheckinRiskUnavailable)
	require.Equal(t, 1, guard.requestCalls)
	require.Equal(t, 1, guard.calls)
	require.Zero(t, repo.applyCalls)
}

func TestCheckinServiceRejectsExcessiveRequestsWithRetryAfter(t *testing.T) {
	guard := &checkinAbuseGuardStub{allowed: true, requestDenied: true, requestRetry: 12 * time.Second}
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
	svc := NewCheckinService(repo, checkinSettings(nil), &config.Config{RunMode: config.RunModeStandard}, nil, guard)

	_, _, err := svc.CheckIn(context.Background(), 7, "normal", "192.0.2.10")

	require.ErrorIs(t, err, ErrCheckinRateLimited)
	require.Equal(t, 12, RetryAfterSecondsFromError(err))
	require.Zero(t, repo.stateCalls)
	require.Zero(t, repo.applyCalls)
}

func TestCheckinServiceFailsClosedWhenEntropyUnavailable(t *testing.T) {
	original := checkinRandInt
	checkinRandInt = func(io.Reader, *big.Int) (*big.Int, error) {
		return nil, errors.New("entropy unavailable")
	}
	t.Cleanup(func() { checkinRandInt = original })

	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
	svc := newCheckinServiceForTest(repo, checkinSettings(nil))

	_, _, err := svc.CheckIn(context.Background(), 7, "normal", "192.0.2.10")

	require.ErrorIs(t, err, ErrCheckinEntropyUnavailable)
	require.Equal(t, 1, repo.calculateCalls)
}

func TestCheckinStatusMarksNewAccountIneligible(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10, CreatedAt: time.Now()}}
	settings := checkinSettings(map[string]string{SettingKeyCheckinRiskEnabled: "true", SettingKeyCheckinMinAccountAge: "24"})
	svc := newCheckinServiceForTest(repo, settings)

	status, err := svc.Status(context.Background(), 7)

	require.NoError(t, err)
	require.True(t, status.Enabled)
	require.False(t, status.Eligible)
	require.False(t, status.CanCheckIn)
	require.Equal(t, "account_too_new", status.UnavailableReason)
}

func TestCheckinConfigRejectsInvalidRiskRanges(t *testing.T) {
	for _, testCase := range []struct {
		age, window, users int
	}{
		{age: -1, window: 10, users: 20},
		{age: 721, window: 10, users: 20},
		{age: 24, window: 0, users: 20},
		{age: 24, window: 1441, users: 20},
		{age: 24, window: 10, users: 0},
		{age: 24, window: 10, users: 10001},
	} {
		require.ErrorIs(t, validateCheckinConfigValues("0.01", "0.05", "-0.05", "0.10", testCase.age, testCase.window, testCase.users), ErrCheckinConfigInput)
	}
}

type adminCheckinRepoStub struct {
	settings      *checkinSettingRepoStub
	forceConflict bool
}

func (adminCheckinRepoStub) AdminOverview(context.Context, string) (*AdminCheckinOverview, error) {
	return &AdminCheckinOverview{}, nil
}

func (adminCheckinRepoStub) AdminList(context.Context, AdminCheckinRecordFilter) ([]AdminCheckinRecord, int64, error) {
	return nil, 0, nil
}

func (s adminCheckinRepoStub) UpdateConfigIfVersion(_ context.Context, expectedVersion int64, values map[string]string) (bool, error) {
	if s.forceConflict {
		return false, nil
	}
	if s.settings == nil {
		return true, nil
	}
	currentVersion, err := strconv.ParseInt(s.settings.values[SettingKeyCheckinConfigVersion], 10, 64)
	if err != nil {
		return false, err
	}
	if currentVersion != expectedVersion {
		return false, nil
	}
	for key, value := range values {
		s.settings.values[key] = value
	}
	s.settings.values[SettingKeyCheckinConfigVersion] = strconv.FormatInt(currentVersion+1, 10)
	return true, nil
}

func TestAdminCheckinConfigRequiresReasonAndMatchingVersion(t *testing.T) {
	settings := checkinSettings(map[string]string{SettingKeyCheckinConfigVersion: "3"})
	svc := NewAdminCheckinService(adminCheckinRepoStub{}, settings)
	input := AdminCheckinConfigUpdate{
		Enabled:            true,
		NormalMin:          "0.01",
		NormalMax:          "0.05",
		LuckyMinMultiply:   "-0.05",
		LuckyMaxMultiply:   "0.10",
		RiskEnabled:        true,
		MinAccountAgeHours: 24,
		IPWindowMinutes:    10,
		IPMaxUsers:         20,
		ExpectedVersion:    2,
		ChangeReason:       "test",
	}

	_, err := svc.UpdateConfig(context.Background(), input)
	require.ErrorIs(t, err, ErrCheckinConfigVersion)

	input.ExpectedVersion = 3
	input.ChangeReason = ""
	_, err = svc.UpdateConfig(context.Background(), input)
	require.ErrorIs(t, err, ErrCheckinChangeReason)

	input.ChangeReason = "test"
	input.ExpectedVersion = 0
	_, err = svc.UpdateConfig(context.Background(), input)
	require.ErrorIs(t, err, ErrCheckinConfigInput)
}

func TestAdminCheckinConfigUpdateIncrementsVersion(t *testing.T) {
	settings := checkinSettings(map[string]string{SettingKeyCheckinConfigVersion: "3"})
	svc := NewAdminCheckinService(adminCheckinRepoStub{settings: settings}, settings)

	result, err := svc.UpdateConfig(context.Background(), AdminCheckinConfigUpdate{
		Enabled:            false,
		NormalMin:          "0.02",
		NormalMax:          "0.08",
		LuckyMinMultiply:   "-0.10",
		LuckyMaxMultiply:   "0.20",
		RiskEnabled:        true,
		MinAccountAgeHours: 24,
		IPWindowMinutes:    10,
		IPMaxUsers:         20,
		ExpectedVersion:    3,
		ChangeReason:       "adjust test range",
	})

	require.NoError(t, err)
	require.Equal(t, int64(4), result.ConfigVersion)
	require.Equal(t, "0.02", settings.values[SettingKeyCheckinNormalMin])
	require.Equal(t, "4", settings.values[SettingKeyCheckinConfigVersion])
}

func TestAdminCheckinConfigRejectsConcurrentVersionChange(t *testing.T) {
	settings := checkinSettings(map[string]string{SettingKeyCheckinConfigVersion: "3"})
	svc := NewAdminCheckinService(adminCheckinRepoStub{settings: settings, forceConflict: true}, settings)

	_, err := svc.UpdateConfig(context.Background(), AdminCheckinConfigUpdate{
		Enabled:            true,
		NormalMin:          "0.01",
		NormalMax:          "0.05",
		LuckyMinMultiply:   "-0.05",
		LuckyMaxMultiply:   "0.10",
		RiskEnabled:        true,
		MinAccountAgeHours: 24,
		IPWindowMinutes:    10,
		IPMaxUsers:         20,
		ExpectedVersion:    3,
		ChangeReason:       "concurrent update",
	})

	require.ErrorIs(t, err, ErrCheckinConfigVersion)
}
