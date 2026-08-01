package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/shopspring/decimal"
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
	rewardType     string
	checkinCount   int64
	hasRecharge    bool
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

func (s *checkinRepoStub) Apply(_ context.Context, userID int64, date, mode string, calculate func(CheckinSettlementState) (decimal.Decimal, decimal.Decimal, string, error)) (*CheckinRecord, bool, error) {
	s.applyCalls++
	if s.existing != nil {
		return s.existing, false, nil
	}
	balance := decimal.NewFromFloat(s.state.Balance)
	reward, randomValue, rewardType, err := calculate(CheckinSettlementState{
		Balance: balance, CheckinCount: s.checkinCount, HasRecharge: s.hasRecharge,
	})
	s.calculateCalls++
	if err != nil {
		return nil, false, err
	}
	s.reward = reward.InexactFloat64()
	s.randomValue = randomValue.InexactFloat64()
	s.rewardType = rewardType
	return &CheckinRecord{
		UserID:        userID,
		CheckinDate:   time.Date(2026, 7, 25, 0, 0, 0, 0, time.Local),
		Mode:          mode,
		RewardType:    rewardType,
		RandomValue:   s.randomValue,
		RewardAmount:  s.reward,
		BalanceBefore: s.state.Balance,
		BalanceAfter:  balance.Add(reward).InexactFloat64(),
	}, true, nil
}

func checkinSettings(values map[string]string) *checkinSettingRepoStub {
	defaults := map[string]string{
		SettingKeyCheckinEnabled:                  "true",
		SettingKeyCheckinNormalEnabled:            "true",
		SettingKeyCheckinLuckyEnabled:             "true",
		SettingKeyCheckinNormalMin:                "0.01",
		SettingKeyCheckinNormalMax:                "0.05",
		SettingKeyCheckinLuckyRewardType:          CheckinRewardTypeMultiplier,
		SettingKeyCheckinLuckyPositiveProbability: "70",
		SettingKeyCheckinLuckyMinMultiply:         "-0.05",
		SettingKeyCheckinLuckyMaxMultiply:         "0.10",
		SettingKeyCheckinLuckyAmountMin:           "-0.05",
		SettingKeyCheckinLuckyAmountMax:           "0.10",
		SettingKeyCheckinRiskEnabled:              "false",
		SettingKeyCheckinMinAccountAge:            "0",
		SettingKeyCheckinIPWindow:                 "10",
		SettingKeyCheckinIPMaxUsers:               "20",
		SettingKeyCheckinUnrechargedEnabled:       "true",
		SettingKeyCheckinUnrechargedThreshold:     "3",
		SettingKeyCheckinUnrechargedNormalPercent: "50",
	}
	for key, value := range values {
		defaults[key] = value
	}
	return &checkinSettingRepoStub{values: defaults}
}

func stubCheckinRandom(t *testing.T, values ...int64) {
	t.Helper()
	original := checkinRandInt
	index := 0
	checkinRandInt = func(_ io.Reader, max *big.Int) (*big.Int, error) {
		if index >= len(values) {
			t.Fatalf("unexpected random read %d", index+1)
		}
		value := values[index]
		index++
		require.GreaterOrEqual(t, value, int64(0))
		require.Less(t, value, max.Int64())
		return big.NewInt(value), nil
	}
	t.Cleanup(func() { checkinRandInt = original })
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
	require.Equal(t, CheckinRewardTypeAmount, record.RewardType)
}

func TestCheckinServiceUnrechargedNormalRewardReduction(t *testing.T) {
	tests := []struct {
		name         string
		checkinCount int64
		hasRecharge  bool
		settings     map[string]string
		wantReward   float64
	}{
		{name: "first check-in", checkinCount: 0, wantReward: 0.03},
		{name: "third check-in", checkinCount: 2, wantReward: 0.03},
		{name: "unrecharged after threshold", checkinCount: 3, wantReward: 0.02},
		{name: "recharged after threshold", checkinCount: 3, hasRecharge: true, wantReward: 0.03},
		{name: "policy disabled", checkinCount: 3, settings: map[string]string{SettingKeyCheckinUnrechargedEnabled: "false"}, wantReward: 0.03},
		{name: "zero percent", checkinCount: 3, settings: map[string]string{SettingKeyCheckinUnrechargedNormalPercent: "0"}, wantReward: 0},
		{name: "full percent", checkinCount: 3, settings: map[string]string{SettingKeyCheckinUnrechargedNormalPercent: "100"}, wantReward: 0.03},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stubCheckinRandom(t, 50_000_000)
			repo := &checkinRepoStub{state: &CheckinUserState{
				Role: RoleUser, Status: StatusActive, Balance: 10,
			}, checkinCount: test.checkinCount, hasRecharge: test.hasRecharge}
			svc := newCheckinServiceForTest(repo, checkinSettings(test.settings))

			record, newlyCheckedIn, err := svc.CheckIn(context.Background(), 7, "normal", "127.0.0.1")

			require.NoError(t, err)
			require.True(t, newlyCheckedIn)
			require.InDelta(t, test.wantReward, record.RewardAmount, 0.000001)
		})
	}
}

