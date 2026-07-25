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
)

const (
	SettingKeyCheckinEnabled          = "checkin_enabled"
	SettingKeyCheckinNormalMin        = "checkin_normal_min"
	SettingKeyCheckinNormalMax        = "checkin_normal_max"
	SettingKeyCheckinLuckyMinMultiply = "checkin_lucky_min_multiplier"
	SettingKeyCheckinLuckyMaxMultiply = "checkin_lucky_max_multiplier"
	SettingKeyCheckinRiskEnabled      = "checkin_risk_control_enabled"
	SettingKeyCheckinMinAccountAge    = "checkin_min_account_age_hours"
	SettingKeyCheckinIPWindow         = "checkin_ip_window_minutes"
	SettingKeyCheckinIPMaxUsers       = "checkin_ip_max_users"
	SettingKeyCheckinConfigVersion    = "checkin_config_version"

	// These hard safety ceilings are intentionally separate from the editable
	// business-risk settings so disabling campaign risk controls cannot disable
	// endpoint protection.
	CheckinRequestWindow      = time.Minute
	CheckinUserRequestLimit   = 10
	CheckinSourceRequestLimit = 60
)

var (
	ErrCheckinDisabled            = infraerrors.Forbidden("CHECKIN_DISABLED", "daily check-in is currently unavailable")
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
	Apply(context.Context, int64, string, string, func(float64) (float64, float64, error)) (*CheckinRecord, bool, error)
}

type CheckinConfig struct {
	Enabled          bool
	NormalMin        float64
	NormalMax        float64
	LuckyMinMultiply float64
	LuckyMaxMultiply float64
	RiskEnabled      bool
	MinAccountAge    time.Duration
	IPWindow         time.Duration
	IPMaxUsers       int
}

// AdminCheckinConfig is the editable configuration exposed on the admin page.
// Monetary values remain strings at this boundary so the UI does not round them.
type AdminCheckinConfig struct {
	Enabled            bool      `json:"enabled"`
	NormalMin          string    `json:"normal_min"`
	NormalMax          string    `json:"normal_max"`
	LuckyMinMultiply   string    `json:"lucky_min_multiplier"`
	LuckyMaxMultiply   string    `json:"lucky_max_multiplier"`
	RiskEnabled        bool      `json:"risk_control_enabled"`
	MinAccountAgeHours int       `json:"min_account_age_hours"`
	IPWindowMinutes    int       `json:"ip_window_minutes"`
	IPMaxUsers         int       `json:"ip_max_users"`
	ConfigVersion      int64     `json:"config_version"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type AdminCheckinConfigUpdate struct {
	Enabled            bool
	NormalMin          string
	NormalMax          string
	LuckyMinMultiply   string
	LuckyMaxMultiply   string
	RiskEnabled        bool
	MinAccountAgeHours int
	IPWindowMinutes    int
	IPMaxUsers         int
	ExpectedVersion    int64
	ChangeReason       string
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
	result.Eligible = state.Role != RoleAdmin && state.Status == StatusActive && s.cfg != nil && s.cfg.RunMode != config.RunModeSimple
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
		if result.Eligible && !result.Enabled {
			result.UnavailableReason = "disabled"
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
		} else if !result.Enabled && result.UnavailableReason == "" {
			result.UnavailableReason = "disabled"
		} else {
			result.CanCheckIn = result.Enabled
		}
	}
	return result, nil
}

type CheckinStatus struct {
	Enabled           bool           `json:"enabled"`
	Eligible          bool           `json:"eligible"`
	CanCheckIn        bool           `json:"can_check_in"`
	UnavailableReason string         `json:"unavailable_reason"`
	BusinessDate      string         `json:"business_date"`
	Timezone          string         `json:"timezone"`
	ServerTime        time.Time      `json:"server_time"`
	NextResetAt       time.Time      `json:"next_reset_at"`
	CheckedInToday    bool           `json:"checked_in_today"`
	TodayRecord       *CheckinRecord `json:"today_record,omitempty"`
	DaysInMonth       int            `json:"days_in_month"`
	FirstWeekday      int            `json:"first_weekday"`
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
	if state.Role == RoleAdmin || state.Status != StatusActive || s.cfg == nil || s.cfg.RunMode == config.RunModeSimple {
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
	record, newlyCheckedIn, err := s.repo.Apply(ctx, userID, businessDate, mode, func(balance float64) (float64, float64, error) {
		if mode == "normal" {
			value, randomErr := secureRandomBetween(checkinConfig.NormalMin, checkinConfig.NormalMax)
			if randomErr != nil {
				return 0, 0, ErrCheckinEntropyUnavailable.WithCause(randomErr)
			}
			value = round8(value)
			return value, value, nil
		}
		multiplier, randomErr := secureRandomBetween(checkinConfig.LuckyMinMultiply, checkinConfig.LuckyMaxMultiply)
		if randomErr != nil {
			return 0, 0, ErrCheckinEntropyUnavailable.WithCause(randomErr)
		}
		multiplier = round8(multiplier)
		reward := round8(balance * multiplier)
		if balance+reward < 0 {
			reward = -balance
		}
		return reward, multiplier, nil
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
		SettingKeyCheckinNormalMin,
		SettingKeyCheckinNormalMax,
		SettingKeyCheckinLuckyMinMultiply,
		SettingKeyCheckinLuckyMaxMultiply,
		SettingKeyCheckinRiskEnabled,
		SettingKeyCheckinMinAccountAge,
		SettingKeyCheckinIPWindow,
		SettingKeyCheckinIPMaxUsers,
	})
	if err != nil {
		return CheckinConfig{}, ErrCheckinConfigInvalid
	}
	enabled, enabledErr := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinEnabled]))
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
	luckyMin, err3 := parse(SettingKeyCheckinLuckyMinMultiply)
	luckyMax, err4 := parse(SettingKeyCheckinLuckyMaxMultiply)
	minAccountAgeHours, err5 := parseCheckinInt(values[SettingKeyCheckinMinAccountAge])
	ipWindowMinutes, err6 := parseCheckinInt(values[SettingKeyCheckinIPWindow])
	ipMaxUsers, err7 := parseCheckinInt(values[SettingKeyCheckinIPMaxUsers])
	if enabledErr != nil || riskEnabledErr != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil ||
		normalMin < 0 || normalMax < normalMin || normalMax > 100 || luckyMin < -1 || luckyMax < luckyMin || luckyMax > 1 ||
		minAccountAgeHours < 0 || minAccountAgeHours > 720 || ipWindowMinutes < 1 || ipWindowMinutes > 1440 || ipMaxUsers < 1 || ipMaxUsers > 10000 {
		return CheckinConfig{}, ErrCheckinConfigInvalid
	}
	return CheckinConfig{
		Enabled:          enabled,
		NormalMin:        normalMin,
		NormalMax:        normalMax,
		LuckyMinMultiply: luckyMin,
		LuckyMaxMultiply: luckyMax,
		RiskEnabled:      riskEnabled,
		MinAccountAge:    time.Duration(minAccountAgeHours) * time.Hour,
		IPWindow:         time.Duration(ipWindowMinutes) * time.Minute,
		IPMaxUsers:       ipMaxUsers,
	}, nil
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
	if len(parts) > 2 || (len(parts) == 2 && len(parts[1]) > 8) {
		return 0, errors.New("check-in values support at most 8 decimal places")
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

func round8(value float64) float64 {
	return math.Round(value*100000000) / 100000000
}
