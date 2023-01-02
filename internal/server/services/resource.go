package services

import (
	"context"
	"secstorage/internal/server/storage/resource/model"
)

type ResourceStore interface {
	Save(context.Context, *model.Resource) (model.ResourceId, error)
	Delete(context.Context, model.ResourceId) error
	ListByUserId(context.Context, model.UserId, model.ResourceType) ([]model.ShortResourceInfo, error)
	Get(context.Context, model.ResourceId) (*model.Resource, error)
}

type ResourceService struct {
	store ResourceStore
}

func NewResourceStoreService(store ResourceStore) *ResourceService {
	return &ResourceService{store: store}
}

func (s *ResourceService) Save(ctx context.Context, data *model.Resource) (model.ResourceId, error) {
	return s.store.Save(ctx, data)
}

func (s *ResourceService) Delete(ctx context.Context, id model.ResourceId) error {
	return s.store.Delete(ctx, id)
}

func (s *ResourceService) ListByUserId(ctx context.Context, userId model.UserId, resourceType model.ResourceType) ([]model.ShortResourceInfo, error) {
	return s.store.ListByUserId(ctx, userId, resourceType)
}

func (s *ResourceService) Get(ctx context.Context, resourceId model.ResourceId) (*model.Resource, error) {
	return s.store.Get(ctx, resourceId)
}
