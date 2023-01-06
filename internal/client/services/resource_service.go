package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"secstorage/internal/api"
	pb "secstorage/internal/api/proto"
	"secstorage/internal/client/model"
	"secstorage/internal/fileutil"
)

type ResourceService struct {
	resourceClient pb.ResourcesClient
	fileStorePath  string
}

func NewResourceService(cl pb.ResourcesClient, fileStorePath string) *ResourceService {
	return &ResourceService{resourceClient: cl, fileStorePath: fileStorePath}
}

func (s *ResourceService) Save(ctx context.Context, dType api.ResourceType, data []byte, meta []byte) (api.ResourceId, error) {
	id, err := s.resourceClient.Save(ctx, &pb.Resource{
		Type: pb.TYPE(dType),
		Data: data,
		Meta: meta,
	})
	if err != nil {
		return uuid.Nil, err
	}
	rId, err := uuid.FromBytes(id.Value)
	if err != nil {
		return uuid.Nil, err
	}
	return rId, nil
}

func (s *ResourceService) Delete(ctx context.Context, resourceId api.ResourceId) error {
	_, err := s.resourceClient.Delete(ctx, &pb.UUID{Value: resourceId[:]})
	return err
}

func (s *ResourceService) ListByUserId(ctx context.Context, rType api.ResourceType) ([]model.ShortResourceInfo, error) {
	stream, err := s.resourceClient.ListByUserId(ctx, &pb.Query{ResourceType: pb.TYPE(rType)})
	if err != nil {
		return nil, err
	}
	results := make([]model.ShortResourceInfo, 0)
	for {
		info, err := stream.Recv()
		if err == io.EOF {
			break
		}
		id, err := uuid.FromBytes(info.Id.Value)
		if err != nil {
			return nil, err
		}
		results = append(results, model.ShortResourceInfo{
			Id:   id,
			Meta: string(info.Meta),
		})
	}
	return results, nil
}

func (s *ResourceService) Get(ctx context.Context, id api.ResourceId) (model.Resource, []byte, error) {
	resource, err := s.resourceClient.Get(ctx, &pb.UUID{Value: id[:]})
	if err != nil {
		return nil, nil, err
	}
	switch api.ResourceType(resource.Type) {
	case api.LoginPassword:
		var lp model.LoginPassword
		if err := json.Unmarshal(resource.Data, &lp); err != nil {
			return nil, nil, err
		}

		return &lp, resource.Meta, nil

	case api.BankCard:
		var bc model.BankCard
		if err := json.Unmarshal(resource.Data, &bc); err != nil {
			return nil, nil, err
		}
		return &bc, resource.Meta, nil
	}
	return nil, nil, fmt.Errorf("undefined type %v", resource.Type)
}

func (s *ResourceService) SaveFile(ctx context.Context, description, path string) (api.ResourceId, error) {
	stream, err := s.resourceClient.SaveFile(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	err = stream.Send(&pb.FileChunk{
		Meta: []byte(description),
		Data: nil,
	})
	if err != nil {
		return uuid.Nil, err
	}
	err = fileutil.Send(path, func(bytes []byte) error {
		return stream.Send(&pb.FileChunk{
			Meta: nil,
			Data: bytes,
		})
	})
	if err != nil {
		return uuid.Nil, err
	}
	id, err := stream.CloseAndRecv()
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.FromBytes(id.Value)
}

func (s *ResourceService) GetFile(ctx context.Context, id api.ResourceId) (string, error) {
	stream, err := s.resourceClient.GetFile(ctx, &pb.UUID{Value: id[:]})
	if err != nil {
		return "", err
	}

	path := s.fileStorePath + "/" + id.String()
	err = fileutil.Get(path, func() ([]byte, error) {
		chunk, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		return chunk.Data, nil
	})
	return path, err
}
