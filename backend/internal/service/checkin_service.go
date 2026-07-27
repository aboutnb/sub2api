package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/shopspring/decimal"
)

const (
	SettingKeyCheckinEnabled                  = "checkin_enabled"
	SettingKeyCheckinNormalEnabled            = "checkin_normal_enabled"
	SettingKeyCheckinLuckyEnabled             = "checkin_lucky_enabled"
	SettingKeyCheckinNormalMin                = "checkin_normal_min"
	SettingKeyCheckinNormalMax                = "checkin_normal_max"
	SettingKeyCheckinLuckyRewardType          = "checkin_lucky_reward_type"
	SettingKeyCheckinLuckyPositiveProbability = "checkin_lucky_positive_probability"
	SettingKeyCheckinLuckyMinMultiply         = "checkin_lucky_min_multiplier"
	SettingKeyCheckinLuckyMaxMultiply         = "checkin_lucky_max_multiplier"
	SettingKeyCheckinLuckyAmountMin           = "checkin_lucky_amount_min"
	SettingKeyCheckinLuckyAmountMax           = "checkin_lucky_amount_max"
	SettingKeyCheckinRiskEnabled              = "checkin_risk_control_enabled"
	SettingKeyCheckinMinAccountAge            = "checkin_min_account_age_hours"
	SettingKeyCheckinIPWindow                 = "checkin_ip_window_minutes"
	SettingKeyCheckinIPMaxUsers               = "checkin_ip_max_users"
	SettingKeyCheckinConfigVersion            = "checkin_config_version"

	// These hard safety ceilings are intentionally separate from the editable
	// business-risk settings so disabling campaign risk controls cannot disable
	// endpoint protection.
	CheckinRequestWindow      = time.Minute
	CheckinUserRequestLimit   = 10
	CheckinSourceRequestLimit = 60

	CheckinRewardTypeAmount     = "amount"
	CheckinRewardTypeMultiplier = "multiplier"
	CheckinCalculationScale     = 2
)

var (
	ErrCheckinDisabled            = infraerrors.Forbidden("CHECKIN_DISABLED", "daily check-in is currently unavailable")
	ErrCheckinModeDisabled        = infraerrors.Forbidden("CHECKIN_MODE_DISABLED", "this check-in mode is currently unavailable")
	ErrCheckinNotEligible         = infraerrors.Forbidden("CHECKIN_NOT_ELIGIBLE", "this account cannot use daily check-in")
	ErrCheckinInvalidMode         = infraerrors.BadRequest("CHECKIN_INVALID_MODE", "mode must be normal or lucky")
	ErrCheckinConfigInvalid       = infraerrors.InternalServer("CHECKIN_CONFIG_INVALID", "daily check-in configuration is invalid")
	ErrCheckinUserNotFound        = infraerrors.NotFound("USER_NOT_FOUND", "user not found")
	ErrCheckinAccountTooNew       = infraerrors.Forbidden("CHECKIN_ACCOUNT_TOO_NEW", "account is too new for daily check-in")
	ErrCheckinNegativeBalance     = infraerrors.Forbidden("CHECKIN_NEGATIVE_BALANCE", "accounts with a negative balance cannot use daily check-in")
	ErrCheckinSourceLimited       = infraerrors.TooManyRequests("CHECKIN_SOURCE_LIMITED", "too many accounts checked in from this source")
	ErrCheckinRiskUnavailable     = infraerrors.ServiceUnavailable("CHECKIN_RISK_UNAVAILABLE", "daily check-in risk control is temporarily unavailable")
	ErrCheckinRateLimited         = infraerrors.TooManyRequests("CHECKIN_RATE_LIMITED", "too many daily check-in requests")
	ErrCheckinSecurityUnavailable = infraerrors.ServiceUnavailable(
		"CHECKIN_SECURITY_UNAVAILABLE",
		"daily check-in security protection is temporarily unavailable",
	)
	ErrCheckinEntropyUnavailable = infraerrors.ServiceUnavailable(
		"CHECKIN_ENTROPY_UNAVAILABLE",
		"daily check-in reward generation is temporarily unavailable",
	)
)

