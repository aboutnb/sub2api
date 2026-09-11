package handler

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"net/netip"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ip"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9" //nolint:depguard // registration risk counters are backed by the route Redis client.
)

const (
	registrationChallengeTTL          = 15 * time.Minute
	registrationChallengeMinElapsed   = 900 * time.Millisecond
	registrationChallengeSigVersion   = "v1"
	registrationChallengeMaxClockSkew = 2 * time.Minute
	registrationRiskWindow            = 10 * time.Minute
	registrationIdentityRiskWindow    = 24 * time.Hour
	registrationRiskTrapWindow        = time.Hour
	registrationUserAgentMaxBytes     = 512
)

var (
	errRegistrationChallengeRequired = infraerrors.BadRequest("REGISTRATION_CHALLENGE_REQUIRED", "注册验证缺失，请重试")
	errRegistrationChallengeInvalid  = infraerrors.BadRequest("REGISTRATION_CHALLENGE_INVALID", "注册验证已失效，请刷新页面后重试")
	errRegistrationTooManyAttempts   = infraerrors.TooManyRequests("REGISTRATION_TOO_MANY_ATTEMPTS", "注册尝试过于频繁，请稍后再试")
	errRegistrationRiskUnavailable   = infraerrors.ServiceUnavailable("REGISTRATION_RISK_CONTROL_UNAVAILABLE", "注册验证服务暂时不可用，请稍后再试")
)

var registrationRiskCounterScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 or redis.call('PTTL', KEYS[1]) < 0 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return current
`)

// Check all dimensions before consuming any quota. Rejected attempts must not
// exhaust unrelated source/email buckets under concurrent traffic.
var registrationRiskLimitsScript = redis.NewScript(`
for i,key in ipairs(KEYS) do
  local count = tonumber(redis.call('GET', key) or '0')
  if count >= tonumber(ARGV[(i-1)*2+1]) then
    local ttl = redis.call('PTTL', key)
    if ttl < 0 then
      ttl = tonumber(ARGV[(i-1)*2+2])
      redis.call('PEXPIRE', key, ttl)
    end
    return ttl
  end
end
for i,key in ipairs(KEYS) do
  redis.call('INCR', key)
  if redis.call('PTTL', key) < 0 then
    redis.call('PEXPIRE', key, ARGV[(i-1)*2+2])
  end
