package resource

import (
	"context"
	"github.com/google/uuid"
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

func (s *Storage) Save(ctx context.Context, resource *model.Resource) (api.ResourceId, error) {
	var id api.ResourceId

	err := s.db.QueryRowContext(
		ctx,
		"insert into resources(id, user_id, type, data, meta) values (gen_random_uuid(), $1,$2,$3,$4) returning id",
		resource.UserId,
		resource.Type,
		resource.Data,
		resource.Meta,
	).Scan(&id)

	if err != nil && storage.IsForeignKeyViolation(err) {
		return uuid.Nil, reservederrors.ErrUserNotFound
	}

	return id, err
}

func (s *Storage) Delete(ctx context.Context, resourceId api.ResourceId) error {
	_, err := s.db.ExecContext(ctx, "delete from resources where id = $1", resourceId)
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

func (s *Storage) Get(ctx context.Context, resourceId api.ResourceId) (*model.Resource, error) {
	var result model.Resource
	err := s.db.GetContext(ctx, &result, "select id, user_id, type, data, meta from resources where id = $1", resourceId)
	return &result, err
}
