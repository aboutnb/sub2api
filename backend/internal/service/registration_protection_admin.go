package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrRegistrationRiskConflict = infraerrors.Conflict("REGISTRATION_RISK_CONFLICT", "账号状态已变化，请刷新后重新审核")
var ErrRegistrationRiskNotFound = infraerrors.NotFound("REGISTRATION_RISK_NOT_FOUND", "风险记录不存在")

type RegistrationRiskFilter struct {
	Kind     string
	Status   string
	Query    string
	Page     int
	PageSize int
}

type RegistrationRiskPage struct {
	Items    json.RawMessage `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type RegistrationProtectionRepository interface {
	List(context.Context, RegistrationRiskFilter) (*RegistrationRiskPage, error)
	ReviewAccount(ctx context.Context, id, actorID int64, action, note string) (int64, error)
	CheckSourceBlock(context.Context, RegistrationSource) (time.Duration, error)
	ActivateSourceBlock(context.Context, RegistrationSource, string, int64) error
	ReleaseSourceBlock(context.Context, int64, int64, string, func(context.Context, string) error) error
	CleanupBefore(context.Context, time.Time, int) (int64, error)
}

type RegistrationProtectionService struct {
	repo        RegistrationProtectionRepository
	Settings    *SettingService
	invalidator APIKeyAuthCacheInvalidator
	counter     AuthIPBanCounter
}

func (s *RegistrationProtectionService) Cleanup(ctx context.Context) error {
	cutoff := time.Now().UTC().AddDate(0, 0, -90)
	for {
		deleted, err := s.repo.CleanupBefore(ctx, cutoff, 1000)
		if err != nil {
			return err
		}
		if deleted == 0 {
			return nil
		}
	}
}

func NewRegistrationProtectionService(repo RegistrationProtectionRepository, settings *SettingService, invalidator APIKeyAuthCacheInvalidator, counter AuthIPBanCounter) *RegistrationProtectionService {
	return &RegistrationProtectionService{repo: repo, Settings: settings, invalidator: invalidator, counter: counter}
}

func (s *RegistrationProtectionService) ReleaseSourceBlock(ctx context.Context, id, actorID int64, note string) error {
	note = strings.TrimSpace(note)
	if id <= 0 || actorID <= 0 || note == "" || len([]rune(note)) > 512 {
		return infraerrors.BadRequest("REGISTRATION_RISK_REVIEW_INVALID", "请填写有效的解除说明")
	}
	if s.counter == nil {
		return ErrServiceUnavailable
	}
	return s.repo.ReleaseSourceBlock(ctx, id, actorID, note, s.counter.Delete)
}

func (s *RegistrationProtectionService) CheckSource(ctx context.Context, source RegistrationSource) error {
	if !source.Policy.Enabled {
		return nil
	}
	retry, err := s.repo.CheckSourceBlock(ctx, source)
	if err != nil {
		return ErrServiceUnavailable
	}
	if retry > 0 {
		return &RegistrationSourceBlocked{RetryAfter: retry}
	}
	return nil
}

func (s *RegistrationProtectionService) RecordSourceFailures(ctx context.Context, source RegistrationSource, ipCount, identityCount int64) error {
	if !source.Policy.Enabled {
		return nil
	}
	scope, count := "", int64(0)
	if ipCount >= int64(source.Policy.IPFailureLimit) {
		scope, count = "ip", ipCount
	} else if identityCount >= int64(source.Policy.IdentityFailureLimit) {
		scope, count = "ip_ua", identityCount
	}
	if scope == "" {
		return nil
	}
	if err := s.repo.ActivateSourceBlock(ctx, source, scope, count); err != nil {
		return ErrServiceUnavailable
	}
	return &RegistrationSourceBlocked{RetryAfter: time.Duration(source.Policy.BlockMinutes) * time.Minute}
}

type RegistrationSourceBlocked struct{ RetryAfter time.Duration }

func (e *RegistrationSourceBlocked) Error() string {
	return "注册验证失败次数过多，当前来源已被临时限制，请稍后再试"
}
func (e *RegistrationSourceBlocked) Unwrap() error {
	return infraerrors.TooManyRequests("REGISTRATION_SOURCE_BLOCKED", e.Error())
}
func (e *RegistrationSourceBlocked) RetryAfterSeconds() int {
	return max(1, int((e.RetryAfter+time.Second-1)/time.Second))
}

func (s *RegistrationProtectionService) List(ctx context.Context, filter RegistrationRiskFilter) (*RegistrationRiskPage, error) {
	switch filter.Kind {
	case "accounts", "sources", "events", "blocks":
	default:
		return nil, infraerrors.BadRequest("REGISTRATION_RISK_FILTER_INVALID", "无效的记录类型")
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	filter.Query = strings.TrimSpace(filter.Query)
	if len(filter.Query) > 256 || filter.Page > 100000 {
		return nil, infraerrors.BadRequest("REGISTRATION_RISK_FILTER_INVALID", "查询范围超出限制")
	}
	return s.repo.List(ctx, filter)
}

func (s *RegistrationProtectionService) ReviewAccount(ctx context.Context, id, actorID int64, action, note string) error {
	note = strings.TrimSpace(note)
	if id <= 0 || actorID <= 0 || (action != "restrict" && action != "release") || note == "" || len([]rune(note)) > 512 {
		return infraerrors.BadRequest("REGISTRATION_RISK_REVIEW_INVALID", "请选择有效操作并填写不超过 512 字的审核说明")
	}
	userID, err := s.repo.ReviewAccount(ctx, id, actorID, action, note)
	if err != nil {
		return err
	}
	if s.invalidator != nil {
		s.invalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	return nil
}