func TestCheckinServiceUnrechargedReductionDoesNotAffectLuckyCheckin(t *testing.T) {
	stubCheckinRandom(t, 0, 0, 0)
	repo := &checkinRepoStub{
		state:        &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10},
		checkinCount: 3,
		hasRecharge:  false,
	}
	settings := checkinSettings(map[string]string{SettingKeyCheckinUnrechargedNormalPercent: "0"})

	record, newlyCheckedIn, err := newCheckinServiceForTest(repo, settings).CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

	require.NoError(t, err)
	require.True(t, newlyCheckedIn)
	require.Equal(t, "lucky", record.Mode)
	require.Equal(t, 0.01, record.RandomValue)
	require.Equal(t, 0.10, record.RewardAmount)
}

func TestCheckinServiceLuckyRewardCannotMakeBalanceNegative(t *testing.T) {
	stubCheckinRandom(t, 7_000_000_000, 99)
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 2}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinLuckyMinMultiply: "-1",
		SettingKeyCheckinLuckyMaxMultiply: "0.1",
	})
	svc := newCheckinServiceForTest(repo, settings)

	record, _, err := svc.CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

	require.NoError(t, err)
	require.Equal(t, -2.0, record.RewardAmount)
	require.Equal(t, 0.0, record.BalanceAfter)
	require.Equal(t, -1.0, record.RandomValue)
	require.Equal(t, CheckinRewardTypeMultiplier, record.RewardType)
}

func TestCheckinServiceLuckyLossKeepsTwoDecimalsWhenBalanceHasFractionalCents(t *testing.T) {
	stubCheckinRandom(t, 7_000_000_000, 99)
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 2.009}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinLuckyMinMultiply: "-1",
		SettingKeyCheckinLuckyMaxMultiply: "0.1",
	})

	record, _, err := newCheckinServiceForTest(repo, settings).CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

	require.NoError(t, err)
	require.Equal(t, -2.0, record.RewardAmount)
	require.InDelta(t, 0.009, record.BalanceAfter, 0.00000001)
	require.Equal(t, -1.0, record.RandomValue)
}

func TestCheckinServiceLuckyFixedAmountUsesIndependentRangeAndClampsLoss(t *testing.T) {
	stubCheckinRandom(t, 7_000_000_000, 299)
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 2}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinLuckyRewardType: CheckinRewardTypeAmount,
		SettingKeyCheckinLuckyAmountMin:  "-3",
		SettingKeyCheckinLuckyAmountMax:  "0.1",
	})
	svc := newCheckinServiceForTest(repo, settings)

	record, _, err := svc.CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

	require.NoError(t, err)
	require.Equal(t, -3.0, record.RandomValue)
	require.Equal(t, -2.0, record.RewardAmount)
	require.Equal(t, 0.0, record.BalanceAfter)
	require.Equal(t, CheckinRewardTypeAmount, record.RewardType)
}

func TestCheckinServiceLuckyRewardHandlesLargeBalance(t *testing.T) {
	stubCheckinRandom(t, 0, 0, 9)
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 396368513823.92}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinLuckyMinMultiply: "-0.1",
		SettingKeyCheckinLuckyMaxMultiply: "0.1",
	})
	svc := newCheckinServiceForTest(repo, settings)

	record, _, err := svc.CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

	require.NoError(t, err)
	require.Greater(t, record.RewardAmount, 0.0)
	require.InDelta(t, 39636851382.39, record.RewardAmount, 0.001)
}

