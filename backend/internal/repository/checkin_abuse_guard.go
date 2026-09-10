package repository

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

var checkinAbuseGuardScript = redis.NewScript(`
local added = redis.call('SADD', KEYS[1], ARGV[1])
local count = redis.call('SCARD', KEYS[1])
local ttl = redis.call('PTTL', KEYS[1])
if ttl < 0 then
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
  ttl = tonumber(ARGV[2])
end
if added == 0 then
  return {1, count, ttl}
end
if count > tonumber(ARGV[3]) then
  redis.call('SREM', KEYS[1], ARGV[1])
  return {0, count, ttl}
end
return {1, count, ttl}
`)

var checkinMultiSourceAbuseGuardScript = redis.NewScript(`
local member = ARGV[1]
local max_count = 0
local max_ttl = 0

-- Check every source before mutating any of them. Each source has its own
-- window/max pair in ARGV: window_ms, max_users.
for i,key in ipairs(KEYS) do
  local offset = 2 + ((i - 1) * 2)
  local window = tonumber(ARGV[offset])
  local max_users = tonumber(ARGV[offset + 1])
  local is_member = redis.call('SISMEMBER', key, member)
  local count = redis.call('SCARD', key)
  if is_member == 0 and count >= max_users then
    local ttl = redis.call('PTTL', key)
    if ttl < 0 then ttl = window end
    return {0, count, ttl}
  end
  if count > max_count then max_count = count end
  local ttl = redis.call('PTTL', key)
  if ttl < 0 then ttl = window end
  if ttl > max_ttl then max_ttl = ttl end
end

for i,key in ipairs(KEYS) do
  local offset = 2 + ((i - 1) * 2)
  local window = tonumber(ARGV[offset])
  redis.call('SADD', key, member)
  local ttl = redis.call('PTTL', key)
  if ttl < 0 then redis.call('PEXPIRE', key, window) end
end
return {1, max_count + 1, max_ttl}
`)

var checkinRequestGuardScript = redis.NewScript(`
local user_count = redis.call('INCR', KEYS[1])
local user_ttl = redis.call('PTTL', KEYS[1])
if user_ttl < 0 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
  user_ttl = tonumber(ARGV[1])
end
if user_count > tonumber(ARGV[2]) then
  return {0, user_ttl}
end

local source_count = redis.call('INCR', KEYS[2])
local source_ttl = redis.call('PTTL', KEYS[2])
if source_ttl < 0 then
  redis.call('PEXPIRE', KEYS[2], ARGV[1])
  source_ttl = tonumber(ARGV[1])
end

local allowed = 1
if source_count > tonumber(ARGV[3]) then allowed = 0 end
local retry_after = user_ttl
if source_ttl > retry_after then
  retry_after = source_ttl
end
return {allowed, retry_after}
`)

type checkinAbuseGuard struct {
	rdb     *redis.Client
	hashKey []byte
}

func NewCheckinAbuseGuard(rdb *redis.Client, cfg *config.Config) service.CheckinAbuseGuard {
	var hashKey []byte
	if cfg != nil {
		hashKey = []byte(strings.TrimSpace(cfg.JWT.Secret))
	}
	return &checkinAbuseGuard{rdb: rdb, hashKey: hashKey}
}

func (g *checkinAbuseGuard) CheckRequest(
	ctx context.Context,
	source string,
	userID int64,
	window time.Duration,
	userLimit int,
	sourceLimit int,
) (bool, time.Duration, error) {
	if g == nil || g.rdb == nil || len(g.hashKey) == 0 {
		return false, 0, fmt.Errorf("nil check-in request guard")
	}
	if strings.TrimSpace(source) == "" || userID <= 0 || window <= 0 || userLimit <= 0 || sourceLimit <= 0 {
		return false, 0, fmt.Errorf("invalid check-in request guard input")
	}

	digest := g.sourceDigest(source)
	keys := []string{
		"checkin:security:user:" + g.userDigest(userID),
		"checkin:security:source:" + digest,
	}
	values, err := checkinRequestGuardScript.Run(ctx, g.rdb, keys,
		window.Milliseconds(), userLimit, sourceLimit).Slice()
	if err != nil {
		return false, 0, err
	}
	if len(values) != 2 {
		return false, 0, fmt.Errorf("check-in request guard returned %d values", len(values))
	}
	allowed, err := checkinAbuseGuardInt64(values[0])
	if err != nil {
		return false, 0, err
	}
	retryMillis, err := checkinAbuseGuardInt64(values[1])
	if err != nil {
		return false, 0, err
	}
	return allowed == 1, time.Duration(retryMillis) * time.Millisecond, nil
}

