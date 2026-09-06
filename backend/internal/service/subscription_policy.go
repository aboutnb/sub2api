package service

import (
	"context"
	"sync"
	"time"
)

// SubscriptionSettingReader is the small settings surface needed by the
// subscription expiration policy. Keeping this interface narrow makes the
// policy easy to reuse in repositories and isolated tests.
type SubscriptionSettingReader interface {
	GetValue(ctx context.Context, key string) (string, error)
}

const subscriptionExpirationPolicyCacheTTL = time.Second

// SubscriptionPolicy centralizes the subscription expiration switch. A
// missing or unreadable setting fails open so a transient settings outage does
// not accidentally turn an expiring subscription into a permanent grant.
type SubscriptionPolicy struct {
	reader SubscriptionSettingReader

	mu        sync.Mutex
	cachedAt  time.Time
	enabled   bool
	staticSet bool
}

// NewSubscriptionPolicy creates a policy backed by the settings repository.
func NewSubscriptionPolicy(reader SubscriptionSettingReader) *SubscriptionPolicy {
	return &SubscriptionPolicy{reader: reader, enabled: true}
}

// NewStaticSubscriptionPolicy creates a deterministic policy for tests and
// embedded callers that do not have a settings repository.
func NewStaticSubscriptionPolicy(enabled bool) *SubscriptionPolicy {
	return &SubscriptionPolicy{enabled: enabled, staticSet: true, cachedAt: time.Now()}
}

// ExpirationEnabled reports whether subscription expiration timestamps should
// be enforced. The value is cached briefly to keep the gateway hot path free of
// repeated settings reads while still making an admin toggle take effect soon.
func (p *SubscriptionPolicy) ExpirationEnabled(ctx context.Context) bool {
	if p == nil {
		return true
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.staticSet {
		return p.enabled
	}
	now := time.Now()
	if !p.cachedAt.IsZero() && now.Sub(p.cachedAt) < subscriptionExpirationPolicyCacheTTL {
		return p.enabled
	}

	enabled := true
	if p.reader != nil {
		if raw, err := p.reader.GetValue(ctx, SettingKeySubscriptionExpirationEnabled); err == nil {
			enabled = !isFalseSettingValue(raw)
		}
	}
	p.enabled = enabled
	p.cachedAt = now
	return enabled
}

// Invalidate causes the next policy check to reread the setting.
func (p *SubscriptionPolicy) Invalidate() {
	if p == nil {
		return
	}
	p.mu.Lock()
	p.cachedAt = time.Time{}
	p.mu.Unlock()
}
