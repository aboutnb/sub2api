package service

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

type providerPricingTestChannelRepo struct {
	ChannelRepository
	channels []Channel
}

func (r *providerPricingTestChannelRepo) ListAll(context.Context) ([]Channel, error) {
	return r.channels, nil
}

type providerPricingTestGroupRepo struct {
	GroupRepository
	groups []Group
}

func (r *providerPricingTestGroupRepo) List(_ context.Context, _ pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	return r.groups, &pagination.PaginationResult{Pages: 1}, nil
}

func TestProviderPricingTokenPriceUsesCNYPerMillionAndOneDecimal(t *testing.T) {
	usdPerToken := 1.2e-6
	got := providerPricingTokenPrice(&usdPerToken, 1)
	if got == nil || *got != 1.2 {
		t.Fatalf("token price = %v, want 1.2", got)
	}
}

func TestProviderPricingUnitPricePreservesCents(t *testing.T) {
	got := providerPricingUnitPrice(3, 0.25)
	if got == nil || *got != 0.75 {
		t.Fatalf("unit price = %v, want 0.75", got)
	}
}

func TestProviderPricingRejectsNonPositiveAndNonFinitePrices(t *testing.T) {
	for _, test := range []struct {
		name       string
		base       float64
		multiplier float64
	}{
		{"zero base", 0, 1},
		{"negative base", -1, 1},
		{"zero rate", 30, 0},
		{"negative rate", 30, -1},
		{"nan base", math.NaN(), 1},
		{"infinite rate", 30, math.Inf(1)},
		{"rounds to zero", 0.001, 0.25},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := providerPricingUnitPrice(test.base, test.multiplier); got != nil {
				t.Fatalf("price = %v, want omitted invalid/zero price", *got)
			}
		})
	}
}

func TestProviderPricingImagePricePrefersTwoK(t *testing.T) {
	oneK := 1.0
	twoK := 2.0
	fourK := 4.0
	group := &Group{ImagePrice1K: &oneK, ImagePrice2K: &twoK, ImagePrice4K: &fourK}
	if got := providerPricingImagePrice(group); got == nil || *got != twoK {
		t.Fatalf("image price = %v, want 2K price", got)
	}
}

func TestProviderPricingRowFromChannelPricing(t *testing.T) {
	input := 2e-6
	output := 8e-6
	pricing := &ChannelModelPricing{
		BillingMode:    BillingModeToken,
		InputPrice:     &input,
		OutputPrice:    &output,
		CacheReadPrice: &input,
	}
	row, ok := providerPricingRowFromChannelPricing(ProviderPricingModel{ModelName: "gpt-test", GroupName: "12", Enabled: true}, pricing, &Group{ID: 12, RateMultiplier: 0.25})
	if !ok {
		t.Fatal("expected a valid provider pricing row")
	}
	if row.PriceUnit != providerPricingDefaultUnit || row.InputPrice == nil || *row.InputPrice != 0.5 || row.OutputPrice == nil || *row.OutputPrice != 2.0 {
		t.Fatalf("unexpected row: %+v", row)
	}

	payload, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal row: %v", err)
	}
	if string(payload) == "" || string(payload)[0] != '{' {
		t.Fatalf("row was not encoded as an object: %s", payload)
	}
}

