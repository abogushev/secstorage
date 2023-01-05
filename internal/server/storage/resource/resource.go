package resource

import (
	"context"
	"github.com/jmoiron/sqlx"
	"secstorage/internal/api"
	"secstorage/internal/server/reservederrors"
	"secstorage/internal/server/storage"
	"secstorage/internal/server/storage/resource/model"
)

type Storage struct {
	ctx context.Context
	db  *sqlx.DB
}

func NewStore(ctx context.Context, db *sqlx.DB) *Storage {
	return &Storage{ctx: ctx, db: db}
}

func (s *Storage) Save(ctx context.Context, resource *model.Resource) error {
	_, err := s.db.ExecContext(
		ctx,
		"insert into resources(id, user_id, type, data, meta) values ($1, $2, $3, $4, $5)",
		resource.Id,
		resource.UserId,
		resource.Type,
		resource.Data,
		resource.Meta,
	)

	if err != nil && storage.IsForeignKeyViolation(err) {
		return reservederrors.ErrUserNotFound
	}

	return err
}

func (s *Storage) Delete(ctx context.Context, resourceId api.ResourceId, userId api.UserId) error {
	_, err := s.db.ExecContext(ctx, "delete from resources where id = $1 and user_id = $2", resourceId, userId)
	return err
}

func (s *Storage) ListByUserId(ctx context.Context, userId api.UserId, resourceType api.ResourceType) ([]model.ShortResourceInfo, error) {
	var results []model.ShortResourceInfo
	err := s.db.SelectContext(
		ctx,
		&results,
		"select id, meta from resources where user_id = $1 and type = $2",
		userId,
		resourceType,
	)
	return results, err
}

func (s *Storage) Get(ctx context.Context, resourceId api.ResourceId, resourceType api.ResourceType, userId api.UserId) (*model.Resource, error) {
	var result model.Resource
	var err error
	if resourceType == 0 {
		err = s.db.GetContext(ctx, &result, "select id, user_id, type, data, meta from resources where id = $1 and user_id = $2", resourceId, userId)
	} else {
		err = s.db.GetContext(ctx, &result, "select id, user_id, type, data, meta from resources where id = $1 and type = $2 and user_id = $3", resourceId, resourceType, userId)
	}
	return &result, err
}

func (s *Storage) DeleteTx(ctx context.Context, id api.ResourceId, userId api.UserId, call func() error) error {
	return storage.RunInTx(
		func(tx *sqlx.Tx) error {
			return call()
		},
		func(tx *sqlx.Tx) error {
			_, err := tx.ExecContext(ctx, "delete from resources where id = $1 and user_id = $2", id, userId)
			return err
		},
	)
}
