package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/netip"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

	"github.com/redis/go-redis/v9"
)

const (
	AuthIPBanScopeIP   = "ip"
	AuthIPBanScopeIPUA = "ip_ua"

	AuthUserAgentBrowser    = "browser"
	AuthUserAgentAutomation = "automation"
	AuthUserAgentEmpty      = "empty"
	AuthUserAgentOther      = "other"
)

var ErrAuthIPBanNotFound = infraerrors.NotFound("AUTH_IP_BAN_NOT_FOUND", "封禁记录不存在或已失效")

type AuthIPBan struct {
	ID               int64      `json:"id"`
	IPAddress        string     `json:"ip_address"`
	BanScope         string     `json:"ban_scope"`
	UAHash           string     `json:"-"`
	UserAgent        string     `json:"user_agent"`
	UACategory       string     `json:"ua_category"`
	Source           string     `json:"source"`
	Reason           string     `json:"reason"`
	TriggerPath      string     `json:"trigger_path"`
	TargetIdentifier string     `json:"target_identifier"`
	FailureCount     int        `json:"failure_count"`
	FirstSeenAt      time.Time  `json:"first_seen_at"`
	LastSeenAt       time.Time  `json:"last_seen_at"`
	BannedAt         time.Time  `json:"banned_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	ReleasedAt       *time.Time `json:"released_at,omitempty"`
	ReleasedByUserID *int64     `json:"released_by_user_id,omitempty"`
	ReleasedByEmail  string     `json:"released_by_email"`
	ReleaseNote      string     `json:"release_note"`
	BanCount         int        `json:"ban_count"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	Status           string     `json:"status"`
}

type AuthIPBanFilter struct {
	Page     int
	PageSize int
	Status   string
	Query    string
}

type AuthIPBanList struct {
	Items    []*AuthIPBan
	Total    int
	Page     int
	PageSize int
}

type AuthIPBanActivation struct {
	IPAddress        string
	BanScope         string
	UAHash           string
	UserAgent        string
	UACategory       string
	Source           string
	Reason           string
	TriggerPath      string
	TargetIdentifier string
	FailureCount     int
	FirstSeenAt      time.Time
	LastSeenAt       time.Time
	BannedAt         time.Time
	ExpiresAt        time.Time
}

type AuthIPBanRepository interface {
	FindActive(ctx context.Context, ipAddress, uaHash string, now time.Time) (*AuthIPBan, error)
	Activate(ctx context.Context, input *AuthIPBanActivation) (*AuthIPBan, error)
	List(ctx context.Context, filter *AuthIPBanFilter) (*AuthIPBanList, error)
	GetByID(ctx context.Context, id int64) (*AuthIPBan, error)
	Release(ctx context.Context, id, releasedByUserID int64, note string, now time.Time) (*AuthIPBan, error)
}

type AuthIPBanPolicy struct {
	UACategory string        `json:"ua_category"`
	BanScope   string        `json:"ban_scope"`
	Threshold  int           `json:"threshold"`
	Window     time.Duration `json:"-"`
	BanFor     time.Duration `json:"-"`
	WindowMins int           `json:"window_minutes"`
	BanMins    int           `json:"ban_minutes"`
}

var authIPBanPolicies = map[string]AuthIPBanPolicy{
	AuthUserAgentAutomation: {
		UACategory: AuthUserAgentAutomation,
		BanScope:   AuthIPBanScopeIP,
		Threshold:  8,
		Window:     30 * time.Minute,
		BanFor:     6 * time.Hour,
		WindowMins: 30,
		BanMins:    360,
	},
	AuthUserAgentEmpty: {
		UACategory: AuthUserAgentEmpty,
		BanScope:   AuthIPBanScopeIP,
		Threshold:  8,
		Window:     30 * time.Minute,
		BanFor:     6 * time.Hour,
		WindowMins: 30,
		BanMins:    360,
	},
	AuthUserAgentOther: {
		UACategory: AuthUserAgentOther,
		BanScope:   AuthIPBanScopeIPUA,
		Threshold:  12,
		Window:     30 * time.Minute,
		BanFor:     2 * time.Hour,
		WindowMins: 30,
		BanMins:    120,
	},
	AuthUserAgentBrowser: {
		UACategory: AuthUserAgentBrowser,
		BanScope:   AuthIPBanScopeIPUA,
		Threshold:  20,
		Window:     30 * time.Minute,
		BanFor:     time.Hour,
		WindowMins: 30,
		BanMins:    60,
	},
}