type CheckinRecord struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	CheckinDate   time.Time `json:"checkin_date"`
	Mode          string    `json:"mode"`
	RewardType    string    `json:"reward_type"`
	RandomValue   float64   `json:"random_value"`
	RewardAmount  float64   `json:"reward_amount"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	CheckedInAt   time.Time `json:"checked_in_at"`
}

type CheckinUserState struct {
	Role      string
	Status    string
	Balance   float64
	CreatedAt time.Time
}

type CheckinAbuseGuard interface {
	CheckRequest(context.Context, string, int64, time.Duration, int, int) (bool, time.Duration, error)
	CheckAndRecord(context.Context, string, int64, time.Duration, int) (bool, int64, time.Duration, error)
}

type CheckinRepository interface {
	GetUserState(context.Context, int64) (*CheckinUserState, error)
	GetByDate(context.Context, int64, string) (*CheckinRecord, error)
	List(context.Context, int64, int, int) ([]CheckinRecord, int64, error)
	Apply(context.Context, int64, string, string, func(decimal.Decimal) (decimal.Decimal, decimal.Decimal, string, error)) (*CheckinRecord, bool, error)
}

type CheckinConfig struct {
	Enabled                  bool
	NormalEnabled            bool
	LuckyEnabled             bool
	NormalMin                float64
	NormalMax                float64
	LuckyRewardType          string
	LuckyPositiveProbability float64
	LuckyMinMultiply         float64
	LuckyMaxMultiply         float64
	LuckyAmountMin           float64
	LuckyAmountMax           float64
	RiskEnabled              bool
	MinAccountAge            time.Duration
	IPWindow                 time.Duration
	IPMaxUsers               int
}

// AdminCheckinConfig is the editable configuration exposed on the admin page.
// Monetary values remain strings at this boundary so the UI does not round them.
type AdminCheckinConfig struct {
	Enabled                  bool      `json:"enabled"`
	NormalEnabled            bool      `json:"normal_enabled"`
	LuckyEnabled             bool      `json:"lucky_enabled"`
	NormalMin                string    `json:"normal_min"`
	NormalMax                string    `json:"normal_max"`
	LuckyRewardType          string    `json:"lucky_reward_type"`
	LuckyPositiveProbability string    `json:"lucky_positive_probability"`
	LuckyMinMultiply         string    `json:"lucky_min_multiplier"`
	LuckyMaxMultiply         string    `json:"lucky_max_multiplier"`
	LuckyAmountMin           string    `json:"lucky_amount_min"`
	LuckyAmountMax           string    `json:"lucky_amount_max"`
	RiskEnabled              bool      `json:"risk_control_enabled"`
	MinAccountAgeHours       int       `json:"min_account_age_hours"`
	IPWindowMinutes          int       `json:"ip_window_minutes"`
	IPMaxUsers               int       `json:"ip_max_users"`
	ConfigVersion            int64     `json:"config_version"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type AdminCheckinConfigUpdate struct {
	Enabled                  bool
	NormalEnabled            bool
	LuckyEnabled             bool
	NormalMin                string
	NormalMax                string
	LuckyRewardType          string
	LuckyPositiveProbability string
	LuckyMinMultiply         string
	LuckyMaxMultiply         string
	LuckyAmountMin           string
	LuckyAmountMax           string
	RiskEnabled              bool
	MinAccountAgeHours       int
	IPWindowMinutes          int
	IPMaxUsers               int
	ExpectedVersion          int64
	ChangeReason             string
}

type AdminCheckinRecord struct {
	CheckinRecord
	UserEmail string `json:"user_email"`
}

type AdminCheckinRecordFilter struct {
	Page     int
	PageSize int
	Date     string
	UserID   *int64
	Email    string
}

type AdminCheckinOverview struct {
	BusinessDate  string  `json:"business_date"`
	Total         int64   `json:"total"`
	NormalCount   int64   `json:"normal_count"`
	LuckyCount    int64   `json:"lucky_count"`
	PositiveTotal float64 `json:"positive_total"`
	NegativeTotal float64 `json:"negative_total"`
}