func TestCheckinServiceLuckyDirectionUsesConfiguredProbabilityInsteadOfRangeWidth(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		stubCheckinRandom(t, 6_999_999_999, 0, 0)
		repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
		settings := checkinSettings(map[string]string{
			SettingKeyCheckinLuckyPositiveProbability: "70",
			SettingKeyCheckinLuckyMinMultiply:         "-1",
			SettingKeyCheckinLuckyMaxMultiply:         "0.01",
		})

		record, _, err := newCheckinServiceForTest(repo, settings).CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

		require.NoError(t, err)
		require.Equal(t, 0.01, record.RandomValue)
		require.Positive(t, record.RewardAmount)
	})

	t.Run("negative", func(t *testing.T) {
		stubCheckinRandom(t, 7_000_000_000, 0)
		repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
		settings := checkinSettings(map[string]string{
			SettingKeyCheckinLuckyPositiveProbability: "70",
			SettingKeyCheckinLuckyMinMultiply:         "-0.01",
			SettingKeyCheckinLuckyMaxMultiply:         "10",
		})

		record, _, err := newCheckinServiceForTest(repo, settings).CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

		require.NoError(t, err)
		require.Equal(t, -0.01, record.RandomValue)
		require.Negative(t, record.RewardAmount)
	})

	t.Run("two decimal probability precision", func(t *testing.T) {
		stubCheckinRandom(t, 7_000_000_000, 0, 0)
		repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
		settings := checkinSettings(map[string]string{
			SettingKeyCheckinLuckyPositiveProbability: "70.01",
		})

		record, _, err := newCheckinServiceForTest(repo, settings).CheckIn(context.Background(), 7, "lucky", "127.0.0.1")

		require.NoError(t, err)
		require.Positive(t, record.RandomValue)
		require.Positive(t, record.RewardAmount)
	})
}

func TestSecureRandomTieredSignedBetweenUsesConfiguredDirectionAndTiers(t *testing.T) {
	tiers := []CheckinPositiveTier{
		{MinStep: 1, MaxStep: 10, WeightUnits: 7000},
		{MinStep: 11, MaxStep: 15, WeightUnits: 2000},
		{MinStep: 16, MaxStep: 20, WeightUnits: 1000},
	}
	tests := []struct {
		name     string
		random   []int64
		expected float64
	}{
		{name: "positive first tier minimum", random: []int64{5_999_999_999, 0, 0}, expected: 0.01},
		{name: "positive third tier maximum", random: []int64{5_999_999_999, 9999, 4}, expected: 0.20},
		{name: "negative nearest zero", random: []int64{6_000_000_000, 0}, expected: -0.01},
		{name: "negative minimum", random: []int64{6_000_000_000, 7}, expected: -0.08},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubCheckinRandom(t, tt.random...)

			value, err := secureRandomTieredSignedBetween(-0.08, tiers, 60)

			require.NoError(t, err)
			require.Equal(t, tt.expected, value)
		})
	}
}

func TestSecureRandomPositiveTierStepUsesNormalizedWeightsAndInclusiveRanges(t *testing.T) {
	tiers := []CheckinPositiveTier{
		{MinStep: 1, MaxStep: 10, WeightUnits: 7000},
		{MinStep: 11, MaxStep: 15, WeightUnits: 2000},
		{MinStep: 16, MaxStep: 20, WeightUnits: 1000},
	}
	tests := []struct {
		name     string
		random   []int64
		expected float64
	}{
		{name: "first tier lower bound", random: []int64{0, 0}, expected: 0.01},
		{name: "first tier upper bound", random: []int64{6999, 9}, expected: 0.10},
		{name: "second tier lower bound", random: []int64{7000, 0}, expected: 0.11},
		{name: "second tier upper bound", random: []int64{8999, 4}, expected: 0.15},
		{name: "third tier lower bound", random: []int64{9000, 0}, expected: 0.16},
		{name: "third tier upper bound", random: []int64{9999, 4}, expected: 0.20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubCheckinRandom(t, tt.random...)

			value, err := secureRandomPositiveTierStep(tiers)

			require.NoError(t, err)
			require.Equal(t, tt.expected, value)
		})
	}
}

func TestNormalizeCheckinPositiveTiersRequiresContinuousCoverage(t *testing.T) {
	valid := []AdminCheckinPositiveTier{
		{Min: "0.01", Max: "0.10", Weight: "70"},
		{Min: "0.11", Max: "0.15", Weight: "20"},
		{Min: "0.16", Max: "0.20", Weight: "10"},
	}

	parsed, normalized, err := normalizeCheckinPositiveTiers(valid, 0.20)
	require.NoError(t, err)
	require.Len(t, parsed, 3)
	require.Equal(t, int64(7000), parsed[0].WeightUnits)
	require.Equal(t, valid, normalized)

	_, _, err = normalizeCheckinPositiveTiers([]AdminCheckinPositiveTier{
		{Min: "0.01", Max: "0.10", Weight: "70"},
		{Min: "0.12", Max: "0.20", Weight: "30"},
	}, 0.20)
	require.Error(t, err)

	_, _, err = normalizeCheckinPositiveTiers([]AdminCheckinPositiveTier{
		{Min: "0.01", Max: "0.19", Weight: "100"},
	}, 0.20)
	require.Error(t, err)
}