func (g *checkinAbuseGuard) CheckAndRecord(
	ctx context.Context,
	source string,
	userID int64,
	window time.Duration,
	maxUsers int,
) (bool, int64, time.Duration, error) {
	if g == nil || g.rdb == nil || len(g.hashKey) == 0 {
		return false, 0, 0, fmt.Errorf("nil check-in abuse guard")
	}
	if strings.TrimSpace(source) == "" || userID <= 0 || window <= 0 || maxUsers <= 0 {
		return false, 0, 0, fmt.Errorf("invalid check-in abuse guard input")
	}

	key := "checkin:risk:source:" + g.sourceDigest(source)
	values, err := checkinAbuseGuardScript.Run(ctx, g.rdb, []string{key},
		g.userDigest(userID), window.Milliseconds(), maxUsers).Slice()
	if err != nil {
		return false, 0, 0, err
	}
	if len(values) != 3 {
		return false, 0, 0, fmt.Errorf("check-in abuse guard returned %d values", len(values))
	}
	allowedValue, err := checkinAbuseGuardInt64(values[0])
	if err != nil {
		return false, 0, 0, err
	}
	count, err := checkinAbuseGuardInt64(values[1])
	if err != nil {
		return false, 0, 0, err
	}
	ttlMillis, err := checkinAbuseGuardInt64(values[2])
	if err != nil {
		return false, 0, 0, err
	}
	return allowedValue == 1, count, time.Duration(ttlMillis) * time.Millisecond, nil
}

func (g *checkinAbuseGuard) CheckAndRecordSources(
	ctx context.Context,
	sources []service.CheckinSourceLimit,
	userID int64,
) (bool, int64, time.Duration, error) {
	if g == nil || g.rdb == nil || len(g.hashKey) == 0 {
		return false, 0, 0, fmt.Errorf("nil check-in abuse guard")
	}
	if userID <= 0 || len(sources) == 0 {
		return false, 0, 0, fmt.Errorf("invalid check-in multi-source guard input")
	}
	keys := make([]string, 0, len(sources))
	args := make([]any, 0, 1+len(sources)*2)
	args = append(args, g.userDigest(userID))
	for _, source := range sources {
		if strings.TrimSpace(source.Source) == "" || source.Window <= 0 || source.MaxUsers <= 0 {
			return false, 0, 0, fmt.Errorf("invalid check-in multi-source guard source")
		}
		keys = append(keys, "checkin:risk:source:"+g.sourceDigest(source.Source))
		args = append(args, source.Window.Milliseconds(), source.MaxUsers)
	}
	values, err := checkinMultiSourceAbuseGuardScript.Run(ctx, g.rdb, keys, args...).Slice()
	if err != nil {
		return false, 0, 0, err
	}
	if len(values) != 3 {
		return false, 0, 0, fmt.Errorf("check-in multi-source abuse guard returned %d values", len(values))
	}
	allowedValue, err := checkinAbuseGuardInt64(values[0])
	if err != nil {
		return false, 0, 0, err
	}
	count, err := checkinAbuseGuardInt64(values[1])
	if err != nil {
		return false, 0, 0, err
	}
	ttlMillis, err := checkinAbuseGuardInt64(values[2])
	if err != nil {
		return false, 0, 0, err
	}
	return allowedValue == 1, count, time.Duration(ttlMillis) * time.Millisecond, nil
}

func (g *checkinAbuseGuard) sourceDigest(source string) string {
	return g.digest("checkin-source-v1", strings.TrimSpace(source))
}

func (g *checkinAbuseGuard) userDigest(userID int64) string {
	return g.digest("checkin-user-v1", strconv.FormatInt(userID, 10))
}

func (g *checkinAbuseGuard) digest(domain, value string) string {
	mac := hmac.New(sha256.New, g.hashKey)
	_, _ = mac.Write([]byte(domain + "\x00"))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func checkinAbuseGuardInt64(value any) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case int:
		return int64(typed), nil
	case string:
		return strconv.ParseInt(typed, 10, 64)
	default:
		return 0, fmt.Errorf("unexpected check-in abuse guard value %T", value)
	}
}
