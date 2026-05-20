package service

import (
	"crypto/rand"
	"strings"
	"time"
)

const (
	redeemCodeAlphabet  = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	redeemCodeLength    = 16
	redeemCodeGroupSize = 4
)

type RedeemCode struct {
	ID        int64
	Code      string
	Type      string
	Value     float64
	Status    string
	UsedBy    *int64
	UsedAt    *time.Time
	Notes     string
	CreatedAt time.Time
	ExpiresAt *time.Time

	GroupID      *int64
	ValidityDays int

	User  *User
	Group *Group
}

func (r *RedeemCode) IsUsed() bool {
	return r.Status == StatusUsed
}

func (r *RedeemCode) IsExpired() bool {
	return r.IsExpiredAt(time.Now())
}

func (r *RedeemCode) IsExpiredAt(now time.Time) bool {
	if r == nil {
		return false
	}
	if r.Status == StatusExpired {
		return true
	}
	return r.Status == StatusUnused && r.ExpiresAt != nil && !r.ExpiresAt.After(now)
}

func (r *RedeemCode) CanUse() bool {
	return r.Status == StatusUnused && !r.IsExpired()
}

func GenerateRedeemCode() (string, error) {
	b := make([]byte, redeemCodeLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.Grow(redeemCodeLength + (redeemCodeLength/redeemCodeGroupSize - 1))

	for i, v := range b {
		if i > 0 && i%redeemCodeGroupSize == 0 {
			builder.WriteByte('-')
		}
		builder.WriteByte(redeemCodeAlphabet[int(v)&31])
	}

	return builder.String(), nil
}
