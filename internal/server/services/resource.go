package services

import (
	"context"
	"github.com/google/uuid"
	"os"
	"secstorage/internal/api"
	"secstorage/internal/fileutil"
	"secstorage/internal/server/storage/resource/model"
)

type ResourceStore interface {
	Save(context.Context, *model.Resource) error
	Delete(context.Context, api.ResourceId, api.UserId) error
	DeleteTx(context.Context, api.ResourceId, api.UserId, func() error) error
	ListByUserId(context.Context, api.UserId, api.ResourceType) ([]model.ShortResourceInfo, error)
	Get(context.Context, api.ResourceId, api.ResourceType, api.UserId) (*model.Resource, error)
}

type ResourceService struct {
	store         ResourceStore
	fileStorePath string
}

func NewResourceStoreService(store ResourceStore, fileStorePath string) *ResourceService {
	return &ResourceService{store: store, fileStorePath: fileStorePath}
}

func (s *ResourceService) Save(ctx context.Context, data *model.Resource) error {
	return s.store.Save(ctx, data)
}

func (s *ResourceService) Delete(ctx context.Context, id api.ResourceId, userId api.UserId) error {
	resource, err := s.store.Get(ctx, id, api.Undefined, userId)
	if err != nil {
		return err
	}
	if resource.Type == api.File {
		return s.store.DeleteTx(ctx, id, userId, func() error {
			return os.Remove(string(resource.Data))
		})
	}
	return s.store.Delete(ctx, id, userId)
}

func (s *ResourceService) ListByUserId(ctx context.Context, userId api.UserId, resourceType api.ResourceType) ([]model.ShortResourceInfo, error) {
	return s.store.ListByUserId(ctx, userId, resourceType)
}

func (s *ResourceService) Get(ctx context.Context, resourceId api.ResourceId, userId api.UserId, rType api.ResourceType) (*model.Resource, error) {
	return s.store.Get(ctx, resourceId, rType, userId)
}

type Close func()

func (s *ResourceService) createFilePath(id api.ResourceId) string {
	return s.fileStorePath + "/" + id.String()
}

func (s *ResourceService) SaveFile(ctx context.Context, userId api.UserId, meta []byte, chunkReceiver func() ([]byte, error)) (api.ResourceId, error) {
	id := uuid.New()
	path := s.createFilePath(id)

	resource := &model.Resource{
		Id:     id,
		UserId: userId,
		Type:   api.File,
		Data:   []byte(path),
		Meta:   meta,
	}

	err := s.store.Save(ctx, resource)

	if err != nil {
		return uuid.Nil, err
	}

	return id, fileutil.Get(path, chunkReceiver)
}

func (s *ResourceService) GetFile(resource *model.Resource, chunkSender func([]byte) error) error {
	return fileutil.Send(string(resource.Data), chunkSender)
}