type AdminCheckinRepository interface {
	AdminOverview(context.Context, string) (*AdminCheckinOverview, error)
	AdminList(context.Context, AdminCheckinRecordFilter) ([]AdminCheckinRecord, int64, error)
	UpdateConfigIfVersion(context.Context, int64, map[string]string) (bool, error)
}

type CheckinService struct {
	repo         CheckinRepository
	settings     SettingRepository
	cfg          *config.Config
	billingCache BillingCache
	abuseGuard   CheckinAbuseGuard
}

func NewCheckinService(repo CheckinRepository, settings SettingRepository, cfg *config.Config, billingCache BillingCache, abuseGuard CheckinAbuseGuard) *CheckinService {
	return &CheckinService{repo: repo, settings: settings, cfg: cfg, billingCache: billingCache, abuseGuard: abuseGuard}
}

func (s *CheckinService) Status(ctx context.Context, userID int64) (*CheckinStatus, error) {
	now := timezone.Now()
	date := now.Format("2006-01-02")
	state, err := s.repo.GetUserState(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := buildCheckinStatus(now, date)
	result.Timezone = timezone.Name()
	result.Eligible = state.Status == StatusActive && s.cfg != nil && s.cfg.RunMode != config.RunModeSimple
	if !result.Eligible {
		result.UnavailableReason = "not_eligible"
	} else if state.Balance < 0 {
		result.Eligible = false
		result.UnavailableReason = "negative_balance"
	}
	checkinConfig, configErr := s.loadConfig(ctx)
	if configErr != nil {
		result.Enabled = false
		if result.Eligible {
			result.UnavailableReason = "config_invalid"
		}
	} else {
		result.Enabled = checkinConfig.Enabled
		result.NormalEnabled = checkinConfig.NormalEnabled
		result.LuckyEnabled = checkinConfig.LuckyEnabled
		result.LuckyRewardType = checkinConfig.LuckyRewardType
		result.LuckyMinMultiplier = checkinConfig.LuckyMinMultiply
		result.LuckyMaxMultiplier = checkinConfig.LuckyMaxMultiply
		if result.Eligible && !result.Enabled {
			result.UnavailableReason = "disabled"
		} else if result.Eligible && !result.NormalEnabled && !result.LuckyEnabled {
			result.UnavailableReason = "no_modes_enabled"
		}
	}
	if result.Eligible {
		result.TodayRecord, err = s.repo.GetByDate(ctx, userID, date)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		result.CheckedInToday = result.TodayRecord != nil
		if result.CheckedInToday {
			result.CanCheckIn = false
			result.UnavailableReason = "already_checked_in"
		} else if configErr == nil && checkinConfig.RiskEnabled && accountTooNew(state.CreatedAt, now, checkinConfig.MinAccountAge) {
			result.Eligible = false
			result.CanCheckIn = false
			result.UnavailableReason = "account_too_new"
		} else if !result.Enabled {
			if result.UnavailableReason == "" {
				result.UnavailableReason = "disabled"
			}
		} else if !result.NormalEnabled && !result.LuckyEnabled {
			result.CanCheckIn = false
			result.UnavailableReason = "no_modes_enabled"
		} else {
			result.CanCheckIn = result.Enabled
		}
	}
	return result, nil
}

type CheckinStatus struct {
	Enabled            bool           `json:"enabled"`
	NormalEnabled      bool           `json:"normal_enabled"`
	LuckyEnabled       bool           `json:"lucky_enabled"`
	LuckyRewardType    string         `json:"lucky_reward_type"`
	LuckyMinMultiplier float64        `json:"lucky_min_multiplier"`
	LuckyMaxMultiplier float64        `json:"lucky_max_multiplier"`
	Eligible           bool           `json:"eligible"`
	CanCheckIn         bool           `json:"can_check_in"`
	UnavailableReason  string         `json:"unavailable_reason"`
	BusinessDate       string         `json:"business_date"`
	Timezone           string         `json:"timezone"`
	ServerTime         time.Time      `json:"server_time"`
	NextResetAt        time.Time      `json:"next_reset_at"`
	CheckedInToday     bool           `json:"checked_in_today"`
	TodayRecord        *CheckinRecord `json:"today_record,omitempty"`
	DaysInMonth        int            `json:"days_in_month"`
	FirstWeekday       int            `json:"first_weekday"`
}

func (s *CheckinService) CheckIn(ctx context.Context, userID int64, mode, source string) (*CheckinRecord, bool, error) {
	if mode != "normal" && mode != "lucky" {
		return nil, false, ErrCheckinInvalidMode
	}
	normalizedSource := net.ParseIP(strings.TrimSpace(source))
	if normalizedSource == nil || s.abuseGuard == nil {
		return nil, false, ErrCheckinSecurityUnavailable
	}
	allowed, retryAfter, guardErr := s.abuseGuard.CheckRequest(
		ctx,
		normalizedSource.String(),
		userID,
		CheckinRequestWindow,
		CheckinUserRequestLimit,
		CheckinSourceRequestLimit,
	)
	if guardErr != nil {
		return nil, false, ErrCheckinSecurityUnavailable
	}
	if !allowed {
		return nil, false, checkinRateLimitedError(retryAfter)
	}
	state, err := s.repo.GetUserState(ctx, userID)
	if err != nil {
		return nil, false, err
	}
	if state.Status != StatusActive || s.cfg == nil || s.cfg.RunMode == config.RunModeSimple {
		return nil, false, ErrCheckinNotEligible
	}
	if state.Balance < 0 {
		return nil, false, ErrCheckinNegativeBalance
	}
	businessDate := timezone.Now().Format("2006-01-02")
	existing, err := s.repo.GetByDate(ctx, userID, businessDate)
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}
	checkinConfig, err := s.loadConfig(ctx)
	if err != nil {
		return nil, false, err
	}
	if !checkinConfig.Enabled {
		return nil, false, ErrCheckinDisabled
	}
	if (mode == "normal" && !checkinConfig.NormalEnabled) || (mode == "lucky" && !checkinConfig.LuckyEnabled) {
		return nil, false, ErrCheckinModeDisabled.WithMetadata(map[string]string{"mode": mode})
	}
	if checkinConfig.RiskEnabled && accountTooNew(state.CreatedAt, timezone.Now(), checkinConfig.MinAccountAge) {
		return nil, false, ErrCheckinAccountTooNew
	}
	if checkinConfig.RiskEnabled {
		allowed, _, _, guardErr := s.abuseGuard.CheckAndRecord(ctx, normalizedSource.String(), userID, checkinConfig.IPWindow, checkinConfig.IPMaxUsers)
		if guardErr != nil {
			return nil, false, ErrCheckinRiskUnavailable
		}
		if !allowed {
			return nil, false, ErrCheckinSourceLimited
		}
	}
	record, newlyCheckedIn, err := s.repo.Apply(ctx, userID, businessDate, mode, func(balance decimal.Decimal) (decimal.Decimal, decimal.Decimal, string, error) {
		if mode == "normal" {
			value, randomErr := secureRandomBetween(checkinConfig.NormalMin, checkinConfig.NormalMax)
			if randomErr != nil {
				return decimal.Zero, decimal.Zero, "", ErrCheckinEntropyUnavailable.WithCause(randomErr)
			}
			value = roundCheckinValue(value)
			reward := decimal.NewFromFloat(value)
			return reward, reward, CheckinRewardTypeAmount, nil
		}
		if checkinConfig.LuckyRewardType == CheckinRewardTypeAmount {
			value, randomErr := secureRandomSignedBetween(checkinConfig.LuckyAmountMin, checkinConfig.LuckyAmountMax, checkinConfig.LuckyPositiveProbability)
			if randomErr != nil {
				return decimal.Zero, decimal.Zero, "", ErrCheckinEntropyUnavailable.WithCause(randomErr)
			}
			value = roundCheckinValue(value)
			randomValue := decimal.NewFromFloat(value)
			reward := clampCheckinReward(balance, randomValue)
			return reward, randomValue, CheckinRewardTypeAmount, nil
		}
		multiplier, randomErr := secureRandomSignedBetween(checkinConfig.LuckyMinMultiply, checkinConfig.LuckyMaxMultiply, checkinConfig.LuckyPositiveProbability)
		if randomErr != nil {
			return decimal.Zero, decimal.Zero, "", ErrCheckinEntropyUnavailable.WithCause(randomErr)
		}
		multiplier = roundCheckinValue(multiplier)
		randomValue := decimal.NewFromFloat(multiplier)
		reward := balance.Mul(randomValue).Round(CheckinCalculationScale)
		reward = clampCheckinReward(balance, reward)
		return reward, randomValue, CheckinRewardTypeMultiplier, nil
	})
	if err != nil {
		return nil, false, err
	}
	if newlyCheckedIn && s.billingCache != nil {
		// The database transaction is authoritative; cache invalidation is best effort.
		_ = s.billingCache.InvalidateUserBalance(ctx, userID)
	}
	return record, newlyCheckedIn, nil
}

