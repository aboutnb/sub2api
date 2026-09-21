//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageStudioModelsChannelMappingAndRestriction(t *testing.T) {
	channel := Channel{ID: 1, Status: StatusActive, GroupIDs: []int64{10}, RestrictModels: true,
		BillingModelSource: BillingModelSourceRequested,
		ModelMapping:       map[string]map[string]string{PlatformOpenAI: {"art": "gpt-image-2", "blocked": "gpt-image-2", "chat": "gpt-5"}},
		ModelPricing:       []ChannelModelPricing{{Platform: PlatformOpenAI, Models: []string{"art", "chat"}}},
	}
	gateway := &GatewayService{
		accountRepo:    studioAccountRepo{accounts: []Account{{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}}},
		channelService: newTestChannelService(makeStandardRepo(channel, map[int64]string{10: PlatformOpenAI})),
	}
	models, err := gateway.ImageStudioModels(context.Background(), Group{ID: 10, Platform: PlatformOpenAI})
	require.NoError(t, err)
	require.Equal(t, []ImageStudioModel{{ID: "art", Platform: PlatformOpenAI, Capabilities: imageStudioCapabilities(PlatformOpenAI)}}, models)
}
