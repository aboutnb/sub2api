package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
)

var (
	ErrCheckinConfigVersion = infraerrors.Conflict("CHECKIN_CONFIG_VERSION_CONFLICT", "daily check-in configuration changed; reload and try again")
	ErrCheckinChangeReason  = infraerrors.BadRequest("CHECKIN_CHANGE_REASON_REQUIRED", "change reason is required")
	ErrCheckinConfigInput   = infraerrors.BadRequest("CHECKIN_CONFIG_INVALID", "daily check-in configuration is invalid")
)

type AdminCheckinService struct {
	repo     AdminCheckinRepository
	settings SettingRepository
}

func NewAdminCheckinService(repo AdminCheckinRepository, settings SettingRepository) *AdminCheckinService {
	return &AdminCheckinService{repo: repo, settings: settings}
}

func (s *AdminCheckinService) Config(ctx context.Context) (*AdminCheckinConfig, error) {
	if s == nil || s.settings == nil {
		return nil, ErrCheckinConfigInvalid
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
		SettingKeyCheckinConfigVersion,
	})
	if err != nil {
		return nil, ErrCheckinConfigInvalid
	}
	version := int64(1)
	if raw := strings.TrimSpace(values[SettingKeyCheckinConfigVersion]); raw != "" {
		if parsed, parseErr := strconv.ParseInt(raw, 10, 64); parseErr == nil && parsed > 0 {
			version = parsed
		}
	}
	updatedAt := time.Time{}
	if setting, getErr := s.settings.Get(ctx, SettingKeyCheckinConfigVersion); getErr == nil {
		updatedAt = setting.UpdatedAt
	}
	minAccountAgeHours, err1 := parseCheckinInt(values[SettingKeyCheckinMinAccountAge])
	ipWindowMinutes, err2 := parseCheckinInt(values[SettingKeyCheckinIPWindow])
	ipMaxUsers, err3 := parseCheckinInt(values[SettingKeyCheckinIPMaxUsers])
	enabled, err4 := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinEnabled]))
	riskEnabled, err5 := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinRiskEnabled]))
	normalEnabled, err6 := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinNormalEnabled]))
	luckyEnabled, err7 := strconv.ParseBool(strings.TrimSpace(values[SettingKeyCheckinLuckyEnabled]))
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil ||
		validateCheckinConfigValues(
			values[SettingKeyCheckinNormalMin], values[SettingKeyCheckinNormalMax],
			values[SettingKeyCheckinLuckyRewardType],
			values[SettingKeyCheckinLuckyPositiveProbability],
			values[SettingKeyCheckinLuckyMinMultiply], values[SettingKeyCheckinLuckyMaxMultiply],
			values[SettingKeyCheckinLuckyAmountMin], values[SettingKeyCheckinLuckyAmountMax],
			minAccountAgeHours, ipWindowMinutes, ipMaxUsers,
		) != nil {
		return nil, ErrCheckinConfigInvalid
	}
	return &AdminCheckinConfig{
		Enabled:                  enabled,
		NormalEnabled:            normalEnabled,
		LuckyEnabled:             luckyEnabled,
		NormalMin:                values[SettingKeyCheckinNormalMin],
		NormalMax:                values[SettingKeyCheckinNormalMax],
		LuckyRewardType:          values[SettingKeyCheckinLuckyRewardType],
		LuckyPositiveProbability: values[SettingKeyCheckinLuckyPositiveProbability],
		LuckyMinMultiply:         values[SettingKeyCheckinLuckyMinMultiply],
		LuckyMaxMultiply:         values[SettingKeyCheckinLuckyMaxMultiply],
		LuckyAmountMin:           values[SettingKeyCheckinLuckyAmountMin],
		LuckyAmountMax:           values[SettingKeyCheckinLuckyAmountMax],
		RiskEnabled:              riskEnabled,
		MinAccountAgeHours:       minAccountAgeHours,
		IPWindowMinutes:          ipWindowMinutes,
		IPMaxUsers:               ipMaxUsers,
		ConfigVersion:            version,
		UpdatedAt:                updatedAt,
	}, nil
}