func TestCheckinServiceRejectsInactiveUserBeforeLoadingConfig(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusDisabled}}
	settings := checkinSettings(map[string]string{SettingKeyCheckinNormalMin: "invalid"})
	svc := newCheckinServiceForTest(repo, settings)

	_, _, err := svc.CheckIn(context.Background(), 7, "normal", "127.0.0.1")

	require.ErrorIs(t, err, ErrCheckinNotEligible)
	require.Zero(t, repo.applyCalls)
}

func TestCheckinServiceAllowsActiveAdminAccount(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleAdmin, Status: StatusActive, Balance: 10}}
	svc := newCheckinServiceForTest(repo, checkinSettings(nil))

	status, err := svc.Status(context.Background(), 7)
	require.NoError(t, err)
	require.True(t, status.Eligible)
	require.True(t, status.CanCheckIn)

	record, newlyCheckedIn, err := svc.CheckIn(context.Background(), 7, "normal", "127.0.0.1")
	require.NoError(t, err)
	require.True(t, newlyCheckedIn)
	require.Equal(t, int64(7), record.UserID)
	require.Equal(t, 1, repo.applyCalls)
}

func TestCheckinStatusReturnsEnabledModesAndLuckyMultiplierRange(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinNormalEnabled:    "false",
		SettingKeyCheckinLuckyEnabled:     "true",
		SettingKeyCheckinLuckyMinMultiply: "-0.25",
		SettingKeyCheckinLuckyMaxMultiply: "0.50",
	})
	svc := newCheckinServiceForTest(repo, settings)

	status, err := svc.Status(context.Background(), 7)

	require.NoError(t, err)
	require.True(t, status.Enabled)
	require.False(t, status.NormalEnabled)
	require.True(t, status.LuckyEnabled)
	require.True(t, status.CanCheckIn)
	require.Equal(t, CheckinRewardTypeMultiplier, status.LuckyRewardType)
	require.Equal(t, -0.25, status.LuckyMinMultiplier)
	require.Equal(t, 0.50, status.LuckyMaxMultiplier)
}

func TestCheckinStatusJSONDoesNotExposeLuckyPositiveProbability(t *testing.T) {
	payload, err := json.Marshal(CheckinStatus{})
	require.NoError(t, err)

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(payload, &fields))
	require.NotContains(t, fields, "lucky_positive_probability")
	require.Contains(t, fields, "lucky_min_multiplier")
	require.Contains(t, fields, "lucky_max_multiplier")
}

func TestCheckinStatusDisablesCheckinWhenNoModeIsEnabled(t *testing.T) {
	repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
	settings := checkinSettings(map[string]string{
		SettingKeyCheckinNormalEnabled: "false",
		SettingKeyCheckinLuckyEnabled:  "false",
	})
	svc := newCheckinServiceForTest(repo, settings)

	status, err := svc.Status(context.Background(), 7)

	require.NoError(t, err)
	require.True(t, status.Enabled)
	require.True(t, status.Eligible)
	require.False(t, status.CanCheckIn)
	require.Equal(t, "no_modes_enabled", status.UnavailableReason)
}

func TestCheckinServiceRejectsDisabledModeBeforeSettlement(t *testing.T) {
	testCases := []struct {
		mode     string
		settings map[string]string
	}{
		{mode: "normal", settings: map[string]string{SettingKeyCheckinNormalEnabled: "false"}},
		{mode: "lucky", settings: map[string]string{SettingKeyCheckinLuckyEnabled: "false"}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.mode, func(t *testing.T) {
			repo := &checkinRepoStub{state: &CheckinUserState{Role: RoleUser, Status: StatusActive, Balance: 10}}
			svc := newCheckinServiceForTest(repo, checkinSettings(testCase.settings))

			_, _, err := svc.CheckIn(context.Background(), 7, testCase.mode, "127.0.0.1")

			require.ErrorIs(t, err, ErrCheckinModeDisabled)
			require.Zero(t, repo.applyCalls)
		})
	}
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
	for _, value := range []string{"NaN", "+Inf", "-Inf", "0.001", "1e-3"} {
		_, err := parseCheckinDecimal(value)
		require.Error(t, err, value)
	}
	require.NoError(t, validateCheckinConfigValues("0.01", "1", CheckinRewardTypeMultiplier, "70", "-0.05", "0.10", "-0.05", "0.10", "50", 24, 10, 20, 3))
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
		require.ErrorIs(t, validateCheckinConfigValues("0.01", "0.05", CheckinRewardTypeMultiplier, "70", "-0.05", "0.10", "-0.05", "0.10", "50", testCase.age, testCase.window, testCase.users, 3), ErrCheckinConfigInput)
	}
}