var authIPBanCounterScript = redis.NewScript(`
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('PTTL', KEYS[1])
if ttl < 0 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
  ttl = tonumber(ARGV[1])
end
return {current, ttl}
`)

type AuthIPBanService struct {
	repo  AuthIPBanRepository
	redis *redis.Client
	now   func() time.Time
}

func NewAuthIPBanService(repo AuthIPBanRepository, redisClient *redis.Client) *AuthIPBanService {
	return &AuthIPBanService{repo: repo, redis: redisClient, now: time.Now}
}

func (s *AuthIPBanService) Policies() []AuthIPBanPolicy {
	order := []string{AuthUserAgentAutomation, AuthUserAgentEmpty, AuthUserAgentOther, AuthUserAgentBrowser}
	result := make([]AuthIPBanPolicy, 0, len(order))
	for _, category := range order {
		result = append(result, authIPBanPolicies[category])
	}
	return result
}

func (s *AuthIPBanService) CheckActive(ctx context.Context, ipAddress, userAgent string) (*AuthIPBan, error) {
	if s == nil || s.repo == nil {
		return nil, nil
	}
	ipAddress, ok := normalizePublicAuthIP(ipAddress)
	if !ok {
		return nil, nil
	}
	return s.repo.FindActive(ctx, ipAddress, authUserAgentHash(userAgent), s.currentTime())
}

func (s *AuthIPBanService) RecordFailure(
	ctx context.Context,
	ipAddress, userAgent, targetIdentifier, triggerPath, reason string,
) (*AuthIPBan, error) {
	if s == nil || s.repo == nil || s.redis == nil {
		return nil, nil
	}
	ipAddress, ok := normalizePublicAuthIP(ipAddress)
	if !ok {
		return nil, nil
	}

	userAgent = truncateAuthBanText(userAgent, 512)
	targetIdentifier = strings.ToLower(truncateAuthBanText(targetIdentifier, 255))
	triggerPath = truncateAuthBanText(triggerPath, 512)
	reason = truncateAuthBanText(reason, 128)
	category := ClassifyAuthUserAgent(userAgent)
	policy := authIPBanPolicies[category]
	uaHash := authUserAgentHash(userAgent)
	key := authIPBanCounterKey(ipAddress, uaHash, targetIdentifier, policy)

	values, err := authIPBanCounterScript.Run(
		ctx,
		s.redis,
		[]string{key},
		policy.Window.Milliseconds(),
	).Slice()
	if err != nil {
		slog.Warn("auth_ip_ban.counter_failed", "ip", ipAddress, "ua_category", category, "error", err)
		return nil, nil
	}
	if len(values) < 2 {
		return nil, fmt.Errorf("auth IP ban counter returned %d values", len(values))
	}
	count, err := authIPBanInt64(values[0])
	if err != nil {
		return nil, err
	}
	ttlMillis, err := authIPBanInt64(values[1])
	if err != nil {
		return nil, err
	}
	if count < int64(policy.Threshold) {
		return nil, nil
	}

	now := s.currentTime()
	elapsed := policy.Window - time.Duration(ttlMillis)*time.Millisecond
	if elapsed < 0 || elapsed > policy.Window {
		elapsed = 0
	}
	recordUAHash := ""
	if policy.BanScope == AuthIPBanScopeIPUA {
		recordUAHash = uaHash
	}
	activation := &AuthIPBanActivation{
		IPAddress:        ipAddress,
		BanScope:         policy.BanScope,
		UAHash:           recordUAHash,
		UserAgent:        userAgent,
		UACategory:       category,
		Source:           "auth_auto_ban",
		Reason:           reason,
		TriggerPath:      triggerPath,
		TargetIdentifier: targetIdentifier,
		FailureCount:     int(count),
		FirstSeenAt:      now.Add(-elapsed),
		LastSeenAt:       now,
		BannedAt:         now,
		ExpiresAt:        now.Add(policy.BanFor),
	}
	ban, err := s.repo.Activate(ctx, activation)
	if err != nil {
		return nil, err
	}
	slog.Warn("auth_ip_ban.activated",
		"ip", ipAddress,
		"scope", policy.BanScope,
		"ua_category", category,
		"failure_count", count,
		"expires_at", activation.ExpiresAt,
	)
	return ban, nil
}

