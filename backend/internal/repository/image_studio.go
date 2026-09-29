package repository

import (
	"context"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *apiKeyRepository) ImageStudioKey(ctx context.Context, userID, groupID int64, credential, name string, create bool) (*service.APIKey, error) {
	if name == "" {
		name = "AI 绘图"
	}
	if create {
		_, err := r.sql.ExecContext(ctx, `INSERT INTO api_keys (user_id, group_id, key, name, purpose, status)
   VALUES ($1, $2, $3, $4, 'image_studio', 'active')
   ON CONFLICT (user_id, group_id) WHERE purpose = 'image_studio' AND deleted_at IS NULL DO NOTHING`, userID, groupID, credential, name)
		if err != nil {
			return nil, err
		}
	}
	entity, err := r.activeQuery().Where(apikey.UserIDEQ(userID), apikey.GroupIDEQ(groupID), apikey.PurposeEQ(service.ImageStudioKeyPurpose)).WithUser().WithGroup().Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrAPIKeyNotFound, nil)
	}
	return apiKeyEntityToService(entity), nil
}

func (r *apiKeyRepository) ImageStudioPromptRouteKey(ctx context.Context, userID int64, credential, name string) (*service.APIKey, error) {
	if name == "" {
		name = service.ImageStudioPromptRouteName
	}
	_, err := r.sql.ExecContext(ctx, `INSERT INTO api_keys (user_id, group_id, key, name, purpose, status)
   VALUES ($1, NULL, $2, $3, 'image_studio', 'active')
   ON CONFLICT (user_id) WHERE purpose = 'image_studio' AND group_id IS NULL AND deleted_at IS NULL DO NOTHING`, userID, credential, name)
	if err != nil {
		return nil, err
	}
	entity, err := r.activeQuery().Where(apikey.UserIDEQ(userID), apikey.GroupIDIsNil(), apikey.PurposeEQ(service.ImageStudioKeyPurpose)).WithUser().WithGroup().Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrAPIKeyNotFound, nil)
	}
	// Older builds stored a generic name. Keep any name the user edited.
	if entity.Name == "智能路由" && name != "" && name != entity.Name {
		if _, err := r.sql.ExecContext(ctx, `UPDATE api_keys SET name = $1, updated_at = NOW() WHERE id = $2 AND name = '智能路由' AND deleted_at IS NULL`, name, entity.ID); err != nil {
			return nil, err
		}
		entity.Name = name
	}
	return apiKeyEntityToService(entity), nil
}