func TestCheckinConfigRejectsInvalidUnrechargedPolicy(t *testing.T) {
	tests := []struct {
		name      string
		threshold int
		percent   string
	}{
		{name: "zero threshold", threshold: 0, percent: "50"},
		{name: "threshold above limit", threshold: 3651, percent: "50"},
		{name: "negative percentage", threshold: 3, percent: "-0.01"},
		{name: "percentage above limit", threshold: 3, percent: "100.01"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCheckinConfigValues(
				"0.01", "0.05", CheckinRewardTypeMultiplier, "70",
				"-0.05", "0.10", "-0.05", "0.10", test.percent,
				24, 10, 20, test.threshold,
			)
			require.ErrorIs(t, err, ErrCheckinConfigInput)
		})
	}
}

func TestCheckinConfigRejectsInvalidLuckyRewardSettings(t *testing.T) {
	testCases := []struct {
		name          string
		rewardType    string
		probability   string
		multiplierMin string
		multiplierMax string
		amountMin     string
		amountMax     string
	}{
		{name: "unknown reward type", rewardType: "credits", probability: "70", multiplierMin: "-0.05", multiplierMax: "0.10", amountMin: "-0.05", amountMax: "0.10"},
		{name: "probability below lower bound", rewardType: CheckinRewardTypeMultiplier, probability: "-0.01", multiplierMin: "-0.05", multiplierMax: "0.10", amountMin: "-0.05", amountMax: "0.10"},
		{name: "probability above upper bound", rewardType: CheckinRewardTypeMultiplier, probability: "100.01", multiplierMin: "-0.05", multiplierMax: "0.10", amountMin: "-0.05", amountMax: "0.10"},
		{name: "multiplier below lower bound", rewardType: CheckinRewardTypeMultiplier, probability: "70", multiplierMin: "-1.00000001", multiplierMax: "0.10", amountMin: "-0.05", amountMax: "0.10"},
		{name: "multiplier must include negative side", rewardType: CheckinRewardTypeMultiplier, probability: "70", multiplierMin: "0", multiplierMax: "0.10", amountMin: "-0.05", amountMax: "0.10"},
		{name: "multiplier above upper bound", rewardType: CheckinRewardTypeMultiplier, probability: "70", multiplierMin: "-0.05", multiplierMax: "10.01", amountMin: "-0.05", amountMax: "0.10"},
		{name: "amount below lower bound", rewardType: CheckinRewardTypeAmount, probability: "70", multiplierMin: "-0.05", multiplierMax: "0.10", amountMin: "-100.01", amountMax: "0.10"},
		{name: "amount must include positive side", rewardType: CheckinRewardTypeAmount, probability: "70", multiplierMin: "-0.05", multiplierMax: "0.10", amountMin: "-0.10", amountMax: "0"},
		{name: "amount above upper bound", rewardType: CheckinRewardTypeAmount, probability: "70", multiplierMin: "-0.05", multiplierMax: "0.10", amountMin: "-0.10", amountMax: "100.01"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := validateCheckinConfigValues(
				"0.01", "0.05",
				testCase.rewardType,
				testCase.probability,
				testCase.multiplierMin, testCase.multiplierMax,
				testCase.amountMin, testCase.amountMax,
				"50", 24, 10, 20, 3,
			)
			require.ErrorIs(t, err, ErrCheckinConfigInput)
		})
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
		Enabled:                     true,
		NormalEnabled:               true,
		LuckyEnabled:                true,
		NormalMin:                   "0.01",
		NormalMax:                   "0.05",
		LuckyRewardType:             CheckinRewardTypeMultiplier,
		LuckyPositiveProbability:    "70",
		LuckyMinMultiply:            "-0.05",
		LuckyMaxMultiply:            "0.10",
		LuckyAmountMin:              "-0.05",
		LuckyAmountMax:              "0.10",
		RiskEnabled:                 true,
		MinAccountAgeHours:          24,
		IPWindowMinutes:             10,
		IPMaxUsers:                  20,
		UnrechargedEnabled:          true,
		UnrechargedCheckinThreshold: 3,
		UnrechargedNormalPercent:    "50",
		ExpectedVersion:             2,
		ChangeReason:                "test",
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
		Enabled:                  false,
		NormalEnabled:            true,
		LuckyEnabled:             false,
		NormalMin:                "0.02",
		NormalMax:                "0.08",
		LuckyRewardType:          CheckinRewardTypeAmount,
		LuckyPositiveProbability: "65",
		LuckyMultiplierPositiveTiers: []AdminCheckinPositiveTier{
			{Min: "0.01", Max: "0.10", Weight: "70"},
			{Min: "0.11", Max: "0.15", Weight: "20"},
			{Min: "0.16", Max: "0.20", Weight: "10"},
		},
		LuckyAmountPositiveTiers:    []AdminCheckinPositiveTier{{Min: "0.01", Max: "1.00", Weight: "100"}},
		LuckyMinMultiply:            "-0.10",
		LuckyMaxMultiply:            "0.20",
		LuckyAmountMin:              "-0.50",
		LuckyAmountMax:              "1.00",
		RiskEnabled:                 true,
		MinAccountAgeHours:          24,
		IPWindowMinutes:             10,
		IPMaxUsers:                  20,
		UnrechargedEnabled:          true,
		UnrechargedCheckinThreshold: 3,
		UnrechargedNormalPercent:    "50",
		ExpectedVersion:             3,
		ChangeReason:                "adjust test range",
	})

	require.NoError(t, err)
	require.Equal(t, int64(4), result.ConfigVersion)
	require.Equal(t, "0.02", settings.values[SettingKeyCheckinNormalMin])
	require.Equal(t, "true", settings.values[SettingKeyCheckinNormalEnabled])
	require.Equal(t, "false", settings.values[SettingKeyCheckinLuckyEnabled])
	require.False(t, result.LuckyEnabled)
	require.Equal(t, CheckinRewardTypeAmount, settings.values[SettingKeyCheckinLuckyRewardType])
	require.Equal(t, "65", settings.values[SettingKeyCheckinLuckyPositiveProbability])
	require.JSONEq(t, `[{"min":"0.01","max":"0.10","weight":"70"},{"min":"0.11","max":"0.15","weight":"20"},{"min":"0.16","max":"0.20","weight":"10"}]`, settings.values[SettingKeyCheckinLuckyMultiplierPositiveTiers])
	require.Len(t, result.LuckyMultiplierPositiveTiers, 3)
	require.Equal(t, "1.00", settings.values[SettingKeyCheckinLuckyAmountMax])
	require.Equal(t, "true", settings.values[SettingKeyCheckinUnrechargedEnabled])
	require.Equal(t, "3", settings.values[SettingKeyCheckinUnrechargedThreshold])
	require.Equal(t, "50", settings.values[SettingKeyCheckinUnrechargedNormalPercent])
	require.Equal(t, "4", settings.values[SettingKeyCheckinConfigVersion])
}