func (s *AdminCheckinService) UpdateConfig(ctx context.Context, input AdminCheckinConfigUpdate) (*AdminCheckinConfig, error) {
	if input.ExpectedVersion <= 0 {
		return nil, ErrCheckinConfigInput
	}
	if strings.TrimSpace(input.ChangeReason) == "" {
		return nil, ErrCheckinChangeReason
	}
	if len([]rune(input.ChangeReason)) > 500 {
		return nil, infraerrors.BadRequest("CHECKIN_CHANGE_REASON_TOO_LONG", "change reason must be at most 500 characters")
	}
	if err := validateCheckinConfigValues(
		input.NormalMin,
		input.NormalMax,
		input.LuckyRewardType,
		input.LuckyPositiveProbability,
		input.LuckyMinMultiply,
		input.LuckyMaxMultiply,
		input.LuckyAmountMin,
		input.LuckyAmountMax,
		input.MinAccountAgeHours,
		input.IPWindowMinutes,
		input.IPMaxUsers,
	); err != nil {
		return nil, err
	}
	current, err := s.Config(ctx)
	if err != nil {
		return nil, err
	}
	if input.ExpectedVersion != current.ConfigVersion {
		return nil, ErrCheckinConfigVersion
	}
	updated, err := s.repo.UpdateConfigIfVersion(ctx, current.ConfigVersion, map[string]string{
		SettingKeyCheckinEnabled:                  strconv.FormatBool(input.Enabled),
		SettingKeyCheckinNormalEnabled:            strconv.FormatBool(input.NormalEnabled),
		SettingKeyCheckinLuckyEnabled:             strconv.FormatBool(input.LuckyEnabled),
		SettingKeyCheckinNormalMin:                strings.TrimSpace(input.NormalMin),
		SettingKeyCheckinNormalMax:                strings.TrimSpace(input.NormalMax),
		SettingKeyCheckinLuckyRewardType:          strings.TrimSpace(input.LuckyRewardType),
		SettingKeyCheckinLuckyPositiveProbability: strings.TrimSpace(input.LuckyPositiveProbability),
		SettingKeyCheckinLuckyMinMultiply:         strings.TrimSpace(input.LuckyMinMultiply),
		SettingKeyCheckinLuckyMaxMultiply:         strings.TrimSpace(input.LuckyMaxMultiply),
		SettingKeyCheckinLuckyAmountMin:           strings.TrimSpace(input.LuckyAmountMin),
		SettingKeyCheckinLuckyAmountMax:           strings.TrimSpace(input.LuckyAmountMax),
		SettingKeyCheckinRiskEnabled:              strconv.FormatBool(input.RiskEnabled),
		SettingKeyCheckinMinAccountAge:            strconv.Itoa(input.MinAccountAgeHours),
		SettingKeyCheckinIPWindow:                 strconv.Itoa(input.IPWindowMinutes),
		SettingKeyCheckinIPMaxUsers:               strconv.Itoa(input.IPMaxUsers),
	})
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, ErrCheckinConfigVersion
	}
	return s.Config(ctx)
}

func (s *AdminCheckinService) Overview(ctx context.Context, date string) (*AdminCheckinOverview, error) {
	if strings.TrimSpace(date) == "" {
		date = timezone.Now().Format("2006-01-02")
	}
	return s.repo.AdminOverview(ctx, date)
}

func (s *AdminCheckinService) Records(ctx context.Context, filter AdminCheckinRecordFilter) ([]AdminCheckinRecord, int64, error) {
	return s.repo.AdminList(ctx, filter)
}

func validateCheckinConfigValues(normalMin, normalMax, luckyRewardType, luckyPositiveProbability, luckyMin, luckyMax, luckyAmountMin, luckyAmountMax string, minAccountAgeHours, ipWindowMinutes, ipMaxUsers int) error {
	parse := parseCheckinDecimal
	min, err1 := parse(normalMin)
	max, err2 := parse(normalMax)
	luckyPositiveValue, err3 := parse(luckyPositiveProbability)
	luckyMinValue, err4 := parse(luckyMin)
	luckyMaxValue, err5 := parse(luckyMax)
	luckyAmountMinValue, err6 := parse(luckyAmountMin)
	luckyAmountMaxValue, err7 := parse(luckyAmountMax)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || min < 0 || max < min || max > 100 ||
		!validCheckinLuckyRewardType(strings.TrimSpace(luckyRewardType)) ||
		luckyPositiveValue < 0 || luckyPositiveValue > 100 ||
		luckyMinValue < -1 || luckyMinValue >= 0 || luckyMaxValue <= 0 || luckyMaxValue > 10 ||
		luckyAmountMinValue < -100 || luckyAmountMinValue >= 0 || luckyAmountMaxValue <= 0 || luckyAmountMaxValue > 100 ||
		minAccountAgeHours < 0 || minAccountAgeHours > 720 || ipWindowMinutes < 1 || ipWindowMinutes > 1440 || ipMaxUsers < 1 || ipMaxUsers > 10000 {
		return ErrCheckinConfigInput
	}
	return nil
}