func checkinRateLimitedError(retryAfter time.Duration) error {
	seconds := int(math.Ceil(retryAfter.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	return ErrCheckinRateLimited.WithMetadata(map[string]string{
		"retry_after": strconv.Itoa(seconds),
	})
}

func (s *CheckinService) Records(ctx context.Context, userID int64, page, pageSize int) ([]CheckinRecord, int64, error) {
	return s.repo.List(ctx, userID, page, pageSize)
}

func (s *CheckinService) loadConfig(ctx context.Context) (CheckinConfig, error) {
	if s.settings == nil {
		return CheckinConfig{}, ErrCheckinConfigInvalid
	}
	values, err := s.settings.GetMultiple(ctx, []string{
		SettingKeyCheckinEnabled,
		SettingKeyCheckinNormalEnabled,
		SettingKeyCheckinLuckyEnabled,
		SettingKeyCheckinNormalMin,
		SettingKeyCheckinNormalMax,
		SettingKeyCheckinLuckyRewardType,
		SettingKeyCheckinLuckyPositiveProbability,
		SettingKeyCheckinLuckyMinMultiply,
		SettingKeyCheckinLuckyMaxMultiply,
		SettingKeyCheckinLuckyAmountMin,
		SettingKeyCheckinLuckyAmountMax,
		SettingKeyCheckinRiskEnabled,
		SettingKeyCheckinMinAccountAge,
		SettingKeyCheckinIPWindow,
		SettingKeyCheckinIPMaxUsers,
	})
	if err != nil {
		return CheckinConfig{}, ErrCheckinConfigInvalid
	}
	enabled, enabledErr := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinEnabled]))
	normalEnabled, normalEnabledErr := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinNormalEnabled]))
	luckyEnabled, luckyEnabledErr := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinLuckyEnabled]))
	riskEnabled, riskEnabledErr := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinRiskEnabled]))
	parse := func(key string) (float64, error) {
		value, err := parseCheckinDecimal(values[key])
		if err != nil {
			return 0, err
		}
		return value, nil
	}
	normalMin, err1 := parse(SettingKeyCheckinNormalMin)
	normalMax, err2 := parse(SettingKeyCheckinNormalMax)
	luckyPositive, err3 := parse(SettingKeyCheckinLuckyPositiveProbability)
	luckyMin, err4 := parse(SettingKeyCheckinLuckyMinMultiply)
	luckyMax, err5 := parse(SettingKeyCheckinLuckyMaxMultiply)
	luckyAmountMin, err6 := parse(SettingKeyCheckinLuckyAmountMin)
	luckyAmountMax, err7 := parse(SettingKeyCheckinLuckyAmountMax)
	luckyRewardType := strings.TrimSpace(values[SettingKeyCheckinLuckyRewardType])
	minAccountAgeHours, err8 := parseCheckinInt(values[SettingKeyCheckinMinAccountAge])
	ipWindowMinutes, err9 := parseCheckinInt(values[SettingKeyCheckinIPWindow])
	ipMaxUsers, err10 := parseCheckinInt(values[SettingKeyCheckinIPMaxUsers])
	if enabledErr != nil || normalEnabledErr != nil || luckyEnabledErr != nil || riskEnabledErr != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil ||
		err10 != nil || normalMin < 0 || normalMax < normalMin || normalMax > 100 || !validCheckinLuckyRewardType(luckyRewardType) ||
		luckyPositive < 0 || luckyPositive > 100 || luckyMin < -1 || luckyMin >= 0 || luckyMax <= 0 || luckyMax > 10 ||
		luckyAmountMin < -100 || luckyAmountMin >= 0 || luckyAmountMax <= 0 || luckyAmountMax > 100 ||
		minAccountAgeHours < 0 || minAccountAgeHours > 720 || ipWindowMinutes < 1 || ipWindowMinutes > 1440 || ipMaxUsers < 1 || ipMaxUsers > 10000 {
		return CheckinConfig{}, ErrCheckinConfigInvalid
	}
	return CheckinConfig{
		Enabled:                  enabled,
		NormalEnabled:            normalEnabled,
		LuckyEnabled:             luckyEnabled,
		NormalMin:                normalMin,
		NormalMax:                normalMax,
		LuckyRewardType:          luckyRewardType,
		LuckyPositiveProbability: luckyPositive,
		LuckyMinMultiply:         luckyMin,
		LuckyMaxMultiply:         luckyMax,
		LuckyAmountMin:           luckyAmountMin,
		LuckyAmountMax:           luckyAmountMax,
		RiskEnabled:              riskEnabled,
		MinAccountAge:            time.Duration(minAccountAgeHours) * time.Hour,
		IPWindow:                 time.Duration(ipWindowMinutes) * time.Minute,
		IPMaxUsers:               ipMaxUsers,
	}, nil
}

