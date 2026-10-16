package service

import (
	"context"
	"encoding/json"
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
	got := providerPricingTokenPrice(&usdPerToken)
	if got == nil || *got != 9.0 {
		t.Fatalf("token price = %v, want 9.0", got)
	}
}

func TestProviderPricingUnitPriceRoundsToOneDecimal(t *testing.T) {
	got := providerPricingUnitPrice(0.1)
	if got == nil || *got != 0.8 {
		t.Fatalf("unit price = %v, want 0.8", got)
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
	row, ok := providerPricingRowFromChannelPricing(ProviderPricingModel{ModelName: "gpt-test", GroupName: "12", Enabled: true}, pricing, &Group{ID: 12})
	if !ok {
		t.Fatal("expected a valid provider pricing row")
	}
	if row.PriceUnit != providerPricingDefaultUnit || row.InputPrice == nil || *row.InputPrice != 15.0 || row.OutputPrice == nil || *row.OutputPrice != 60.0 {
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
		{ID: 3, Name: "inactive", Platform: PlatformOpenAI, Status: StatusDisabled},
		{ID: 4, Name: "images", Platform: PlatformOpenAI, Status: StatusActive, RateMultiplier: 3, ImagePrice2K: &imagePrice},
		{ID: 5, Name: "orphan", Platform: PlatformOpenAI, Status: StatusActive, ModelPricing: []ChannelModelPricing{{
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
	if got := byGroup["public-composite"]; !got.Enabled || got.InputPrice == nil || *got.InputPrice != 15 {
		t.Fatalf("expected active model without group multiplier, got %+v", got)
	}
	if got := byGroup["inactive"]; got.Enabled || got.InputPrice == nil {
		t.Fatalf("expected priced inactive group, got %+v", got)
	}
	if got := byGroup["orphan"]; got.Enabled || got.InputPrice == nil {
		t.Fatalf("expected unserved configured model to be disabled, got %+v", got)
	}
	if got := byGroup["images"]; !got.Enabled || got.PriceUnit != "per_call" || got.UnitPrice == nil || *got.UnitPrice != 1.5 {
		t.Fatalf("expected 2K image price without group multiplier, got %+v", got)
	}
}