func TestProviderPricingSnapshotPublishesUniquePublicGroupModels(t *testing.T) {
	input := 2e-6
	otherInput := 4e-6
	imagePrice := 0.2
	channelImagePrice := 0.1
	videoPrice := 0.3
	groups := []Group{
		{ID: 1, Name: "public-composite", Platform: PlatformComposite, Status: StatusActive, RateMultiplier: 4},
		{ID: 2, Name: "private", Platform: PlatformOpenAI, Status: StatusActive, IsExclusive: true},
		{ID: 3, Name: "inactive", Platform: PlatformOpenAI, Status: StatusDisabled, RateMultiplier: 1},
		{ID: 4, Name: "images", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 3, ImagePrice2K: &imagePrice},
		{ID: 5, Name: "orphan", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1, ModelPricing: []ChannelModelPricing{{
			Platform: PlatformOpenAI, Models: []string{"orphan-model"}, BillingMode: BillingModeToken, InputPrice: &input,
		}}},
	}
	channels := []Channel{
		{Name: "a-inactive", Status: StatusDisabled, GroupIDs: []int64{1}, ModelPricing: []ChannelModelPricing{
			{Platform: PlatformAnthropic, Models: []string{"shared-model"}, BillingMode: BillingModeToken, InputPrice: &otherInput},
			{Platform: PlatformOpenAI, Models: []string{"shared-model"}, BillingMode: BillingModeToken, InputPrice: &otherInput},
		}},
		{Name: "b-active", Status: StatusActive, GroupIDs: []int64{1, 2, 3}, ModelPricing: []ChannelModelPricing{{
			Platform: PlatformOpenAI, Models: []string{"shared-model"}, BillingMode: BillingModeToken, InputPrice: &input,
		}}},
		{Name: "c-image", Status: StatusActive, GroupIDs: []int64{4}, ModelPricing: []ChannelModelPricing{
			{Platform: PlatformOpenAI, Models: []string{"image-model"}, BillingMode: BillingModeImage, PerRequestPrice: &channelImagePrice},
			{Platform: PlatformOpenAI, Models: []string{"video-model"}, BillingMode: BillingModeVideo, PerRequestPrice: &videoPrice},
		}},
	}
	service := NewProviderPricingService(&providerPricingTestChannelRepo{channels: channels}, &providerPricingTestGroupRepo{groups: groups}, nil, nil, nil)
	snapshot, err := service.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.Models) != 4 {
		t.Fatalf("models = %+v, want four public group/model pairs", snapshot.Models)
	}
	byGroup := make(map[string]ProviderPricingModel)
	for _, model := range snapshot.Models {
		if _, exists := byGroup[model.GroupName]; exists {
			t.Fatalf("duplicate group/model pair: %+v", snapshot.Models)
		}
		byGroup[model.GroupName] = model
	}
	if _, exists := byGroup["private"]; exists {
		t.Fatal("exclusive group was published")
	}
	if got := byGroup["public-composite"]; !got.Enabled || got.InputPrice == nil || *got.InputPrice != 8 {
		t.Fatalf("expected active model with group multiplier applied once, got %+v", got)
	}
	if got := byGroup["inactive"]; got.Enabled || got.InputPrice == nil {
		t.Fatalf("expected priced inactive group, got %+v", got)
	}
	if got := byGroup["orphan"]; got.Enabled || got.InputPrice == nil {
		t.Fatalf("expected unserved configured model to be disabled, got %+v", got)
	}
	if got := byGroup["images"]; !got.Enabled || got.PriceUnit != "per_call" || got.UnitPrice == nil || *got.UnitPrice != 0.6 {
		t.Fatalf("expected 2K image price with group multiplier, got %+v", got)
	}
}

