//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type promptPoolRepo struct {
	GroupRepository
	groups  []Group
	sources []int64
	bound   []int64
	deleted int
	nextID  int64
}

func (r *promptPoolRepo) ListActive(context.Context) ([]Group, error) {
	return append([]Group(nil), r.groups...), nil
}

func (r *promptPoolRepo) Create(_ context.Context, group *Group) error {
	r.nextID++
	group.ID = r.nextID
	r.groups = append(r.groups, *group)
	return nil
}

func (r *promptPoolRepo) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	r.deleted++
	return 1, nil
}

func (r *promptPoolRepo) GetAccountIDsByGroupIDs(_ context.Context, ids []int64) ([]int64, error) {
	r.sources = append([]int64(nil), ids...)
	if len(ids) == 1 && ids[0] == 4 {
		return []int64{40}, nil
	}
	return []int64{30, 40}, nil
}

func (r *promptPoolRepo) BindAccountsToGroup(_ context.Context, _ int64, accountIDs []int64) error {
	r.bound = append([]int64(nil), accountIDs...)
	return nil
}

func TestPlanImageStudioPromptRoutePrefersOpenAIWithoutCreatingGroup(t *testing.T) {
	repo := &promptPoolRepo{groups: []Group{
		{ID: 3, Name: "openai", Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1},
		{ID: 4, Name: "gemini", Status: StatusActive, Platform: PlatformGemini, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 0.5},
		{ID: 9, Name: "exclusive", Status: StatusActive, Platform: PlatformAnthropic, SubscriptionType: SubscriptionTypeStandard, IsExclusive: true, RateMultiplier: 0.1},
		{ID: 10, Name: "subscription", Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription, RateMultiplier: 0.1},
	}}
	models := func(group Group) []string {
		switch group.ID {
		case 3:
			return []string{"gpt-image-2", "gpt-4o"}
		case 4:
			return []string{"gemini-2.5-flash"}
		case 9:
			return []string{"claude-3-5-haiku"}
		case 10:
			return []string{"gpt-4o-mini"}
		default:
			return nil
		}
	}
	keys := &APIKeyService{groupRepo: repo}
	plan, err := keys.PlanImageStudioPromptRoute(context.Background(), models)
	require.NoError(t, err)
	require.Equal(t, "gpt-4o", plan.Model)
	require.Equal(t, "openai", plan.Label)
	require.Equal(t, []int64{3, 4}, []int64{plan.Groups[0].ID, plan.Groups[1].ID})
	require.Equal(t, int64(0), repo.nextID)
	require.Empty(t, repo.bound)
	require.Equal(t, 0, repo.deleted)
}

func TestGetAvailableGroupsHidesImageStudioSmartGroup(t *testing.T) {
	svc := &APIKeyService{
		userRepo:    &visibilityUserRepo{user: &User{ID: 1}},
		userSubRepo: &visibilitySubRepo{},
		groupRepo: &visibilityGroupRepo{groups: []Group{
			{ID: 1, Name: "公开", Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard},
			{ID: 2, Name: ImageStudioSmartGroupName, Description: ImageStudioSmartGroupDescription, Status: StatusActive, Platform: PlatformComposite, SubscriptionType: SubscriptionTypeStandard},
		}},
	}
	got, err := svc.GetAvailableGroups(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, int64(1), got[0].ID)
}

func TestImageStudioPromptModelPrefersGPT56Sol(t *testing.T) {
	models := []string{"gpt-image-2", "gpt-5.4-mini", "GPT-5.6-Sol", "claude-3-5-haiku"}
	require.Equal(t, "GPT-5.6-Sol", ImageStudioPromptModel(models))
	require.Equal(t, []string{"gpt-5.4-mini", "GPT-5.6-Sol", "claude-3-5-haiku"}, ImageStudioPromptModels(models, 0))
}

func TestPlanImageStudioPromptRoutePrefersDefaultGPTModel(t *testing.T) {
	repo := &promptPoolRepo{groups: []Group{
		{ID: 1, Name: "mini", Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 0.2},
		{ID: 2, Name: "sol", Status: StatusActive, Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 1},
		{ID: 4, Name: "gemini", Status: StatusActive, Platform: PlatformGemini, SubscriptionType: SubscriptionTypeStandard, RateMultiplier: 0.1},
	}}
	models := func(group Group) []string {
		switch group.ID {
		case 1:
			return []string{"gpt-5.4-mini"}
		case 2:
			return []string{"gpt-4o", "gpt-5.6-sol"}
		default:
			return []string{"gemini-2.5-flash"}
		}
	}
	plan, err := (&APIKeyService{groupRepo: repo}).PlanImageStudioPromptRoute(context.Background(), models)
	require.NoError(t, err)
	require.Equal(t, "gpt-5.6-sol", plan.Model)
	require.Equal(t, "sol", plan.Label)
	require.Equal(t, int64(2), plan.Groups[0].ID)
}
