package services

import (
	"bufio"
	"context"
	"github.com/google/uuid"
	"os"
	"secstorage/internal/api"
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

func (s *ResourceService) Get(ctx context.Context, resourceId api.ResourceId, userId api.UserId) (*model.Resource, error) {
	return s.store.Get(ctx, resourceId, api.Undefined, userId)
}

type Close func()

func (s *ResourceService) createFilePath(id api.ResourceId) string {
	return s.fileStorePath + "/" + id.String()
}
func (s *ResourceService) SaveFile(ctx context.Context, userId api.UserId, meta []byte) (api.ResourceId, *bufio.Writer, Close, error) {
	id := uuid.New()
	path := s.createFilePath(id)

	resource := &model.Resource{
		Id:     id,
		UserId: userId,
		Type:   2,
		Data:   []byte(path),
		Meta:   meta,
	}

	err := s.store.Save(ctx, resource)

	if err != nil {
		return uuid.Nil, nil, nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return uuid.Nil, nil, nil, err
	}
	writer := bufio.NewWriter(file)
	return id, writer, func() { writer.Flush(); file.Close() }, nil
}

func (s *ResourceService) GetFile(ctx context.Context, id api.ResourceId, userId api.UserId) (*bufio.Reader, []byte, Close, error) {
	resource, err := s.store.Get(ctx, id, api.File, userId)
	if err != nil {
		return nil, nil, nil, err
	}
	path := string(resource.Data)
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, err
	}
	reader := bufio.NewReader(file)
	return reader, resource.Meta, func() { file.Close() }, nil
}
