package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

const (
	smartRouteCachePrefix = "smart_route:config:v1:"
	smartRouteL1TTL       = time.Minute
	smartRouteRedisTTL    = 5 * time.Minute
)

type smartRouteCacheEntry struct {
	config    *service.SmartRouteConfig
	expiresAt time.Time
}

type smartRouteRepository struct {
	client *dbent.Client
	rdb    *redis.Client
	mu     sync.Mutex
	l1     map[int64]smartRouteCacheEntry
}

func NewSmartRouteRepository(client *dbent.Client, rdb *redis.Client) service.SmartRouteRepository {
	return &smartRouteRepository{client: client, rdb: rdb, l1: make(map[int64]smartRouteCacheEntry)}
}

func (r *smartRouteRepository) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return fn(ctx)
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin smart route transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(dbent.NewTxContext(ctx, tx)); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit smart route transaction: %w", err)
	}
	return nil
}

func (r *smartRouteRepository) Get(ctx context.Context, apiKeyID int64) (*service.SmartRouteConfig, error) {
	configs, err := r.GetMany(ctx, []int64{apiKeyID})
	if err != nil {
		return nil, err
	}
	return configs[apiKeyID], nil
}

func (r *smartRouteRepository) GetMany(ctx context.Context, apiKeyIDs []int64) (map[int64]*service.SmartRouteConfig, error) {
	out := make(map[int64]*service.SmartRouteConfig)
	if len(apiKeyIDs) == 0 {
		return out, nil
	}
	if dbent.TxFromContext(ctx) != nil {
		return r.getManyDB(ctx, apiKeyIDs)
	}
	missing := make([]int64, 0, len(apiKeyIDs))
	for _, id := range apiKeyIDs {
		if config, hit := r.getL1(id); hit {
			if config != nil {
				out[id] = config
			}
			continue
		}
		missing = append(missing, id)
	}
	if r.rdb != nil && len(missing) > 0 {
		keys := make([]string, len(missing))
		for i, id := range missing {
			keys[i] = smartRouteCacheKey(id)
		}
		if values, err := r.rdb.MGet(ctx, keys...).Result(); err == nil {
			dbMissing := make([]int64, 0, len(missing))
			for i, value := range values {
				if value == nil {
					dbMissing = append(dbMissing, missing[i])
					continue
				}
				var config *service.SmartRouteConfig
				if err := json.Unmarshal([]byte(fmt.Sprint(value)), &config); err != nil {
					dbMissing = append(dbMissing, missing[i])
					continue
				}
				r.setL1(missing[i], config)
				if config != nil {
					out[missing[i]] = cloneSmartRouteConfig(config)
				}
			}
			missing = dbMissing
		}
	}
	if len(missing) == 0 {
		return out, nil
	}
	loaded, err := r.getManyDB(ctx, missing)
	if err != nil {
		return nil, err
	}
	for _, id := range missing {
		config := loaded[id]
		r.setCached(ctx, id, config)
		if config != nil {
			out[id] = cloneSmartRouteConfig(config)
		}
	}
	return out, nil
}