func TestAdminCheckinConfigRejectsConcurrentVersionChange(t *testing.T) {
	settings := checkinSettings(map[string]string{SettingKeyCheckinConfigVersion: "3"})
	svc := NewAdminCheckinService(adminCheckinRepoStub{settings: settings, forceConflict: true}, settings)

	_, err := svc.UpdateConfig(context.Background(), AdminCheckinConfigUpdate{
		Enabled:                     true,
		NormalEnabled:               true,
		LuckyEnabled:                true,
		NormalMin:                   "0.01",
		NormalMax:                   "0.05",
		LuckyRewardType:             CheckinRewardTypeMultiplier,
		LuckyPositiveProbability:    "70",
		LuckyMinMultiply:            "-0.05",
		LuckyMaxMultiply:            "0.10",
		LuckyAmountMin:              "-0.05",
		LuckyAmountMax:              "0.10",
		RiskEnabled:                 true,
		MinAccountAgeHours:          24,
		IPWindowMinutes:             10,
		IPMaxUsers:                  20,
		UnrechargedEnabled:          true,
		UnrechargedCheckinThreshold: 3,
		UnrechargedNormalPercent:    "50",
		ExpectedVersion:             3,
		ChangeReason:                "concurrent update",
	})

	require.ErrorIs(t, err, ErrCheckinConfigVersion)
}
