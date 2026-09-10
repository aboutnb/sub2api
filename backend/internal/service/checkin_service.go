package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/shopspring/decimal"
)

const (
	SettingKeyCheckinEnabled                      = "checkin_enabled"
	SettingKeyCheckinNormalEnabled                = "checkin_normal_enabled"
	SettingKeyCheckinLuckyEnabled                 = "checkin_lucky_enabled"
	SettingKeyCheckinNormalMin                    = "checkin_normal_min"
	SettingKeyCheckinNormalMax                    = "checkin_normal_max"
	SettingKeyCheckinLuckyRewardType              = "checkin_lucky_reward_type"
	SettingKeyCheckinLuckyPositiveProbability     = "checkin_lucky_positive_probability"
	SettingKeyCheckinLuckyMultiplierPositiveTiers = "checkin_lucky_multiplier_positive_tiers"
	SettingKeyCheckinLuckyAmountPositiveTiers     = "checkin_lucky_amount_positive_tiers"
	SettingKeyCheckinLuckyMinMultiply             = "checkin_lucky_min_multiplier"
	SettingKeyCheckinLuckyMaxMultiply             = "checkin_lucky_max_multiplier"
	SettingKeyCheckinLuckyAmountMin               = "checkin_lucky_amount_min"
	SettingKeyCheckinLuckyAmountMax               = "checkin_lucky_amount_max"
	SettingKeyCheckinRiskEnabled                  = "checkin_risk_control_enabled"
	SettingKeyCheckinMinAccountAge                = "checkin_min_account_age_hours"
	SettingKeyCheckinIPWindow                     = "checkin_ip_window_minutes"
	SettingKeyCheckinIPMaxUsers                   = "checkin_ip_max_users"
	SettingKeyCheckinFingerprintWindow            = "checkin_fingerprint_window_minutes"
	SettingKeyCheckinFingerprintMaxUsers          = "checkin_fingerprint_max_users"
	SettingKeyCheckinUnrechargedEnabled           = "checkin_unrecharged_reduction_enabled"
	SettingKeyCheckinUnrechargedThreshold         = "checkin_unrecharged_checkin_threshold"
	SettingKeyCheckinUnrechargedNormalPercent     = "checkin_unrecharged_normal_reward_percent"
	SettingKeyCheckinTurnstileEnabled             = "checkin_turnstile_enabled"
	SettingKeyCheckinConfigVersion                = "checkin_config_version"

	// These hard safety ceilings are intentionally separate from the editable
	// business-risk settings so disabling campaign risk controls cannot disable
	// endpoint protection.
	CheckinRequestWindow      = time.Minute
	CheckinUserRequestLimit   = 10
	CheckinSourceRequestLimit = 60
	CheckinUserAgentMaxBytes  = 512

	// These minimum source limits always remain active. The editable risk
	// settings can add stricter checks, but must not disable bulk protection.
	CheckinHardSourceWindow        = 10 * time.Minute
	CheckinHardSourceMaxUsers      = 5
	CheckinHardFingerprintWindow   = 24 * time.Hour
	CheckinHardFingerprintMaxUsers = 1

	CheckinRewardTypeAmount     = "amount"
	CheckinRewardTypeMultiplier = "multiplier"
	CheckinCalculationScale     = 2
	MaxCheckinPositiveTiers     = 10
)

