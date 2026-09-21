package service

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type studioAccountRepo struct {
	AccountRepository
	accounts []Account
}

type studioPrivateKeyRepo struct{ APIKeyRepository }

func (studioPrivateKeyRepo) GetByID(context.Context, int64) (*APIKey, error) {
	return &APIKey{ID: 1, UserID: 1, Purpose: ImageStudioKeyPurpose}, nil
}

func TestImageStudioKeyRejectsOrdinaryCRUD(t *testing.T) {
	ctx := context.Background()
	keys := &APIKeyService{apiKeyRepo: studioPrivateKeyRepo{}}
	_, err := keys.GetByID(ctx, 1)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	_, err = keys.Update(ctx, 1, 1, UpdateAPIKeyRequest{})
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	admin := &adminServiceImpl{apiKeyRepo: studioPrivateKeyRepo{}}
	_, err = admin.AdminResetAPIKeyRateLimitUsage(ctx, 1)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	_, err = admin.AdminUpdateAPIKeyGroupID(ctx, 1, nil)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
}

func (r studioAccountRepo) ListSchedulableByGroupID(context.Context, int64) ([]Account, error) {
	return r.accounts, nil
}

func TestImageStudioModelsMappingsAndPermissions(t *testing.T) {
	gateway := &GatewayService{accountRepo: studioAccountRepo{accounts: []Account{
		{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"drawing": "gpt-image-2", "text": "gpt-5", "wild*": "gpt-image-2"}}},
		{Platform: PlatformGemini, Credentials: map[string]any{"model_mapping": map[string]any{"gemini-image": "gemini-image"}}},
	}}}
	group := Group{ID: 1, Platform: PlatformOpenAI, ModelAllowlist: GroupModelAllowlist{Enabled: true, Models: []string{"drawing"}}}
	models, err := gateway.ImageStudioModels(context.Background(), group)
	require.NoError(t, err)
	require.Equal(t, []ImageStudioModel{{ID: "drawing", Platform: PlatformOpenAI, Capabilities: imageStudioCapabilities(PlatformOpenAI)}}, models)
	group.ModelAllowlist.Models = []string{"missing"}
	models, err = gateway.ImageStudioModels(context.Background(), group)
	require.NoError(t, err)
	require.Empty(t, models)
}

func TestImageStudioModelsCompositeRoute(t *testing.T) {
	resolver := NewCompositeRouteResolver(compositeRouteRepoStub{routes: []CompositeModelRoute{
		{GroupID: 1, PublicModel: "art", UpstreamModel: "gpt-image-2", TargetPlatform: PlatformOpenAI, MatchType: CompositeRouteMatchExact, Endpoint: CompositeRouteEndpointImages, Enabled: true},
		{GroupID: 1, PublicModel: "text", UpstreamModel: "gpt-5", TargetPlatform: PlatformOpenAI, MatchType: CompositeRouteMatchExact, Endpoint: CompositeRouteEndpointImages, Enabled: true},
	}})
	gateway := &GatewayService{accountRepo: studioAccountRepo{accounts: []Account{{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}}}, compositeResolver: resolver}
	models, err := gateway.ImageStudioModels(context.Background(), Group{ID: 1, Platform: PlatformComposite})
	require.NoError(t, err)
	require.Contains(t, models, ImageStudioModel{ID: "art", Platform: PlatformOpenAI, Capabilities: imageStudioCapabilities(PlatformOpenAI)})
	for _, model := range models {
		require.NotEqual(t, "text", model.ID)
	}
	gateway.accountRepo = studioAccountRepo{}
	models, err = gateway.ImageStudioModels(context.Background(), Group{ID: 1, Platform: PlatformComposite})
	require.NoError(t, err)
	require.Empty(t, models)
}

func TestImageStudioCredentialCannotAuthenticatePublicly(t *testing.T) {
	s := &APIKeyService{}
	key := imageStudioKeyPrefix + strings.Repeat("a", 64)
	_, err := s.GetByKey(context.Background(), key)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	_, err = s.GetByKey(WithImageStudioCredential(context.Background(), key+"other"), key)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
}

func TestImageStudioAliasRequiresTrustedRequest(t *testing.T) {
	for _, trusted := range []bool{false, true} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("POST", "/v1/images/generations", nil)
		c.Request.Header.Set("Content-Type", "application/json")
		if trusted {
			c.Request = c.Request.WithContext(WithImageStudioCredential(c.Request.Context(), imageStudioKeyPrefix+"test"))
		}
		_, err := (&OpenAIGatewayService{}).ParseOpenAIImagesRequest(c, []byte(`{"model":"drawing","prompt":"test"}`))
		if trusted {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}
