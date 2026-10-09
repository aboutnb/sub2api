package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// RechargeBonusTier grants an additional balance percentage when the recharge
// principal reaches MinAmount. The highest matching threshold wins.

func defaultRechargeBonusTiers() []RechargeBonusTier {
	return []RechargeBonusTier{
		{MinAmount: 50, BonusPercent: 5},
		{MinAmount: 100, BonusPercent: 10},
	}
}

func parseRechargeBonusTiersSetting(raw string) []RechargeBonusTier {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultRechargeBonusTiers()
	}

	var tiers []RechargeBonusTier
	if err := json.Unmarshal([]byte(raw), &tiers); err != nil || tiers == nil {
		return defaultRechargeBonusTiers()
	}
	normalized, err := normalizeRechargeBonusTiers(tiers)
	if err != nil {
		return defaultRechargeBonusTiers()
	}
	return normalized
}

func marshalRechargeBonusTiersSetting(tiers []RechargeBonusTier) (string, error) {
	normalized, err := normalizeRechargeBonusTiers(tiers)
	if err != nil {
		return "", err
	}
	if normalized == nil {
		normalized = []RechargeBonusTier{}
	}
	raw, err := json.Marshal(normalized)
	if err != nil {
		return "", fmt.Errorf("marshal recharge bonus tiers: %w", err)
	}
	return string(raw), nil
}

func normalizeRechargeBonusTiers(tiers []RechargeBonusTier) ([]RechargeBonusTier, error) {
	return NormalizeRechargeBonusTiers(tiers)
}

func resolveRechargeBonusPercent(amount float64, tiers []RechargeBonusTier) float64 {
	if !isFinitePositive(amount) {
		return 0
	}
	bestThreshold := 0.0
	bonusPercent := 0.0
	for _, tier := range tiers {
		if !isFinitePositive(tier.MinAmount) || !isFinitePositive(tier.BonusPercent) {
			continue
		}
		if amount >= tier.MinAmount && tier.MinAmount >= bestThreshold {
			bestThreshold = tier.MinAmount
			bonusPercent = tier.BonusPercent
		}
	}
	return bonusPercent
}

func isFinitePositive(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}