end
return 0
`)

type RegistrationChallengeResponse struct {
	Token        string `json:"token"`
	IssuedAt     int64  `json:"issued_at"`
	ExpiresAt    int64  `json:"expires_at"`
	MinElapsedMS int64  `json:"min_elapsed_ms"`
	TrapField    string `json:"trap_field"`
	Salt         string `json:"salt"`
}

type RegistrationChallengeSubmission struct {
	Token       string `json:"token"`
	CompletedAt int64  `json:"completed_at"`
	Proof       string `json:"proof"`
	TrapField   string `json:"trap_field"`
	TrapValue   string `json:"trap_value"`
}

type registrationChallengeTokenPayload struct {
	Version      string `json:"v"`
	ID           string `json:"id"`
	IssuedAt     int64  `json:"iat"`
	ExpiresAt    int64  `json:"exp"`
	MinElapsedMS int64  `json:"min_ms"`
	TrapField    string `json:"trap"`
	Salt         string `json:"salt"`
	Fingerprint  string `json:"fp"`
}

// GetRegistrationChallenge returns a short-lived browser-bound token required by
// high-risk registration endpoints. It is deliberately lightweight and stateless.
func (h *AuthHandler) GetRegistrationChallenge(c *gin.Context) {
	challenge, err := h.issueRegistrationChallenge(c)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, challenge)
}

func (h *AuthHandler) issueRegistrationChallenge(c *gin.Context) (*RegistrationChallengeResponse, error) {
	now := time.Now()
	payload := registrationChallengeTokenPayload{
		Version:      registrationChallengeSigVersion,
		ID:           randomRegistrationChallengeID(18),
		IssuedAt:     now.UnixMilli(),
		ExpiresAt:    now.Add(registrationChallengeTTL).UnixMilli(),
		MinElapsedMS: registrationChallengeMinElapsed.Milliseconds(),
		TrapField:    "company_website_" + randomRegistrationChallengeID(8),
		Salt:         randomRegistrationChallengeID(16),
		Fingerprint:  h.registrationChallengeFingerprint(c),
	}
	token, err := h.signRegistrationChallengePayload(payload)
	if err != nil {
		return nil, errRegistrationChallengeInvalid.WithCause(err)
	}
	return &RegistrationChallengeResponse{
		Token:        token,
		IssuedAt:     payload.IssuedAt,
		ExpiresAt:    payload.ExpiresAt,
		MinElapsedMS: payload.MinElapsedMS,
		TrapField:    payload.TrapField,
		Salt:         payload.Salt,
	}, nil
}

func (h *AuthHandler) requireRegistrationChallenge(c *gin.Context, action, email string, submission *RegistrationChallengeSubmission) error {
	if err := h.checkRegistrationSource(c); err != nil {
		return err
	}
	payload, err := h.validateRegistrationChallenge(c, action, email, submission)
	if err != nil {
		if h.registrationProtection != nil {
			if h.registrationChallengeFailureIsAbuse(submission) {
				if riskErr := h.recordRegistrationSourceFailure(c); riskErr != nil {
					return riskErr
				}
			}
			return err
		}
		if riskErr := h.recordRegistrationChallengeFailure(c, action, email, submission); riskErr != nil {
			return riskErr
		}
		return err
	}

	if err := h.enforceRegistrationRiskLimits(c, action, email); err != nil {
		return err
	}
	if err := h.consumeRegistrationChallenge(c.Request.Context(), payload.ID, action); err != nil {
		if err == errRegistrationChallengeInvalid {
			if riskErr := h.recordRegistrationSourceFailure(c); riskErr != nil {
				return riskErr
			}
		}
		return err
	}
	h.attachRegistrationVerificationContext(c, action)
	return nil
}

func (h *AuthHandler) validateRegistrationChallenge(c *gin.Context, action, email string, submission *RegistrationChallengeSubmission) (registrationChallengeTokenPayload, error) {
	var empty registrationChallengeTokenPayload
	if submission == nil || strings.TrimSpace(submission.Token) == "" {
		return empty, errRegistrationChallengeRequired
	}
	if strings.TrimSpace(submission.TrapValue) != "" {
		return empty, errRegistrationChallengeInvalid
	}

	payload, err := h.parseRegistrationChallengeToken(strings.TrimSpace(submission.Token))
	if err != nil {
		return empty, errRegistrationChallengeInvalid.WithCause(err)
	}
	now := time.Now()
	nowMillis := now.UnixMilli()
	if payload.ExpiresAt <= nowMillis {
		return empty, errRegistrationChallengeInvalid
	}
	if payload.IssuedAt > now.Add(registrationChallengeMaxClockSkew).UnixMilli() {
		return empty, errRegistrationChallengeInvalid
	}
	if nowMillis-payload.IssuedAt < payload.MinElapsedMS {
		return empty, errRegistrationChallengeInvalid
	}
	if strings.TrimSpace(payload.TrapField) == "" || submission.TrapField != payload.TrapField {
		return empty, errRegistrationChallengeInvalid
	}
	if payload.Fingerprint == "" || payload.Fingerprint != h.registrationChallengeFingerprint(c) {
		return empty, errRegistrationChallengeInvalid
	}
	if submission.CompletedAt <= 0 {
		return empty, errRegistrationChallengeInvalid
	}

	expectedProof := registrationChallengeProof(
		strings.TrimSpace(submission.Token),
		normalizeRegistrationChallengeEmail(email),
		strings.TrimSpace(action),
		submission.CompletedAt,
		payload.TrapField,
		payload.Salt,
	)
	fallbackProof := registrationChallengeFallbackProof(
		strings.TrimSpace(submission.Token),
		normalizeRegistrationChallengeEmail(email),
		strings.TrimSpace(action),
		submission.CompletedAt,
		payload.TrapField,
		payload.Salt,
	)
	if !constantTimeStringEqual(submission.Proof, expectedProof) && !constantTimeStringEqual(submission.Proof, fallbackProof) {
		return empty, errRegistrationChallengeInvalid
	}

	return payload, nil
}

type registrationRiskLimit struct {
	name   string
	value  string
	limit  int64
	window time.Duration
}

func (h *AuthHandler) enforceRegistrationRiskLimits(c *gin.Context, action, email string) error {
	if h == nil || h.redisClient == nil {
		return nil
	}
	action = strings.TrimSpace(action)
	email = normalizeRegistrationChallengeEmail(email)

	emailLimit, ipLimit, identityLimit := registrationRiskLimitsForAction(action)
	limits := make([]registrationRiskLimit, 0, 3)
	if email != "" && emailLimit > 0 {
		limits = append(limits, registrationRiskLimit{
			name:   "email",
			value:  registrationRiskHash(email),
			limit:  emailLimit,
			window: registrationRiskWindow,
		})
	}
	if ipHash := h.registrationClientIPHash(c); ipHash != "" && ipLimit > 0 {
		limits = append(limits, registrationRiskLimit{
			name:   "ip",
			value:  ipHash,
			limit:  ipLimit,
			window: registrationRiskWindow,
		})
	}
	if identityHash := h.registrationClientIdentityHash(c); identityHash != "" && identityLimit > 0 {
		limits = append(limits, registrationRiskLimit{
			name:   "ip_ua",
			value:  identityHash,
			limit:  identityLimit,
			window: registrationIdentityRiskWindow,
		})
	}
	keys := make([]string, 0, len(limits))
	args := make([]any, 0, len(limits)*2)
	// Email and OAuth must not provide independent request budgets.
	budgetAction := "register"
	if action == "send_verify_code" || action == "oauth_pending_send_verify_code" {
		budgetAction = "send_verify_code"
	}
	for _, limit := range limits {
		keys = append(keys, registrationRiskKey("limit_v2", budgetAction, limit.name, limit.value))
		args = append(args, limit.limit, limit.window.Milliseconds())
	}
	if len(keys) > 0 {
		ttl, err := registrationRiskLimitsScript.Run(c.Request.Context(), h.redisClient, keys, args...).Int64()
		if err != nil {
			return errRegistrationRiskUnavailable.WithCause(err)
		}
		if ttl > 0 {
			c.Header("Retry-After", fmt.Sprint((ttl+999)/1000))
			return errRegistrationTooManyAttempts
		}
	}

	h.observeRegistrationNetworkBucket(c, action)
	return nil
}

func registrationRiskLimitsForAction(action string) (emailLimit int64, ipLimit int64, identityLimit int64) {
	switch strings.TrimSpace(action) {
	case "send_verify_code", "oauth_pending_send_verify_code":
		return 3, 10, 0
	case "register", "oauth_pending_create_account":
		return 8, 50, 0
	default:
		return 8, 50, 0
	}
}

func (h *AuthHandler) recordRegistrationChallengeFailure(c *gin.Context, action, email string, submission *RegistrationChallengeSubmission) error {
	if h == nil || h.redisClient == nil || c == nil || c.Request == nil {
		return nil
	}
	action = strings.TrimSpace(action)
	ipLimit := int64(100)
	emailLimit := int64(20)
	window := registrationRiskWindow
	event := "challenge_failure"
	if submission != nil && strings.TrimSpace(submission.TrapValue) != "" {
		ipLimit = 20
		emailLimit = 3
		window = registrationRiskTrapWindow
		event = "trap_hit"
	}

	if ipHash := h.registrationClientIPHash(c); ipHash != "" {
		key := registrationRiskKey(event, action, "ip", ipHash)
		count, err := h.incrementRegistrationRiskCounter(c.Request.Context(), key, window)
		if err != nil {
			slog.Warn("registration challenge failure counter redis error", "action", action, "dimension", "ip", "error", err)
		} else if count > ipLimit {
			slog.Warn("registration challenge failures exceeded", "action", action, "event", event, "dimension", "ip", "count", count, "limit", ipLimit)
			return errRegistrationTooManyAttempts
		}
	}
	if email = normalizeRegistrationChallengeEmail(email); email != "" {
		key := registrationRiskKey(event, action, "email", registrationRiskHash(email))
		if count, err := h.incrementRegistrationRiskCounter(c.Request.Context(), key, window); err == nil && count > emailLimit {
			return errRegistrationTooManyAttempts
		}
	}
	return nil
}

func (h *AuthHandler) observeRegistrationNetworkBucket(c *gin.Context, action string) {
	if h == nil || h.redisClient == nil || c == nil || c.Request == nil {
		return
	}
	bucketHash := h.registrationNetworkBucketHash(c)
	if bucketHash == "" {
		return
	}
	key := registrationRiskKey("observe", strings.TrimSpace(action), "network_bucket", bucketHash)
	count, err := h.incrementRegistrationRiskCounter(c.Request.Context(), key, registrationRiskWindow)
	if err != nil {
		slog.Warn("registration network bucket observation failed", "action", strings.TrimSpace(action), "error", err)
		return
	}
	if count == 50 || count == 100 || count%200 == 0 {
		slog.Warn("registration network bucket activity observed", "action", strings.TrimSpace(action), "count", count)
	}
}

func (h *AuthHandler) consumeRegistrationChallenge(ctx context.Context, challengeID, action string) error {
	if h == nil || h.redisClient == nil {
		return nil
	}
	challengeID = strings.TrimSpace(challengeID)
	action = strings.TrimSpace(action)
	if challengeID == "" || action == "" {
		return errRegistrationChallengeInvalid
	}
	key := "registration_challenge:used:" + registrationRiskHash(action+"\n"+challengeID)
	ok, err := h.redisClient.SetNX(ctx, key, "1", registrationChallengeTTL).Result()
	if err != nil {
		slog.Error("registration challenge consume redis error", "action", action, "error", err)
		return errRegistrationRiskUnavailable.WithCause(err)
	}
	if !ok {
		return errRegistrationChallengeInvalid
	}
	return nil
}

func (h *AuthHandler) attachRegistrationVerificationContext(c *gin.Context, action string) {
	if c == nil || c.Request == nil {
		return
	}
	registrationCtx := service.RegistrationVerificationContext{
		Action:            strings.TrimSpace(action),
		ClientIPHash:      h.registrationClientIPHash(c),
		UserAgentHash:     h.registrationUserAgentHash(c),
		NetworkBucketHash: h.registrationNetworkBucketHash(c),
	}
	c.Request = c.Request.WithContext(service.WithRegistrationVerificationContext(c.Request.Context(), registrationCtx))
}

// AttachSignupRiskIdentity makes the trusted IP+normalized User-Agent identity
// available to every auth flow, including OAuth routes that do not use the
// browser registration challenge endpoint.
func (h *AuthHandler) AttachSignupRiskIdentity(c *gin.Context) {
	if h == nil || c == nil || c.Request == nil {
		return
	}
	fingerprint := h.registrationClientIdentityHash(c)
	if fingerprint == "" {
		return
	}
	c.Request = c.Request.WithContext(service.WithSignupRiskIdentity(c.Request.Context(), fingerprint))
	// Resolve settings only when a user is created, so an unavailable settings
	// store cannot turn this registration guard into an existing-user login ban.
	if h.settingSvc != nil {
		source := service.RegistrationSource{
			IPHash: h.registrationClientIPHash(c), IdentityHash: fingerprint,
			IPAddress:  h.registrationSecurityClientIP(c),
			UserAgent:  normalizeRegistrationUserAgent(c.Request.UserAgent()),
			Path:       c.Request.URL.Path,
			LoadPolicy: h.settingSvc.GetRegistrationProtectionSettingsCached,
		}
		if h.registrationProtection != nil {
			base := source
			source.LoadPolicy = func(ctx context.Context) (service.RegistrationProtectionSettings, error) {
				policy, err := h.settingSvc.GetRegistrationProtectionSettingsCached(ctx)
				if err != nil {
					return policy, err
				}
				current := base
				current.Policy = policy
				return policy, h.registrationProtection.CheckSource(ctx, current)
			}
		}
		c.Request = c.Request.WithContext(service.WithRegistrationSource(c.Request.Context(), source))
	}
}

func (h *AuthHandler) signRegistrationChallengePayload(payload registrationChallengeTokenPayload) (string, error) {
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(rawPayload)
	mac := hmac.New(sha256.New, h.registrationChallengeSecret())
	_, _ = mac.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature, nil
}

func (h *AuthHandler) parseRegistrationChallengeToken(token string) (registrationChallengeTokenPayload, error) {
	var payload registrationChallengeTokenPayload
	encodedPayload, encodedSignature, ok := strings.Cut(token, ".")
	if !ok || encodedPayload == "" || encodedSignature == "" {
		return payload, fmt.Errorf("malformed challenge token")
	}

	mac := hmac.New(sha256.New, h.registrationChallengeSecret())
	_, _ = mac.Write([]byte(encodedPayload))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !constantTimeStringEqual(encodedSignature, expectedSignature) {
		return payload, fmt.Errorf("invalid challenge signature")
	}

	rawPayload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return payload, err
	}
	if err := json.Unmarshal(rawPayload, &payload); err != nil {
		return payload, err
	}
	if payload.Version != registrationChallengeSigVersion || payload.ID == "" || payload.Salt == "" {
		return payload, fmt.Errorf("invalid challenge payload")
	}
	if payload.MinElapsedMS < 0 || payload.MinElapsedMS > registrationChallengeTTL.Milliseconds() {
		return payload, fmt.Errorf("invalid challenge timing")
	}
	return payload, nil
}

func (h *AuthHandler) registrationChallengeSecret() []byte {
	if h != nil && h.cfg != nil && strings.TrimSpace(h.cfg.JWT.Secret) != "" {
		return []byte("registration-challenge:" + strings.TrimSpace(h.cfg.JWT.Secret))
	}
	return []byte("registration-challenge:sub2api-local-fallback")
}

func (h *AuthHandler) registrationChallengeFingerprint(c *gin.Context) string {
	userAgent := ""
	acceptLanguage := ""
	if c != nil && c.Request != nil {
		userAgent = strings.TrimSpace(c.GetHeader("User-Agent"))
		acceptLanguage = strings.TrimSpace(c.GetHeader("Accept-Language"))
	}
	return h.registrationSensitiveHash("challenge-fingerprint-v1", userAgent+"\n"+acceptLanguage)
}

func (h *AuthHandler) registrationClientIPHash(c *gin.Context) string {
	return h.registrationSensitiveHash("client-ip-v1", h.registrationSecurityClientIP(c))
}

func (h *AuthHandler) registrationUserAgentHash(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	return h.registrationSensitiveHash("user-agent-v1", normalizeRegistrationUserAgent(c.GetHeader("User-Agent")))
}

func (h *AuthHandler) registrationNetworkBucketHash(c *gin.Context) string {
	return h.registrationSensitiveHash("network-bucket-v1", registrationChallengeIPBucket(h.registrationSecurityClientIP(c)))
}

func (h *AuthHandler) registrationClientIdentityHash(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	clientIP := h.registrationSecurityClientIP(c)
	userAgent := normalizeRegistrationUserAgent(c.GetHeader("User-Agent"))
	if clientIP == "" {
		return ""
	}
	if userAgent == "" {
		// Missing UA must not become a bypass; it is a stable bucket for this IP.
		userAgent = "<missing-user-agent>"
	}
	return h.registrationSensitiveHash("client-ip-user-agent-v1", clientIP+"\x00"+userAgent)
}

func (h *AuthHandler) registrationSecurityClientIP(c *gin.Context) string {
	return ip.GetSecurityClientIP(c, false)
}

func (h *AuthHandler) registrationSensitiveHash(domain, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	mac := hmac.New(sha256.New, h.registrationChallengeSecret())
	_, _ = mac.Write([]byte(domain + "\x00" + value))
	return hex.EncodeToString(mac.Sum(nil))
}

func normalizeRegistrationUserAgent(value string) string {
	value = strings.TrimSpace(strings.ToValidUTF8(value, ""))
	if len(value) <= registrationUserAgentMaxBytes {
		return value
	}
	value = value[:registrationUserAgentMaxBytes]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

func registrationRiskHash(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func registrationRiskKey(parts ...string) string {
	normalized := make([]string, 0, len(parts)+1)
	normalized = append(normalized, "registration_risk")
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		part = strings.NewReplacer(":", "_", "\n", "_", "\r", "_", " ", "_").Replace(part)
		if part == "" {
			part = "unknown"
		}
		normalized = append(normalized, part)
	}
	return strings.Join(normalized, ":")
}

func (h *AuthHandler) incrementRegistrationRiskCounter(ctx context.Context, key string, window time.Duration) (int64, error) {
	if h == nil || h.redisClient == nil {
		return 1, nil
	}
	if window <= 0 {
		window = registrationRiskWindow
	}
	count, err := registrationRiskCounterScript.Run(ctx, h.redisClient, []string{key}, window.Milliseconds()).Int64()
	if err != nil {
		return 0, err
	}
	return count, nil
}

func registrationChallengeIPBucket(rawIP string) string {
	rawIP = strings.TrimSpace(rawIP)
	addr, err := netip.ParseAddr(rawIP)
	if err != nil {
		return rawIP
	}
	if addr.Is4() {
		prefix, err := addr.Prefix(24)
		if err == nil {
			return prefix.String()
		}
	}
	if addr.Is6() {
		prefix, err := addr.Prefix(64)
		if err == nil {
			return prefix.String()
		}
	}
	return addr.String()
}

func registrationChallengeProof(token, email, action string, completedAt int64, trapField, salt string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{
		token,
		email,
		action,
		fmt.Sprintf("%d", completedAt),
		trapField,
		salt,
	}, "\n")))
	return hex.EncodeToString(sum[:])
}

func registrationChallengeFallbackProof(token, email, action string, completedAt int64, trapField, salt string) string {
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(strings.Join([]string{
		token,
		email,
		action,
		fmt.Sprintf("%d", completedAt),
		trapField,
		salt,
	}, "\n")))
	return fmt.Sprintf("fnv1a:%016x", hasher.Sum64())
}

func normalizeRegistrationChallengeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func constantTimeStringEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return hmac.Equal([]byte(a), []byte(b))
}

func randomRegistrationChallengeID(size int) string {
	if size <= 0 {
		size = 16
	}
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		fallback := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		return base64.RawURLEncoding.EncodeToString(fallback[:])[:size]
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