func TestProviderPricingSnapshotUsesOfficialPriceTimesGroupRate(t *testing.T) {
	catalog := &PricingService{}
	data, err := catalog.parsePricingData([]byte(`{
		"provider-official-test": {
			"input_cost_per_token": 0.000006,
			"output_cost_per_token": 0.000030,
			"cache_read_input_token_cost": 0.0000008,
			"cache_creation_input_token_cost": 0.0000075,
			"cache_creation_input_token_cost_above_1hr": 0.000012
		}
	}`))
	if err != nil {
		t.Fatalf("parse catalog: %v", err)
	}
	catalog.pricingData = data
	billing := NewBillingService(nil, catalog)
	resolver := NewModelPricingResolver(nil, billing)
	override := 99e-6
	groups := []Group{
		{ID: 1, Name: "quarter", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 0.25,
			PeakRateEnabled: true, PeakStart: "00:00", PeakEnd: "24:00", PeakRateMultiplier: 10,
			ModelPricing: []ChannelModelPricing{{Models: []string{"provider-official-test"}, InputPrice: &override, OutputPrice: &override}}},
		{ID: 2, Name: "full", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 1},
		{ID: 3, Name: "inactive", Platform: PlatformOpenAI, Status: StatusDisabled, RateMultiplier: 0.25},
	}
	channels := []Channel{{Name: "channel", Status: StatusActive, GroupIDs: []int64{1, 2, 3},
		ModelPricing: []ChannelModelPricing{{Platform: PlatformOpenAI, Models: []string{"provider-official-test"}, InputPrice: &override, OutputPrice: &override}},
	}}
	svc := NewProviderPricingService(&providerPricingTestChannelRepo{channels: channels}, &providerPricingTestGroupRepo{groups: groups}, catalog, billing, resolver)
	snapshot, err := svc.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	if len(snapshot.Models) != 3 || snapshot.Currency != "CNY" {
		t.Fatalf("unexpected snapshot: %+v", snapshot)
	}
	for _, row := range snapshot.Models {
		want := []float64{1.5, 7.5, 0.2, 1.9, 3}
		if row.GroupName == "full" {
			want = []float64{6, 30, 0.8, 7.5, 12}
		}
		prices := []*float64{row.InputPrice, row.OutputPrice, row.CacheInputPrice, row.CacheCreatePrice, row.CacheCreatePrice1h}
		for i, got := range prices {
			if got == nil || *got != want[i] {
				t.Fatalf("group %q price %d = %v, want %v", row.GroupName, i, got, want[i])
			}
		}
		if row.Enabled != (row.GroupName != "inactive") {
			t.Fatalf("incorrect enabled state: %+v", row)
		}
	}
}

func TestProviderPricingPerCallUsesBillingMultiplier(t *testing.T) {
	for _, test := range []struct {
		name        string
		mode        BillingMode
		independent bool
		want        float64
	}{
		{"image group rate", BillingModeImage, false, 0.30},
		{"image independent rate", BillingModeImage, true, 0.60},
		{"per request group rate", BillingModePerRequest, true, 0.30},
	} {
		t.Run(test.name, func(t *testing.T) {
			price := 1.2
			group := &Group{RateMultiplier: 0.25, ImageRateIndependent: test.independent, ImageRateMultiplier: 0.5}
			resolved := &ResolvedPricing{Mode: test.mode, DefaultPerRequestPrice: price}
			svc := &ProviderPricingService{}
			row, ok := svc.buildPerCallRow(ProviderPricingModel{}, group, resolved)
			if !ok || row.UnitPrice == nil || *row.UnitPrice != test.want {
				t.Fatalf("resolved row = %+v, want unit_price %v", row, test.want)
			}
			fallback, ok := providerPricingRowFromChannelPricing(ProviderPricingModel{}, &ChannelModelPricing{BillingMode: test.mode, PerRequestPrice: &price}, group)
			if !ok || fallback.UnitPrice == nil || *fallback.UnitPrice != test.want {
				t.Fatalf("fallback row = %+v, want unit_price %v", fallback, test.want)
			}
		})
	}
}

func TestProviderPricingCustomModelFallsBackToConfiguredRates(t *testing.T) {
	input, output := 6e-6, 30e-6
	billing := NewBillingService(nil, nil)
	resolver := NewModelPricingResolver(nil, billing)
	group := &Group{
		ID: 1, Name: "custom", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 0.25,
		ModelPricing: []ChannelModelPricing{{Models: []string{"provider-custom-test"}, InputPrice: &input, OutputPrice: &output}},
	}
	svc := NewProviderPricingService(nil, nil, nil, billing, resolver)
	row, ok := svc.buildModelRow(context.Background(), group, &providerPricingCandidate{
		name: "provider-custom-test", platform: PlatformOpenAI, enabled: true,
	})
	if !ok || row.InputPrice == nil || *row.InputPrice != 1.5 || row.OutputPrice == nil || *row.OutputPrice != 7.5 {
		t.Fatalf("expected configured fallback with group rate, got %+v", row)
	}
}
