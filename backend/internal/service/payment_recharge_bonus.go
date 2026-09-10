package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

const maxRechargeBonusTiers = 20

// RechargeBonusTier grants an additional balance percentage when the recharge
// principal reaches MinAmount. The highest matching threshold wins.
type RechargeBonusTier struct {
	MinAmount    float64 `json:"min_amount"`
	BonusPercent float64 `json:"bonus_percent"`
}

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
	if len(tiers) > maxRechargeBonusTiers {
		return nil, fmt.Errorf("recharge bonus tiers cannot exceed %d entries", maxRechargeBonusTiers)
	}
	normalized := append([]RechargeBonusTier(nil), tiers...)
	for i := range normalized {
		tier := normalized[i]
		if !isFinitePositive(tier.MinAmount) {
			return nil, fmt.Errorf("recharge bonus tier %d minimum amount must be greater than 0", i+1)
		}
		if !isFinitePositive(tier.BonusPercent) || tier.BonusPercent > 100 {
			return nil, fmt.Errorf("recharge bonus tier %d percentage must be greater than 0 and at most 100", i+1)
		}
		if !decimal.NewFromFloat(tier.MinAmount).Equal(decimal.NewFromFloat(tier.MinAmount).Round(2)) {
			return nil, fmt.Errorf("recharge bonus tier %d minimum amount allows at most 2 decimal places", i+1)
		}
		if !decimal.NewFromFloat(tier.BonusPercent).Equal(decimal.NewFromFloat(tier.BonusPercent).Round(2)) {
			return nil, fmt.Errorf("recharge bonus tier %d percentage allows at most 2 decimal places", i+1)
		}
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].MinAmount < normalized[j].MinAmount
	})
	for i := 1; i < len(normalized); i++ {
		if normalized[i-1].MinAmount == normalized[i].MinAmount {
			return nil, fmt.Errorf("recharge bonus tier minimum amounts must be unique")
		}
	}
	return normalized, nil
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