func validCheckinLuckyRewardType(value string) bool {
	return value == CheckinRewardTypeAmount || value == CheckinRewardTypeMultiplier
}

func parseCheckinInt(raw string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(raw))
}

func accountTooNew(createdAt, now time.Time, minAge time.Duration) bool {
	return minAge > 0 && (createdAt.IsZero() || now.Before(createdAt.Add(minAge)))
}

func parseCheckinDecimal(raw string) (float64, error) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.ContainsAny(value, "eE") {
		return 0, errors.New("check-in values must be finite decimal numbers")
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || (len(parts) == 2 && len(parts[1]) > CheckinCalculationScale) {
		return 0, errors.New("check-in values support at most 2 decimal places")
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, errors.New("check-in values must be finite decimal numbers")
	}
	return parsed, nil
}

func buildCheckinStatus(now time.Time, date string) *CheckinStatus {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	// The UI calendar is Monday-first; Go's Weekday is Sunday-first.
	firstWeekday := (int(start.Weekday()) + 6) % 7
	return &CheckinStatus{BusinessDate: date, ServerTime: now, NextResetAt: time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()), DaysInMonth: next.AddDate(0, 0, -1).Day(), FirstWeekday: firstWeekday}
}

var checkinRandInt = rand.Int

func secureRandomBetween(min, max float64) (float64, error) {
	if max <= min {
		return min, nil
	}
	n, err := checkinRandInt(rand.Reader, big.NewInt(100000001))
	if err != nil {
		return 0, fmt.Errorf("read cryptographic randomness: %w", err)
	}
	return min + (max-min)*float64(n.Int64())/100000000, nil
}

func secureRandomSignedBetween(min, max, positiveProbability float64) (float64, error) {
	const probabilityScale int64 = 100000000
	threshold := int64(math.Round(positiveProbability * float64(probabilityScale)))
	n, err := checkinRandInt(rand.Reader, big.NewInt(100*probabilityScale))
	if err != nil {
		return 0, fmt.Errorf("read cryptographic randomness: %w", err)
	}
	positive := n.Int64() < threshold
	if positive {
		value, randomErr := secureRandomBetween(0, max)
		if randomErr != nil {
			return 0, randomErr
		}
		value = roundCheckinValue(value)
		if value <= 0 {
			return 0.01, nil
		}
		return value, nil
	}
	value, randomErr := secureRandomBetween(min, 0)
	if randomErr != nil {
		return 0, randomErr
	}
	value = roundCheckinValue(value)
	if value >= 0 {
		return -0.01, nil
	}
	return value, nil
}

func clampCheckinReward(balance, reward decimal.Decimal) decimal.Decimal {
	if balance.Add(reward).IsNegative() {
		return balance.Truncate(CheckinCalculationScale).Neg()
	}
	return reward
}

func roundCheckinValue(value float64) float64 {
	return math.Round(value*100) / 100
}