func (r *smartRouteRepository) getManyDB(ctx context.Context, apiKeyIDs []int64) (map[int64]*service.SmartRouteConfig, error) {
	out := make(map[int64]*service.SmartRouteConfig)
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
		SELECT r.api_key_id, r.platform, r.subscription_type, r.strategy,
		       r.price_weight, r.speed_weight, r.success_weight,
		       r.rate_guard_enabled, r.max_rate_multiplier, r.max_image_rate_multiplier,
		       r.updated_at, g.group_id, g.position
		FROM api_key_smart_routes r
		LEFT JOIN api_key_smart_route_groups g ON g.api_key_id = r.api_key_id
		WHERE r.api_key_id = ANY($1)
		ORDER BY r.api_key_id, g.position`, pq.Array(apiKeyIDs))
	if err != nil {
		return nil, fmt.Errorf("query smart route configs: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id int64
		var groupID sql.NullInt64
		var position sql.NullInt64
		var maxRate, maxImage sql.NullFloat64
		var platform, subscriptionType, strategy string
		var price, speed, success int
		var guard bool
		var updated time.Time
		if err := rows.Scan(&id, &platform, &subscriptionType, &strategy, &price, &speed, &success, &guard, &maxRate, &maxImage, &updated, &groupID, &position); err != nil {
			return nil, err
		}
		cfg := out[id]
		if cfg == nil {
			cfg = &service.SmartRouteConfig{APIKeyID: id, Mode: service.SmartRouteModeSmart, Platform: platform, SubscriptionType: subscriptionType, Strategy: strategy, Weights: service.SmartRouteWeights{Price: price, Speed: speed, Success: success}, RateGuard: service.SmartRouteRateGuard{Enabled: guard}, UpdatedAt: updated}
			if maxRate.Valid {
				v := maxRate.Float64
				cfg.RateGuard.MaxRateMultiplier = &v
			}
			if maxImage.Valid {
				v := maxImage.Float64
				cfg.RateGuard.MaxImageRateMultiplier = &v
			}
			out[id] = cfg
		}
		if groupID.Valid {
			cfg.CandidateGroupIDs = append(cfg.CandidateGroupIDs, groupID.Int64)
		}
	}
	return out, rows.Err()
}

func smartRouteCacheKey(apiKeyID int64) string {
	return smartRouteCachePrefix + strconv.FormatInt(apiKeyID, 10)
}

func cloneSmartRouteConfig(config *service.SmartRouteConfig) *service.SmartRouteConfig {
	if config == nil {
		return nil
	}
	clone := *config
	clone.CandidateGroupIDs = append([]int64(nil), config.CandidateGroupIDs...)
	clone.RuntimeGroups = nil
	return &clone
}

func (r *smartRouteRepository) getL1(apiKeyID int64) (*service.SmartRouteConfig, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.l1[apiKeyID]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(r.l1, apiKeyID)
		return nil, false
	}
	return cloneSmartRouteConfig(entry.config), true
}

func (r *smartRouteRepository) setL1(apiKeyID int64, config *service.SmartRouteConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.l1[apiKeyID] = smartRouteCacheEntry{config: cloneSmartRouteConfig(config), expiresAt: time.Now().Add(smartRouteL1TTL)}
}

func (r *smartRouteRepository) setCached(ctx context.Context, apiKeyID int64, config *service.SmartRouteConfig) {
	r.setL1(apiKeyID, config)
	if r.rdb == nil {
		return
	}
	payload, err := json.Marshal(config)
	if err == nil {
		_ = r.rdb.Set(ctx, smartRouteCacheKey(apiKeyID), payload, smartRouteRedisTTL).Err()
	}
}

func (r *smartRouteRepository) Invalidate(ctx context.Context, apiKeyID int64) {
	r.mu.Lock()
	delete(r.l1, apiKeyID)
	r.mu.Unlock()
	if r.rdb != nil {
		_ = r.rdb.Del(ctx, smartRouteCacheKey(apiKeyID)).Err()
	}
}

func (r *smartRouteRepository) Replace(ctx context.Context, cfg *service.SmartRouteConfig) error {
	if cfg == nil {
		return errors.New("nil smart route config")
	}
	client := clientFromContext(ctx, r.client)
	if _, err := client.ExecContext(ctx, `UPDATE api_keys SET group_id = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, cfg.APIKeyID); err != nil {
		return fmt.Errorf("clear smart route key group: %w", err)
	}
	if _, err := client.ExecContext(ctx, `
		INSERT INTO api_key_smart_routes (
			api_key_id, platform, subscription_type, strategy, price_weight, speed_weight,
			success_weight, rate_guard_enabled, max_rate_multiplier, max_image_rate_multiplier
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (api_key_id) DO UPDATE SET
			platform=EXCLUDED.platform, subscription_type=EXCLUDED.subscription_type,
			strategy=EXCLUDED.strategy, price_weight=EXCLUDED.price_weight,
			speed_weight=EXCLUDED.speed_weight, success_weight=EXCLUDED.success_weight,
			rate_guard_enabled=EXCLUDED.rate_guard_enabled,
			max_rate_multiplier=EXCLUDED.max_rate_multiplier,
			max_image_rate_multiplier=EXCLUDED.max_image_rate_multiplier, updated_at=NOW()`,
		cfg.APIKeyID, cfg.Platform, cfg.SubscriptionType, cfg.Strategy, cfg.Weights.Price, cfg.Weights.Speed, cfg.Weights.Success, cfg.RateGuard.Enabled, cfg.RateGuard.MaxRateMultiplier, cfg.RateGuard.MaxImageRateMultiplier); err != nil {
		return fmt.Errorf("upsert smart route config: %w", err)
	}
	if _, err := client.ExecContext(ctx, `DELETE FROM api_key_smart_route_groups WHERE api_key_id = $1`, cfg.APIKeyID); err != nil {
		return err
	}
	for position, groupID := range cfg.CandidateGroupIDs {
		if _, err := client.ExecContext(ctx, `INSERT INTO api_key_smart_route_groups (api_key_id, group_id, position) VALUES ($1,$2,$3)`, cfg.APIKeyID, groupID, position); err != nil {
			return fmt.Errorf("insert smart route candidate: %w", err)
		}
	}
	return nil
}

