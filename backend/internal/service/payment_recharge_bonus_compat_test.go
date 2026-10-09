package service

import (
	"context"
	"math"
	"testing"
)

func TestParseRechargeBonusTiersSetting(t *testing.T) {
	t.Parallel()

	defaults := parseRechargeBonusTiersSetting("")
	if len(defaults) != 2 || defaults[0].MinAmount != 50 || defaults[0].BonusPercent != 5 || defaults[1].MinAmount != 100 || defaults[1].BonusPercent != 10 {
		t.Fatalf("default recharge bonus tiers = %#v", defaults)
	}
	if got := parseRechargeBonusTiersSetting("[]"); len(got) != 0 {
		t.Fatalf("explicit empty recharge bonus tiers = %#v, want empty", got)
	}
	custom := parseRechargeBonusTiersSetting(`[{"min_amount":100,"bonus_percent":10},{"min_amount":50,"bonus_percent":5}]`)
	if len(custom) != 2 || custom[0].MinAmount != 50 || custom[1].MinAmount != 100 {
		t.Fatalf("sorted custom recharge bonus tiers = %#v", custom)
	}
	for _, raw := range []string{"invalid", "null", `[{"min_amount":-1,"bonus_percent":5}]`, `[{"min_amount":50,"bonus_percent":1001}]`} {
		got := parseRechargeBonusTiersSetting(raw)
		if len(got) != 2 || got[0].MinAmount != 50 || got[1].MinAmount != 100 {
			t.Fatalf("invalid setting %q fallback = %#v", raw, got)
		}
	}
}

func TestNormalizeRechargeBonusTiersRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := [][]RechargeBonusTier{
		{{MinAmount: 50, BonusPercent: 5}, {MinAmount: 50, BonusPercent: 10}},
		{{MinAmount: math.NaN(), BonusPercent: 5}},
		{{MinAmount: 50, BonusPercent: math.Inf(1)}},
		{{MinAmount: 50.001, BonusPercent: 5}},
		{{MinAmount: 50, BonusPercent: 5.001}},
	}
	for _, tiers := range tests {
		if _, err := normalizeRechargeBonusTiers(tiers); err == nil {
			t.Fatalf("expected recharge bonus tiers %#v to fail", tiers)
		}
	}
}

func TestResolveRechargeBonusPercentUsesHighestMatchingThreshold(t *testing.T) {
	t.Parallel()

	tiers := []RechargeBonusTier{
		{MinAmount: 100, BonusPercent: 10},
		{MinAmount: 50, BonusPercent: 5},
	}
	for amount, want := range map[float64]float64{
		49.99: 0,
		50:    5,
		55:    5,
		60:    5,
		70:    5,
		99.99: 5,
		100:   10,
		500:   10,
	} {
		if got := resolveRechargeBonusPercent(amount, tiers); got != want {
			t.Fatalf("bonus percent for %.2f = %v, want %v", amount, got, want)
		}
	}
}

func TestUpdatePaymentConfigPersistsRechargeBonusTiers(t *testing.T) {
	repo := &paymentConfigSettingRepoStub{values: map[string]string{}}
	svc := &PaymentConfigService{settingRepo: repo}
	tiers := []RechargeBonusTier{
		{MinAmount: 100, BonusPercent: 10},
		{MinAmount: 50, BonusPercent: 5},
	}
	if err := svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{RechargeBonusTiers: &tiers}); err != nil {
		t.Fatalf("persist recharge bonus tiers: %v", err)
	}
	if got := repo.values[SettingRechargeBonusTiers]; got != `[{"min_amount":50,"bonus_percent":5},{"min_amount":100,"bonus_percent":10}]` {
		t.Fatalf("stored recharge bonus tiers = %q", got)
	}

	empty := []RechargeBonusTier{}
	if err := svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{RechargeBonusTiers: &empty}); err != nil {
		t.Fatalf("disable recharge bonus tiers: %v", err)
	}
	if got := repo.values[SettingRechargeBonusTiers]; got != "[]" {
		t.Fatalf("stored empty recharge bonus tiers = %q, want []", got)
	}

	invalid := []RechargeBonusTier{{MinAmount: 50, BonusPercent: 1001}}
	if err := svc.UpdatePaymentConfig(context.Background(), UpdatePaymentConfigRequest{RechargeBonusTiers: &invalid}); err == nil {
		t.Fatal("expected invalid recharge bonus tiers to fail")
	}
}

func TestRechargePromotionAndFeeCredit(t *testing.T) {
	for _, tc := range []struct {
		mode                          string
		multiplier, paid, want, bonus float64
	}{
		{RechargeBonusModeBonus, 1, 102, 112, 10},
		{RechargeBonusModeBonus, 0.14, 102, 15.68, 1.4},
		{RechargeBonusModeDiscount, 1, 91.8, 101.8, 10},
		{RechargeBonusModeDiscount, 0.14, 91.8, 14.25, 1.4},
	} {
		cfg := &PaymentConfig{RechargeBonusMode: tc.mode, BalanceRechargeMultiplier: tc.multiplier, RechargeBonusTiers: defaultRechargeBonusTiers()}
		quote := quoteRechargeBonus(cfg, 100, "USD")
		got := addCreditedRechargeFee(quote.Credited, tc.paid, quote.PayBase, tc.multiplier)
		if got != tc.want || quote.Bonus != tc.bonus {
			t.Fatalf("mode=%s multiplier=%v total=%v bonus=%v want=%v/%v", tc.mode, tc.multiplier, got, quote.Bonus, tc.want, tc.bonus)
		}
	}
}

func TestRechargeBonusExpandedTiersRoundTrip(t *testing.T) {
	tiers := []RechargeBonusTier{{MinAmount: 0, BonusPercent: 150}}
	encoded, err := marshalRechargeBonusTiersSetting(tiers)
	if err != nil {
		t.Fatal(err)
	}
	parsed := parseRechargeBonusTiersSetting(encoded)
	if len(parsed) != 1 || parsed[0] != tiers[0] {
		t.Fatalf("tiers reset during read: %#v", parsed)
	}
}
