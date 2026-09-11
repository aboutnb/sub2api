package handler

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9" //nolint:depguard // Same Redis as the registration challenge.
)

func (h *AuthHandler) registrationError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	if service.IsRegistrationVerificationFailure(err) {
		if riskErr := h.recordRegistrationSourceFailure(c); riskErr != nil {
			err = riskErr
		}
	}
	return response.ErrorFrom(c, err)
}

func ProvideAuthHandler(cfg *config.Config, auth *service.AuthService, users *service.UserService, settings *service.SettingService, promos *service.PromoService, redeem *service.RedeemService, totp *service.TotpService, attributes *service.UserAttributeService, protection *service.RegistrationProtectionService) *AuthHandler {
	h := NewAuthHandler(cfg, auth, users, settings, promos, redeem, totp, attributes)
	h.registrationProtection = protection
	return h
}

var registrationFailurePairScript = redis.NewScript(`
local counts = {}
for i,key in ipairs(KEYS) do
  local current = tonumber(redis.call('GET',key) or '0')
  if current < tonumber(ARGV[i+1]) then current = redis.call('INCR',key) end
  if redis.call('PTTL',key) < 0 then redis.call('PEXPIRE',key,ARGV[1]) end
  counts[i] = current
end
return counts
`)

func (h *AuthHandler) registrationSource(c *gin.Context) (service.RegistrationSource, error) {
	settings, err := h.settingSvc.GetRegistrationProtectionSettingsCached(c.Request.Context())
	if err != nil {
		return service.RegistrationSource{}, errRegistrationRiskUnavailable
	}
	return service.RegistrationSource{
		IPHash: h.registrationClientIPHash(c), IdentityHash: h.registrationClientIdentityHash(c),
		IPAddress: h.registrationSecurityClientIP(c), UserAgent: normalizeRegistrationUserAgent(c.Request.UserAgent()),
		Path: c.Request.URL.Path, Policy: settings,
	}, nil
}

func (h *AuthHandler) checkRegistrationSource(c *gin.Context) error {
	if h.registrationProtection == nil {
		return nil
	}
	source, err := h.registrationSource(c)
	if err != nil {
		return err
	}
	return h.registrationProtection.CheckSource(c.Request.Context(), source)
}

func (h *AuthHandler) recordRegistrationSourceFailure(c *gin.Context) error {
	if h.registrationProtection == nil {
		return nil
	}
	source, err := h.registrationSource(c)
	if err != nil {
		return err
	}
	if !source.Policy.Enabled {
		return nil
	}
	if h.redisClient == nil {
		return errRegistrationRiskUnavailable
	}
	// Do not let disconnects erase an already verified abuse event.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 3*time.Second)
	defer cancel()
	values, err := registrationFailurePairScript.Run(ctx, h.redisClient, []string{
		"registration_failure:ip:" + source.IPHash, "registration_failure:identity:" + source.IdentityHash,
	}, int64(source.Policy.FailureWindowMinutes)*60000, source.Policy.IPFailureLimit, source.Policy.IdentityFailureLimit).Int64Slice()
	if err != nil || len(values) != 2 {
		return errRegistrationRiskUnavailable
	}
	return h.registrationProtection.RecordSourceFailures(ctx, source, values[0], values[1])
}

// Only a cryptographically valid expired challenge is treated as ordinary
// expiry; an attacker cannot evade counting by editing its expiry field.
func (h *AuthHandler) registrationChallengeFailureIsAbuse(submission *RegistrationChallengeSubmission) bool {
	if submission == nil || strings.TrimSpace(submission.Token) == "" {
		return true
	}
	if strings.TrimSpace(submission.TrapValue) != "" {
		return true
	}
	payload, err := h.parseRegistrationChallengeToken(strings.TrimSpace(submission.Token))
	if err != nil {
		return true
	}
	now := time.Now().UnixMilli()
	if payload.ExpiresAt <= now || now-payload.IssuedAt < payload.MinElapsedMS {
		return false
	}
	return true
}