func (r *smartRouteRepository) Delete(ctx context.Context, apiKeyID int64) error {
	_, err := clientFromContext(ctx, r.client).ExecContext(ctx, `DELETE FROM api_key_smart_routes WHERE api_key_id = $1`, apiKeyID)
	return err
}

func (r *smartRouteRepository) GetMetrics(ctx context.Context, groupIDs []int64, model, metric string, start, end time.Time) (map[int64]service.SmartRouteMetric, error) {
	out := make(map[int64]service.SmartRouteMetric)
	if len(groupIDs) == 0 {
		return out, nil
	}
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `
		SELECT group_id, SUM(success_requests), SUM(upstream_affected_requests)
		FROM channel_monitor_v2_metrics_1m
		WHERE group_id = ANY($1) AND bucket_start >= $2 AND bucket_start < $3
		  AND ($4 = '' OR model = $4)
		GROUP BY group_id`, pq.Array(groupIDs), start, end, model)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, success, failed int64
		if err := rows.Scan(&id, &success, &failed); err != nil {
			_ = rows.Close()
			return nil, err
		}
		total := success + failed
		value := service.SmartRouteMetric{GroupID: id, Samples: total}
		if total > 0 {
			value.SuccessRate = float64(success) / float64(total)
		}
		out[id] = value
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	histRows, err := client.QueryContext(ctx, `
		SELECT group_id, upper_bound_ms, SUM(sample_count)
		FROM channel_monitor_v2_latency_histograms_1m
		WHERE group_id = ANY($1) AND bucket_start >= $2 AND bucket_start < $3
		  AND user_id = 0 AND metric = $4 AND ($5 = '' OR model = $5)
		GROUP BY group_id, upper_bound_ms
		ORDER BY group_id, upper_bound_ms`, pq.Array(groupIDs), start, end, metric, model)
	if err != nil {
		return nil, err
	}
	type bucket struct{ upper, count int64 }
	hist := make(map[int64][]bucket)
	for histRows.Next() {
		var id, upper, count int64
		if err := histRows.Scan(&id, &upper, &count); err != nil {
			_ = histRows.Close()
			return nil, err
		}
		hist[id] = append(hist[id], bucket{upper, count})
	}
	if err := histRows.Close(); err != nil {
		return nil, err
	}
	for id, buckets := range hist {
		var total int64
		for _, b := range buckets {
			total += b.count
		}
		target := (total + 1) / 2
		var seen int64
		m := out[id]
		for _, b := range buckets {
			seen += b.count
			if seen >= target {
				m.LatencyP50MS = float64(b.upper)
				break
			}
		}
		out[id] = m
	}
	return out, nil
}
