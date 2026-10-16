package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type providerPricingChannelRepoStub struct {
	service.ChannelRepository
	channels  []service.Channel
	listCalls int
}

func (s *providerPricingChannelRepoStub) ListAll(context.Context) ([]service.Channel, error) {
	s.listCalls++
	return s.channels, nil
}

type providerPricingGroupRepoStub struct {
	service.GroupRepository
	groups    []service.Group
	listCalls int
}

func (s *providerPricingGroupRepoStub) List(context.Context, pagination.PaginationParams) ([]service.Group, *pagination.PaginationResult, error) {
	s.listCalls++
	return s.groups, &pagination.PaginationResult{Total: int64(len(s.groups)), Page: 1, PageSize: 1000, Pages: 1}, nil
}

func TestProviderPricingHandlerReturnsHvoyDocument(t *testing.T) {
	gin.SetMode(gin.TestMode)
	input := 2e-6
	output := 8e-6
	channelRepo := &providerPricingChannelRepoStub{channels: []service.Channel{{
		Name:     "channel",
		Status:   service.StatusActive,
		GroupIDs: []int64{7},
		ModelPricing: []service.ChannelModelPricing{{
			Platform:    service.PlatformOpenAI,
			Models:      []string{"gpt-test"},
			BillingMode: service.BillingModeToken,
			InputPrice:  &input,
			OutputPrice: &output,
		}},
	}}}
	groupRepo := &providerPricingGroupRepoStub{groups: []service.Group{{
		ID: 7, Name: "公开分组", Platform: service.PlatformOpenAI, Status: service.StatusActive,
	}}}
	pricingService := service.NewProviderPricingService(
		channelRepo,
		groupRepo,
		nil, nil, nil,
	)
	h := NewProviderPricingHandler(pricingService)
	r := gin.New()
	r.GET("/api/provider/pricing", h.Get)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/provider/pricing", nil)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	cacheControl := w.Header().Get("Cache-Control")
	if !strings.HasPrefix(cacheControl, "public, max-age=") {
		t.Fatalf("cache-control = %q", cacheControl)
	}
	maxAge, err := strconv.Atoi(strings.TrimPrefix(cacheControl, "public, max-age="))
	if err != nil || maxAge < 0 || maxAge > 1800 {
		t.Fatalf("cache-control = %q, want remaining snapshot lifetime", cacheControl)
	}
	var payload struct {
		SchemaVersion string `json:"schema_version"`
		Success       bool   `json:"success"`
		Data          struct {
			Currency string                         `json:"currency"`
			Models   []service.ProviderPricingModel `json:"models"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.SchemaVersion != "1.1" || !payload.Success || payload.Data.Currency != "CNY" {
		t.Fatalf("unexpected envelope: %+v", payload)
	}
	if len(payload.Data.Models) != 1 {
		t.Fatalf("models = %+v", payload.Data.Models)
	}
	model := payload.Data.Models[0]
	if model.ModelName != "gpt-test" || model.GroupName != "公开分组" || !model.Enabled {
		t.Fatalf("unexpected model row: %+v", model)
	}
	if model.InputPrice == nil || *model.InputPrice != 15.0 || model.OutputPrice == nil || *model.OutputPrice != 60.0 {
		t.Fatalf("unexpected converted prices: %+v", model)
	}

	second := httptest.NewRecorder()
	r.ServeHTTP(second, httptest.NewRequest("GET", "/api/provider/pricing", nil))
	if channelRepo.listCalls != 1 || groupRepo.listCalls != 1 {
		t.Fatalf("snapshot was not cached: channel calls=%d group calls=%d", channelRepo.listCalls, groupRepo.listCalls)
	}
}