var (
	ErrCheckinDisabled            = infraerrors.Forbidden("CHECKIN_DISABLED", "daily check-in is currently unavailable")
	ErrCheckinModeDisabled        = infraerrors.Forbidden("CHECKIN_MODE_DISABLED", "this check-in mode is currently unavailable")
	ErrCheckinNotEligible         = infraerrors.Forbidden("CHECKIN_NOT_ELIGIBLE", "this account cannot use daily check-in")
	ErrCheckinInvalidMode         = infraerrors.BadRequest("CHECKIN_INVALID_MODE", "mode must be normal or lucky")
	ErrCheckinConfigInvalid       = infraerrors.InternalServer("CHECKIN_CONFIG_INVALID", "daily check-in configuration is invalid")
	ErrCheckinUserNotFound        = infraerrors.NotFound("USER_NOT_FOUND", "user not found")
	ErrCheckinAccountTooNew       = infraerrors.Forbidden("CHECKIN_ACCOUNT_TOO_NEW", "account is too new for daily check-in")
	ErrCheckinGrantRestricted     = infraerrors.Forbidden("CHECKIN_GRANT_RESTRICTED", "this account must recharge before using daily check-in")
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
	ErrCheckinTurnstileRequired = infraerrors.BadRequest(
		"CHECKIN_TURNSTILE_REQUIRED",
		"daily check-in verification is required",
	)
	ErrCheckinTurnstileFailed = infraerrors.BadRequest(
		"CHECKIN_TURNSTILE_FAILED",
		"daily check-in verification failed",
	)
	ErrCheckinTurnstileUnavailable = infraerrors.ServiceUnavailable(
		"CHECKIN_TURNSTILE_UNAVAILABLE",
		"daily check-in verification is temporarily unavailable",
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
	Role                  string
	Status                string
	Balance               float64
	CreatedAt             time.Time
	HasRecharge           bool
	SignupGrantRestricted bool
}

type CheckinSettlementState struct {
	Balance      decimal.Decimal
	CheckinCount int64
	HasRecharge  bool
}

type CheckinAbuseGuard interface {
	CheckRequest(context.Context, string, int64, time.Duration, int, int) (bool, time.Duration, error)
	CheckAndRecord(context.Context, string, int64, time.Duration, int) (bool, int64, time.Duration, error)
}

// CheckinMultiSourceAbuseGuard atomically applies independent limits to more
// than one source identity (for example IP and IP+User-Agent). Implementations
// must not retain a user in any source set when another source is rejected.
type CheckinMultiSourceAbuseGuard interface {
	CheckAndRecordSources(context.Context, []CheckinSourceLimit, int64) (bool, int64, time.Duration, error)
}

type CheckinSourceLimit struct {
	Source   string
	Window   time.Duration
	MaxUsers int
}

// CheckinIdentity contains only request metadata used by the abuse guard. The
// raw values are never persisted; the repository hashes them with its server
// secret before writing Redis.
type CheckinIdentity struct {
	IP        string
	UserAgent string
}

type CheckinRepository interface {
	GetUserState(context.Context, int64) (*CheckinUserState, error)
	GetByDate(context.Context, int64, string) (*CheckinRecord, error)
	List(context.Context, int64, int, int) ([]CheckinRecord, int64, error)
	Apply(context.Context, int64, string, string, func(CheckinSettlementState) (decimal.Decimal, decimal.Decimal, string, error)) (*CheckinRecord, bool, error)
}

type CheckinConfig struct {
	Enabled                      bool
	NormalEnabled                bool
	LuckyEnabled                 bool
	NormalMin                    float64
	NormalMax                    float64
	LuckyRewardType              string
	LuckyPositiveProbability     float64
	LuckyMultiplierPositiveTiers []CheckinPositiveTier
	LuckyAmountPositiveTiers     []CheckinPositiveTier
	LuckyMinMultiply             float64
	LuckyMaxMultiply             float64
	LuckyAmountMin               float64
	LuckyAmountMax               float64
	RiskEnabled                  bool
	MinAccountAge                time.Duration
	IPWindow                     time.Duration
	IPMaxUsers                   int
	FingerprintWindow            time.Duration
	FingerprintMaxUsers          int
	UnrechargedEnabled           bool
	UnrechargedThreshold         int64
	UnrechargedNormalPercent     float64
	TurnstileEnabled             bool
	TurnstileSiteKey             string
}

type CheckinPositiveTier struct {
	MinStep     int64
	MaxStep     int64
	WeightUnits int64
}

type AdminCheckinPositiveTier struct {
	Min    string `json:"min"`
	Max    string `json:"max"`
	Weight string `json:"weight"`
}

// AdminCheckinConfig is the editable configuration exposed on the admin page.
// Monetary values remain strings at this boundary so the UI does not round them.
type AdminCheckinConfig struct {
	Enabled                      bool                       `json:"enabled"`
	NormalEnabled                bool                       `json:"normal_enabled"`
	LuckyEnabled                 bool                       `json:"lucky_enabled"`
	NormalMin                    string                     `json:"normal_min"`
	NormalMax                    string                     `json:"normal_max"`
	LuckyRewardType              string                     `json:"lucky_reward_type"`
	LuckyPositiveProbability     string                     `json:"lucky_positive_probability"`
	LuckyMultiplierPositiveTiers []AdminCheckinPositiveTier `json:"lucky_multiplier_positive_tiers"`
	LuckyAmountPositiveTiers     []AdminCheckinPositiveTier `json:"lucky_amount_positive_tiers"`
	LuckyMinMultiply             string                     `json:"lucky_min_multiplier"`
	LuckyMaxMultiply             string                     `json:"lucky_max_multiplier"`
	LuckyAmountMin               string                     `json:"lucky_amount_min"`
	LuckyAmountMax               string                     `json:"lucky_amount_max"`
	RiskEnabled                  bool                       `json:"risk_control_enabled"`
	MinAccountAgeHours           int                        `json:"min_account_age_hours"`
	IPWindowMinutes              int                        `json:"ip_window_minutes"`
	IPMaxUsers                   int                        `json:"ip_max_users"`
	FingerprintWindowMinutes     int                        `json:"fingerprint_window_minutes"`
	FingerprintMaxUsers          int                        `json:"fingerprint_max_users"`
	UnrechargedEnabled           bool                       `json:"unrecharged_reduction_enabled"`
	UnrechargedCheckinThreshold  int                        `json:"unrecharged_checkin_threshold"`
	UnrechargedNormalPercent     string                     `json:"unrecharged_normal_reward_percent"`
	TurnstileEnabled             bool                       `json:"turnstile_enabled"`
	ConfigVersion                int64                      `json:"config_version"`
	UpdatedAt                    time.Time                  `json:"updated_at"`
}

type AdminCheckinConfigUpdate struct {
	Enabled                      bool
	NormalEnabled                bool
	LuckyEnabled                 bool
	NormalMin                    string
	NormalMax                    string
	LuckyRewardType              string
	LuckyPositiveProbability     string
	LuckyMultiplierPositiveTiers []AdminCheckinPositiveTier
	LuckyAmountPositiveTiers     []AdminCheckinPositiveTier
	LuckyMinMultiply             string
	LuckyMaxMultiply             string
	LuckyAmountMin               string
	LuckyAmountMax               string
	RiskEnabled                  bool
	MinAccountAgeHours           int
	IPWindowMinutes              int
	IPMaxUsers                   int
	FingerprintWindowMinutes     int
	FingerprintMaxUsers          int
	UnrechargedEnabled           bool
	UnrechargedCheckinThreshold  int
	UnrechargedNormalPercent     string
	TurnstileEnabled             *bool
	ExpectedVersion              int64
	ChangeReason                 string
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
	turnstile    CheckinTurnstileVerifier
}

// CheckinTurnstileVerifier keeps the check-in flow testable while the live
// application uses the shared TurnstileService.
type CheckinTurnstileVerifier interface {
	VerifyTokenWithSecret(context.Context, string, string, string) error
}

func NewCheckinService(repo CheckinRepository, settings SettingRepository, cfg *config.Config, billingCache BillingCache, abuseGuard CheckinAbuseGuard) *CheckinService {
	return &CheckinService{repo: repo, settings: settings, cfg: cfg, billingCache: billingCache, abuseGuard: abuseGuard}
}

func ProvideCheckinService(repo CheckinRepository, settings SettingRepository, cfg *config.Config, billingCache BillingCache, abuseGuard CheckinAbuseGuard, turnstile *TurnstileService) *CheckinService {
	svc := NewCheckinService(repo, settings, cfg, billingCache, abuseGuard)
	svc.SetTurnstileService(turnstile)
	return svc
}

func (s *CheckinService) SetTurnstileService(turnstile CheckinTurnstileVerifier) {
	s.turnstile = turnstile
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
	if result.Eligible && state.SignupGrantRestricted && !state.HasRecharge {
		result.Eligible = false
		result.CanCheckIn = false
		result.UnavailableReason = "grant_restricted"
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
		result.TurnstileEnabled = checkinConfig.TurnstileEnabled
		result.TurnstileSiteKey = checkinConfig.TurnstileSiteKey
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
		if result.CanCheckIn && result.TurnstileEnabled && result.TurnstileSiteKey == "" {
			result.CanCheckIn = false
			result.UnavailableReason = "turnstile_not_configured"
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
	TurnstileEnabled   bool           `json:"turnstile_enabled"`
	TurnstileSiteKey   string         `json:"turnstile_site_key,omitempty"`
}

// CheckIn preserves the service API used by non-HTTP callers. HTTP handlers
// should use CheckInWithIdentity so the IP+User-Agent fingerprint is enforced.
func (s *CheckinService) CheckIn(ctx context.Context, userID int64, mode, source string) (*CheckinRecord, bool, error) {
	return s.CheckInWithIdentity(ctx, userID, mode, CheckinIdentity{IP: source})
}

func (s *CheckinService) CheckInWithIdentity(ctx context.Context, userID int64, mode string, identity CheckinIdentity) (*CheckinRecord, bool, error) {
	return s.checkInWithIdentityAndCaptcha(ctx, userID, mode, identity, "", false)
}

// CheckInWithIdentityAndCaptcha is the HTTP-facing check-in method. The
// verification token is only required for a new settlement; an existing
// settlement is returned idempotently without consuming another token.
func (s *CheckinService) CheckInWithIdentityAndCaptcha(ctx context.Context, userID int64, mode string, identity CheckinIdentity, token string) (*CheckinRecord, bool, error) {
	return s.checkInWithIdentityAndCaptcha(ctx, userID, mode, identity, token, true)
}

func (s *CheckinService) checkInWithIdentityAndCaptcha(ctx context.Context, userID int64, mode string, identity CheckinIdentity, token string, enforceCaptcha bool) (*CheckinRecord, bool, error) {
	if mode != "normal" && mode != "lucky" {
		return nil, false, ErrCheckinInvalidMode
	}
	normalizedSource := net.ParseIP(strings.TrimSpace(identity.IP))
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
	if state.SignupGrantRestricted && !state.HasRecharge {
		return nil, false, ErrCheckinGrantRestricted
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
	sourceRiskConfig := effectiveCheckinSourceRiskConfig(checkinConfig)
	sourceAllowed, _, _, guardErr := s.checkinSourceRisk(ctx, userID, normalizedSource.String(), identity.UserAgent, sourceRiskConfig)
	if guardErr != nil {
		return nil, false, ErrCheckinRiskUnavailable
	}
	if !sourceAllowed {
		return nil, false, ErrCheckinSourceLimited
	}
	if enforceCaptcha && checkinConfig.TurnstileEnabled {
		if err := s.verifyCheckinTurnstile(ctx, token, normalizedSource.String()); err != nil {
			return nil, false, err
		}
	}
	record, newlyCheckedIn, err := s.repo.Apply(ctx, userID, businessDate, mode, func(settlement CheckinSettlementState) (decimal.Decimal, decimal.Decimal, string, error) {
		balance := settlement.Balance
		if mode == "normal" {
			value, randomErr := secureRandomBetween(checkinConfig.NormalMin, checkinConfig.NormalMax)
			if randomErr != nil {
				return decimal.Zero, decimal.Zero, "", ErrCheckinEntropyUnavailable.WithCause(randomErr)
			}
			value = roundCheckinValue(value)
			reward := decimal.NewFromFloat(value)
			if checkinConfig.UnrechargedEnabled && !settlement.HasRecharge && settlement.CheckinCount >= checkinConfig.UnrechargedThreshold {
				reward = reward.Mul(decimal.NewFromFloat(checkinConfig.UnrechargedNormalPercent / 100)).Round(CheckinCalculationScale)
			}
			return reward, reward, CheckinRewardTypeAmount, nil
		}
		if checkinConfig.LuckyRewardType == CheckinRewardTypeAmount {
			value, randomErr := secureRandomTieredSignedBetween(checkinConfig.LuckyAmountMin, checkinConfig.LuckyAmountPositiveTiers, checkinConfig.LuckyPositiveProbability)
			if randomErr != nil {
				return decimal.Zero, decimal.Zero, "", ErrCheckinEntropyUnavailable.WithCause(randomErr)
			}
			value = roundCheckinValue(value)
			randomValue := decimal.NewFromFloat(value)
			reward := clampCheckinReward(balance, randomValue)
			return reward, randomValue, CheckinRewardTypeAmount, nil
		}
		multiplier, randomErr := secureRandomTieredSignedBetween(checkinConfig.LuckyMinMultiply, checkinConfig.LuckyMultiplierPositiveTiers, checkinConfig.LuckyPositiveProbability)
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

func (s *CheckinService) verifyCheckinTurnstile(ctx context.Context, token, remoteIP string) error {
	if strings.TrimSpace(token) == "" {
		return ErrCheckinTurnstileRequired
	}
	if s.turnstile == nil || s.settings == nil {
		return ErrCheckinTurnstileUnavailable
	}
	secretKey, err := s.settings.GetValue(ctx, SettingKeyTurnstileSecretKey)
	if err != nil || strings.TrimSpace(secretKey) == "" {
		return ErrCheckinTurnstileUnavailable
	}
	if err := s.turnstile.VerifyTokenWithSecret(ctx, strings.TrimSpace(secretKey), token, remoteIP); err != nil {
		if errors.Is(err, ErrTurnstileVerificationFailed) {
			return ErrCheckinTurnstileFailed
		}
		return ErrCheckinTurnstileUnavailable
	}
	return nil
}

func (s *CheckinService) checkinSourceRisk(ctx context.Context, userID int64, normalizedIP, rawUserAgent string, cfg CheckinConfig) (bool, int64, time.Duration, error) {
	userAgent := normalizeCheckinUserAgent(rawUserAgent)
	if multiGuard, ok := s.abuseGuard.(CheckinMultiSourceAbuseGuard); ok && userAgent != "" {
		fingerprintSource := normalizedIP + "\x00" + userAgent
		return multiGuard.CheckAndRecordSources(
			ctx,
			[]CheckinSourceLimit{
				{Source: normalizedIP, Window: cfg.IPWindow, MaxUsers: cfg.IPMaxUsers},
				{Source: fingerprintSource, Window: cfg.FingerprintWindow, MaxUsers: cfg.FingerprintMaxUsers},
			},
			userID,
		)
	}
	return s.abuseGuard.CheckAndRecord(ctx, normalizedIP, userID, cfg.IPWindow, cfg.IPMaxUsers)
}

func hardCheckinSourceRiskConfig() CheckinConfig {
	return CheckinConfig{
		IPWindow:            CheckinHardSourceWindow,
		IPMaxUsers:          CheckinHardSourceMaxUsers,
		FingerprintWindow:   CheckinHardFingerprintWindow,
		FingerprintMaxUsers: CheckinHardFingerprintMaxUsers,
	}
}

func effectiveCheckinSourceRiskConfig(cfg CheckinConfig) CheckinConfig {
	hard := hardCheckinSourceRiskConfig()
	if !cfg.RiskEnabled {
		return hard
	}
	if cfg.IPWindow > hard.IPWindow {
		hard.IPWindow = cfg.IPWindow
	}
	if cfg.IPMaxUsers < hard.IPMaxUsers {
		hard.IPMaxUsers = cfg.IPMaxUsers
	}
	if cfg.FingerprintWindow > hard.FingerprintWindow {
		hard.FingerprintWindow = cfg.FingerprintWindow
	}
	if cfg.FingerprintMaxUsers < hard.FingerprintMaxUsers {
		hard.FingerprintMaxUsers = cfg.FingerprintMaxUsers
	}
	return hard
}

func normalizeCheckinUserAgent(value string) string {
	value = strings.TrimSpace(strings.ToValidUTF8(value, ""))
	if len(value) <= CheckinUserAgentMaxBytes {
		return value
	}
	value = value[:CheckinUserAgentMaxBytes]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
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
		SettingKeyCheckinLuckyMultiplierPositiveTiers,
		SettingKeyCheckinLuckyAmountPositiveTiers,
		SettingKeyCheckinLuckyMinMultiply,
		SettingKeyCheckinLuckyMaxMultiply,
		SettingKeyCheckinLuckyAmountMin,
		SettingKeyCheckinLuckyAmountMax,
		SettingKeyCheckinRiskEnabled,
		SettingKeyCheckinMinAccountAge,
		SettingKeyCheckinIPWindow,
		SettingKeyCheckinIPMaxUsers,
		SettingKeyCheckinFingerprintWindow,
		SettingKeyCheckinFingerprintMaxUsers,
		SettingKeyCheckinUnrechargedEnabled,
		SettingKeyCheckinUnrechargedThreshold,
		SettingKeyCheckinUnrechargedNormalPercent,
		SettingKeyCheckinTurnstileEnabled,
		SettingKeyTurnstileSiteKey,
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
	fingerprintWindowMinutes, err15 := parseCheckinInt(values[SettingKeyCheckinFingerprintWindow])
	fingerprintMaxUsers, err16 := parseCheckinInt(values[SettingKeyCheckinFingerprintMaxUsers])
	unrechargedEnabled, unrechargedEnabledErr := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinUnrechargedEnabled]))
	unrechargedThreshold, err13 := parseCheckinInt(values[SettingKeyCheckinUnrechargedThreshold])
	unrechargedNormalPercent, err14 := parse(SettingKeyCheckinUnrechargedNormalPercent)
	turnstileEnabled, turnstileEnabledErr := parseOptionalCheckinBool(values[SettingKeyCheckinTurnstileEnabled])
	multiplierPositiveTiers, _, err11 := parseStoredCheckinPositiveTiers(values[SettingKeyCheckinLuckyMultiplierPositiveTiers], luckyMax)
	amountPositiveTiers, _, err12 := parseStoredCheckinPositiveTiers(values[SettingKeyCheckinLuckyAmountPositiveTiers], luckyAmountMax)
	if enabledErr != nil || normalEnabledErr != nil || luckyEnabledErr != nil || riskEnabledErr != nil || unrechargedEnabledErr != nil || turnstileEnabledErr != nil || err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil ||
		err10 != nil || err11 != nil || err12 != nil || err13 != nil || err14 != nil || err15 != nil || err16 != nil || normalMin < 0 || normalMax < normalMin || normalMax > 100 || !validCheckinLuckyRewardType(luckyRewardType) ||
		luckyPositive < 0 || luckyPositive > 100 || luckyMin < -1 || luckyMin >= 0 || luckyMax <= 0 || luckyMax > 10 ||
		luckyAmountMin < -100 || luckyAmountMin >= 0 || luckyAmountMax <= 0 || luckyAmountMax > 100 ||
		minAccountAgeHours < 0 || minAccountAgeHours > 720 || ipWindowMinutes < 1 || ipWindowMinutes > 1440 || ipMaxUsers < 1 || ipMaxUsers > 10000 || fingerprintWindowMinutes < 1 || fingerprintWindowMinutes > 10080 || fingerprintMaxUsers < 1 || fingerprintMaxUsers > 100 ||
		unrechargedThreshold < 1 || unrechargedThreshold > 3650 || unrechargedNormalPercent < 0 || unrechargedNormalPercent > 100 {
		return CheckinConfig{}, ErrCheckinConfigInvalid
	}
	return CheckinConfig{
		Enabled:                      enabled,
		NormalEnabled:                normalEnabled,
		LuckyEnabled:                 luckyEnabled,
		NormalMin:                    normalMin,
		NormalMax:                    normalMax,
		LuckyRewardType:              luckyRewardType,
		LuckyPositiveProbability:     luckyPositive,
		LuckyMultiplierPositiveTiers: multiplierPositiveTiers,
		LuckyAmountPositiveTiers:     amountPositiveTiers,
		LuckyMinMultiply:             luckyMin,
		LuckyMaxMultiply:             luckyMax,
		LuckyAmountMin:               luckyAmountMin,
		LuckyAmountMax:               luckyAmountMax,
		RiskEnabled:                  riskEnabled,
		MinAccountAge:                time.Duration(minAccountAgeHours) * time.Hour,
		IPWindow:                     time.Duration(ipWindowMinutes) * time.Minute,
		IPMaxUsers:                   ipMaxUsers,
		FingerprintWindow:            time.Duration(fingerprintWindowMinutes) * time.Minute,
		FingerprintMaxUsers:          fingerprintMaxUsers,
		UnrechargedEnabled:           unrechargedEnabled,
		UnrechargedThreshold:         int64(unrechargedThreshold),
		UnrechargedNormalPercent:     unrechargedNormalPercent,
		TurnstileEnabled:             turnstileEnabled,
		TurnstileSiteKey:             strings.TrimSpace(values[SettingKeyTurnstileSiteKey]),
	}, nil
}

func validCheckinLuckyRewardType(value string) bool {
	return value == CheckinRewardTypeAmount || value == CheckinRewardTypeMultiplier
}

func parseCheckinInt(raw string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(raw))
}

func parseOptionalCheckinBool(raw string) (bool, error) {
	if strings.TrimSpace(raw) == "" {
		return false, nil
	}
	return strconv.ParseBool(strings.TrimSpace(raw))
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

func parseStoredCheckinPositiveTiers(raw string, positiveMax float64) ([]CheckinPositiveTier, []AdminCheckinPositiveTier, error) {
	var tiers []AdminCheckinPositiveTier
	value := strings.TrimSpace(raw)
	if value != "" && value != "[]" {
		if err := json.Unmarshal([]byte(value), &tiers); err != nil {
			return nil, nil, err
		}
	}
	return normalizeCheckinPositiveTiers(tiers, positiveMax)
}

func normalizeCheckinPositiveTiers(input []AdminCheckinPositiveTier, positiveMax float64) ([]CheckinPositiveTier, []AdminCheckinPositiveTier, error) {
	maxStep := int64(math.Round(positiveMax * 100))
	if maxStep < 1 || len(input) > MaxCheckinPositiveTiers {
		return nil, nil, errors.New("invalid positive check-in tier range")
	}
	if len(input) == 0 {
		input = []AdminCheckinPositiveTier{{Min: "0.01", Max: formatCheckinStep(maxStep), Weight: "100"}}
	}

	tiers := make([]CheckinPositiveTier, 0, len(input))
	normalized := make([]AdminCheckinPositiveTier, 0, len(input))
	var previousMax int64
	for index, item := range input {
		minValue, minErr := parseCheckinDecimal(item.Min)
		maxValue, maxErr := parseCheckinDecimal(item.Max)
		weightValue, weightErr := parseCheckinDecimal(item.Weight)
		minStep := int64(math.Round(minValue * 100))
		itemMaxStep := int64(math.Round(maxValue * 100))
		weightUnits := int64(math.Round(weightValue * 100))
		expectedMin := int64(1)
		if index > 0 {
			expectedMin = previousMax + 1
		}
		if minErr != nil || maxErr != nil || weightErr != nil || minStep != expectedMin || itemMaxStep < minStep || itemMaxStep > maxStep || weightUnits < 1 || weightValue > 10000 {
			return nil, nil, errors.New("invalid positive check-in tier")
		}
		tiers = append(tiers, CheckinPositiveTier{MinStep: minStep, MaxStep: itemMaxStep, WeightUnits: weightUnits})
		normalized = append(normalized, AdminCheckinPositiveTier{Min: formatCheckinStep(minStep), Max: formatCheckinStep(itemMaxStep), Weight: formatCheckinWeight(weightUnits)})
		previousMax = itemMaxStep
	}
	if previousMax != maxStep {
		return nil, nil, errors.New("positive check-in tiers must cover the configured range")
	}
	return tiers, normalized, nil
}

func formatCheckinStep(step int64) string {
	return fmt.Sprintf("%.2f", float64(step)/100)
}

func formatCheckinWeight(units int64) string {
	return strconv.FormatFloat(float64(units)/100, 'f', -1, 64)
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

func secureRandomTieredSignedBetween(min float64, positiveTiers []CheckinPositiveTier, positiveProbability float64) (float64, error) {
	positive, err := secureRandomPositiveDirection(positiveProbability)
	if err != nil {
		return 0, err
	}
	if positive {
		return secureRandomPositiveTierStep(positiveTiers)
	}
	value, randomErr := secureRandomHundredthStep(math.Abs(min))
	return -value, randomErr
}

func secureRandomPositiveDirection(positiveProbability float64) (bool, error) {
	const probabilityScale int64 = 100000000
	threshold := int64(math.Round(positiveProbability * float64(probabilityScale)))
	n, err := checkinRandInt(rand.Reader, big.NewInt(100*probabilityScale))
	if err != nil {
		return false, fmt.Errorf("read cryptographic randomness: %w", err)
	}
	return n.Int64() < threshold, nil
}

func secureRandomPositiveTierStep(tiers []CheckinPositiveTier) (float64, error) {
	var totalWeight int64
	for _, tier := range tiers {
		if tier.MinStep < 1 || tier.MaxStep < tier.MinStep || tier.WeightUnits < 1 {
			return 0, fmt.Errorf("invalid positive check-in tier")
		}
		totalWeight += tier.WeightUnits
	}
	if totalWeight < 1 {
		return 0, fmt.Errorf("positive check-in tiers are empty")
	}

	pick, err := checkinRandInt(rand.Reader, big.NewInt(totalWeight))
	if err != nil {
		return 0, fmt.Errorf("read cryptographic randomness: %w", err)
	}
	selected := tiers[len(tiers)-1]
	remaining := pick.Int64()
	for _, tier := range tiers {
		if remaining < tier.WeightUnits {
			selected = tier
			break
		}
		remaining -= tier.WeightUnits
	}

	stepCount := selected.MaxStep - selected.MinStep + 1
	step, err := checkinRandInt(rand.Reader, big.NewInt(stepCount))
	if err != nil {
		return 0, fmt.Errorf("read cryptographic randomness: %w", err)
	}
	return float64(selected.MinStep+step.Int64()) / 100, nil
}

func secureRandomHundredthStep(maxMagnitude float64) (float64, error) {
	steps := int64(math.Round(maxMagnitude * 100))
	if steps < 1 {
		return 0, fmt.Errorf("check-in range must contain a non-zero hundredth step")
	}
	n, err := checkinRandInt(rand.Reader, big.NewInt(steps))
	if err != nil {
		return 0, fmt.Errorf("read cryptographic randomness: %w", err)
	}
	return float64(n.Int64()+1) / 100, nil
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
