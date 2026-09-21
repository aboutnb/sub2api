package repository

import (
	"context"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *apiKeyRepository) ImageStudioKey(ctx context.Context, userID, groupID int64, credential string, create bool) (*service.APIKey, error) {
	if create {
		_, err := r.sql.ExecContext(ctx, `INSERT INTO api_keys (user_id, group_id, key, name, purpose, status)
   VALUES ($1, $2, $3, 'AI 绘图', 'image_studio', 'active')
   ON CONFLICT (user_id, group_id) WHERE purpose = 'image_studio' AND deleted_at IS NULL DO NOTHING`, userID, groupID, credential)
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
