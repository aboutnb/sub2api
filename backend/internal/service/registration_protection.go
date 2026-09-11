package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const SettingKeyRegistrationProtection = "registration_protection"

// RegistrationProtectionSettings separates successful account creation quotas
// from request throttling. Disabling it does not disable CAPTCHA or route limits.
type RegistrationProtectionSettings struct {
	Enabled                     bool `json:"enabled"`
	IPSuccessLimit              int  `json:"ip_success_limit"`
	IdentitySuccessLimit        int  `json:"identity_success_limit"`
	SuccessWindowHours          int  `json:"success_window_hours"`
	IPFailureLimit              int  `json:"ip_failure_limit"`
	IdentityFailureLimit        int  `json:"identity_failure_limit"`
	FailureWindowMinutes        int  `json:"failure_window_minutes"`
	BlockMinutes                int  `json:"block_minutes"`
	ObserveIdentitySuccessLimit int  `json:"observe_identity_success_limit"`
}

func DefaultRegistrationProtectionSettings() RegistrationProtectionSettings {
	return RegistrationProtectionSettings{
		Enabled: true, IPSuccessLimit: 10, IdentitySuccessLimit: 2,
		SuccessWindowHours: 24, IPFailureLimit: 100, IdentityFailureLimit: 20,
		FailureWindowMinutes: 10, BlockMinutes: 15, ObserveIdentitySuccessLimit: 4,
	}
}

func (s RegistrationProtectionSettings) Validate() error {
	limits := []int{s.IPSuccessLimit, s.IdentitySuccessLimit, s.IPFailureLimit, s.IdentityFailureLimit, s.ObserveIdentitySuccessLimit}
	for _, n := range limits {
		if n < 1 || n > 100000 {
			return fmt.Errorf("注册防护数量必须在 1 至 100000 之间")
		}
	}
	if s.SuccessWindowHours < 1 || s.SuccessWindowHours > 720 || s.FailureWindowMinutes < 1 || s.FailureWindowMinutes > 1440 || s.BlockMinutes < 1 || s.BlockMinutes > 10080 {
		return fmt.Errorf("注册防护时间窗口超出允许范围")
	}
	if s.IdentitySuccessLimit > s.IPSuccessLimit || s.IdentityFailureLimit > s.IPFailureLimit {
		return fmt.Errorf("IP 与 UA 的限制数量不能大于 IP 总量限制")
	}
	return nil
}

func (s *SettingService) GetRegistrationProtectionSettings(ctx context.Context) (RegistrationProtectionSettings, error) {
	settings := DefaultRegistrationProtectionSettings()
	if s == nil || s.settingRepo == nil {
		return settings, ErrServiceUnavailable
	}
	value, err := s.settingRepo.GetValue(ctx, SettingKeyRegistrationProtection)
	if errors.Is(err, ErrSettingNotFound) {
		return settings, nil
	}
	if err != nil {
		return settings, err
	}
	if strings.TrimSpace(value) != "" {
		if err := json.Unmarshal([]byte(value), &settings); err != nil {
			return settings, err
		}
	}
	return settings, settings.Validate()
}

func (s *SettingService) SetRegistrationProtectionSettings(ctx context.Context, settings RegistrationProtectionSettings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	if s == nil || s.settingRepo == nil {
		return ErrServiceUnavailable
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyRegistrationProtection, string(data)); err != nil {
		return err
	}
	s.registrationProtectionCache.Store(&cachedRegistrationProtection{settings: settings, expiresAt: time.Now().Add(time.Minute)})
	return nil
}

type cachedRegistrationProtection struct {
	settings  RegistrationProtectionSettings
	expiresAt time.Time
}

// Expired settings are never used to silently bypass registration protection.
// The independent read deadline prevents one disconnected request from
// cancelling a refresh shared by other registrations.
func (s *SettingService) GetRegistrationProtectionSettingsCached(ctx context.Context) (RegistrationProtectionSettings, error) {
	if s == nil || s.settingRepo == nil {
		return DefaultRegistrationProtectionSettings(), ErrServiceUnavailable
	}
	if entry, ok := s.registrationProtectionCache.Load().(*cachedRegistrationProtection); ok && time.Now().Before(entry.expiresAt) {
		return entry.settings, nil
	}
	value, err, _ := s.registrationProtectionSF.Do("registration_protection", func() (any, error) {
		if entry, ok := s.registrationProtectionCache.Load().(*cachedRegistrationProtection); ok && time.Now().Before(entry.expiresAt) {
			return entry.settings, nil
		}
		readCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		settings, err := s.GetRegistrationProtectionSettings(readCtx)
		if err != nil {
			return nil, err
		}
		s.registrationProtectionCache.Store(&cachedRegistrationProtection{settings: settings, expiresAt: time.Now().Add(time.Minute)})
		return settings, nil
	})
	if err != nil {
		return DefaultRegistrationProtectionSettings(), err
	}
	settings, ok := value.(RegistrationProtectionSettings)
	if !ok {
		return DefaultRegistrationProtectionSettings(), ErrServiceUnavailable
	}
	return settings, nil
}

// RegistrationSource is set only by trusted public registration handlers.
// The source hashes are server-derived; no request field is trusted as identity.
type RegistrationSource struct {
	IPHash       string
	IdentityHash string
	IPAddress    string
	UserAgent    string
	Path         string
	Policy       RegistrationProtectionSettings
	LoadPolicy   func(context.Context) (RegistrationProtectionSettings, error)
}

type registrationSourceContextKey struct{}

type registrationRollbackContextKey struct{}

// WithRegistrationRollback is only for compensation of a failed new-account
// flow. Normal account deletion must retain its consumed registration quota.
func WithRegistrationRollback(ctx context.Context) context.Context {
	return context.WithValue(ctx, registrationRollbackContextKey{}, true)
}

func IsRegistrationRollback(ctx context.Context) bool {
	rollback, _ := ctx.Value(registrationRollbackContextKey{}).(bool)
	return rollback
}

func WithRegistrationSource(ctx context.Context, source RegistrationSource) context.Context {
	return context.WithValue(ctx, registrationSourceContextKey{}, source)
}

func RegistrationSourceFromContext(ctx context.Context) (RegistrationSource, bool) {
	if ctx == nil {
		return RegistrationSource{}, false
	}
	source, ok := ctx.Value(registrationSourceContextKey{}).(RegistrationSource)
	return source, ok && source.IPHash != "" && source.IdentityHash != ""
}

// RegistrationQuotaExceeded carries the remaining window without disclosing
// the matched source dimension in a public error response.
type RegistrationQuotaExceeded struct {
	RetryAfter time.Duration
	Dimension  string
}

func (e *RegistrationQuotaExceeded) Error() string {
	return "当前注册来源已达到注册数量限制，请稍后再试"
}

func (e *RegistrationQuotaExceeded) Unwrap() error {
	return infraerrors.TooManyRequests("REGISTRATION_SOURCE_QUOTA_EXCEEDED", e.Error())
}

func (e *RegistrationQuotaExceeded) RetryAfterSeconds() int {
	return max(1, int(math.Ceil(e.RetryAfter.Seconds())))
}

func registrationQuotaError(err error) error {
	var blocked *RegistrationSourceBlocked
	if errors.As(err, &blocked) {
		return blocked
	}
	var quota *RegistrationQuotaExceeded
	if errors.As(err, &quota) {
		return quota
	}
	return nil
}
