package services

import (
	"context"
	"secstorage/internal/api"
	"secstorage/internal/server/storage/resource/model"
)

type ResourceStore interface {
	Save(context.Context, *model.Resource) (api.ResourceId, error)
	Delete(context.Context, api.ResourceId) error
	ListByUserId(context.Context, api.UserId, api.ResourceType) ([]model.ShortResourceInfo, error)
	Get(context.Context, api.ResourceId) (*model.Resource, error)
}

type ResourceService struct {
	store ResourceStore
}

func NewResourceStoreService(store ResourceStore) *ResourceService {
	return &ResourceService{store: store}
}

func (s *ResourceService) Save(ctx context.Context, data *model.Resource) (api.ResourceId, error) {
	return s.store.Save(ctx, data)
}

func (s *ResourceService) Delete(ctx context.Context, id api.ResourceId) error {
	return s.store.Delete(ctx, id)
}

func (s *ResourceService) ListByUserId(ctx context.Context, userId api.UserId, resourceType api.ResourceType) ([]model.ShortResourceInfo, error) {
	return s.store.ListByUserId(ctx, userId, resourceType)
}

func (s *ResourceService) Get(ctx context.Context, resourceId api.ResourceId) (*model.Resource, error) {
	return s.store.Get(ctx, resourceId)
}
