package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

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

type authIPBanCounter struct {
	rdb *redis.Client
}

func NewAuthIPBanCounter(rdb *redis.Client) service.AuthIPBanCounter {
	return &authIPBanCounter{rdb: rdb}
}

func (c *authIPBanCounter) Increment(ctx context.Context, key string, window time.Duration) (int64, time.Duration, error) {
	if c == nil || c.rdb == nil {
		return 0, 0, fmt.Errorf("nil auth IP ban counter")
	}
	values, err := authIPBanCounterScript.Run(ctx, c.rdb, []string{key}, window.Milliseconds()).Slice()
	if err != nil {
		return 0, 0, err
	}
	if len(values) < 2 {
		return 0, 0, fmt.Errorf("auth IP ban counter returned %d values", len(values))
	}
	count, err := authIPBanCounterInt64(values[0])
	if err != nil {
		return 0, 0, err
	}
	ttlMillis, err := authIPBanCounterInt64(values[1])
	if err != nil {
		return 0, 0, err
	}
	return count, time.Duration(ttlMillis) * time.Millisecond, nil
}

func (c *authIPBanCounter) Delete(ctx context.Context, key string) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Del(ctx, key).Err()
}

func authIPBanCounterInt64(value any) (int64, error) {
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