func (s *AuthIPBanService) ClearFailures(ctx context.Context, ipAddress, userAgent, targetIdentifier string) {
	if s == nil || s.redis == nil {
		return
	}
	ipAddress, ok := normalizePublicAuthIP(ipAddress)
	if !ok {
		return
	}
	category := ClassifyAuthUserAgent(userAgent)
	policy := authIPBanPolicies[category]
	key := authIPBanCounterKey(
		ipAddress,
		authUserAgentHash(userAgent),
		strings.ToLower(truncateAuthBanText(targetIdentifier, 255)),
		policy,
	)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		slog.Warn("auth_ip_ban.clear_counter_failed", "ip", ipAddress, "error", err)
	}
}

func (s *AuthIPBanService) List(ctx context.Context, filter *AuthIPBanFilter) (*AuthIPBanList, error) {
	if s == nil || s.repo == nil {
		return &AuthIPBanList{Items: []*AuthIPBan{}, Page: 1, PageSize: 20}, nil
	}
	return s.repo.List(ctx, filter)
}

func (s *AuthIPBanService) Release(ctx context.Context, id, releasedByUserID int64, note string) (*AuthIPBan, error) {
	if s == nil || s.repo == nil {
		return nil, ErrAuthIPBanNotFound
	}
	record, err := s.repo.Release(ctx, id, releasedByUserID, truncateAuthBanText(note, 255), s.currentTime())
	if err != nil {
		return nil, err
	}
	if s.redis != nil && record != nil {
		category := record.UACategory
		policy, ok := authIPBanPolicies[category]
		if ok {
			uaHash := record.UAHash
			if uaHash == "" {
				uaHash = authUserAgentHash(record.UserAgent)
			}
			key := authIPBanCounterKey(record.IPAddress, uaHash, record.TargetIdentifier, policy)
			if err := s.redis.Del(ctx, key).Err(); err != nil {
				slog.Warn("auth_ip_ban.release_counter_failed", "id", id, "ip", record.IPAddress, "error", err)
			}
		}
	}
	return record, nil
}

func ClassifyAuthUserAgent(userAgent string) string {
	ua := strings.ToLower(strings.TrimSpace(userAgent))
	if ua == "" {
		return AuthUserAgentEmpty
	}
	automationMarkers := []string{
		"go-http-client", "curl/", "wget/", "python-requests", "python-urllib",
		"aiohttp", "httpie/", "postmanruntime/", "insomnia/", "okhttp/", "java/",
		"libwww-perl", "powershell/", "apache-httpclient", "restsharp/", "node-fetch",
		"axios/", "got/", "scrapy/", "headlesschrome", "phantomjs", "selenium", "playwright",
	}
	for _, marker := range automationMarkers {
		if strings.Contains(ua, marker) {
			return AuthUserAgentAutomation
		}
	}
	if strings.Contains(ua, "mozilla/5.0") && containsAnyAuthUA(ua,
		"chrome/", "crios/", "firefox/", "fxios/", "safari/", "edg/", "edga/", "edgios/",
	) {
		return AuthUserAgentBrowser
	}
	return AuthUserAgentOther
}

func normalizePublicAuthIP(raw string) (string, bool) {
	addr, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return "", false
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsUnspecified() {
		return "", false
	}
	return addr.String(), true
}

func authUserAgentHash(userAgent string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(userAgent))))
	return hex.EncodeToString(sum[:])
}

func authIPBanCounterKey(ipAddress, uaHash, targetIdentifier string, policy AuthIPBanPolicy) string {
	dimension := ipAddress
	if policy.BanScope == AuthIPBanScopeIPUA {
		dimension += "|" + uaHash + "|" + strings.ToLower(strings.TrimSpace(targetIdentifier))
	}
	sum := sha256.Sum256([]byte(dimension))
	return "auth_ip_ban:failure:" + hex.EncodeToString(sum[:])
}

func (s *AuthIPBanService) currentTime() time.Time {
	if s != nil && s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

func containsAnyAuthUA(value string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func truncateAuthBanText(value string, max int) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func authIPBanInt64(value any) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case string:
		var parsed int64
		_, err := fmt.Sscan(v, &parsed)
		return parsed, err
	default:
		return 0, fmt.Errorf("unexpected auth IP ban counter value %T", value)
	}
}
