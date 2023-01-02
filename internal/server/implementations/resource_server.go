package implementations

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "secstorage/internal/api/proto"
	"secstorage/internal/server/storage/resource/model"
)

type ResourceService interface {
	Save(context.Context, *model.Resource) (model.ResourceId, error)
	Delete(context.Context, model.ResourceId) error
	ListByUserId(context.Context, model.UserId, model.ResourceType) ([]model.ShortResourceInfo, error)
	Get(context.Context, model.ResourceId) (*model.Resource, error)
}

type ResourceServer struct {
	pb.UnimplementedResourcesServer
	service ResourceService
}

func NewResourcesServer(service ResourceService) *ResourceServer {
	return &ResourceServer{
		service: service,
	}
}

func (s *ResourceServer) Save(ctx context.Context, resource *pb.Resource) (*pb.UUID, error) {
	id, err := s.service.Save(ctx, &model.Resource{
		UserId: extractUserId(ctx),
		Type:   model.ResourceType(resource.Type),
		Data:   resource.Data,
		Meta:   resource.Meta,
	})
	return &pb.UUID{Value: id[:]}, err
}

func (s *ResourceServer) Delete(ctx context.Context, id *pb.UUID) (*emptypb.Empty, error) {
	rId, err := uuid.FromBytes(id.Value)
	if err != nil {
		return nil, err
	}
	if err := s.service.Delete(ctx, rId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ResourceServer) ListByUserId(query *pb.Query, stream pb.Resources_ListByUserIdServer) error {
	t := model.ResourceType(query.ResourceType)
	userId := extractUserId(stream.Context())
	list, err := s.service.ListByUserId(stream.Context(), userId, t)
	if err != nil {
		return err
	}

	for i := 0; i < len(list); i++ {
		err := stream.Send(&pb.ShortResourceInfo{
			Id:   &pb.UUID{Value: list[i].Id[:]},
			Meta: list[i].Meta,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ResourceServer) Get(ctx context.Context, id *pb.UUID) (*pb.Resource, error) {
	rId, err := uuid.FromBytes(id.Value)
	if err != nil {
		return nil, err
	}
	result, err := s.service.Get(ctx, rId)
	if err != nil {
		return nil, err
	}
	return &pb.Resource{
		Type: pb.TYPE(result.Type),
		Data: result.Data,
		Meta: result.Meta,
	}, nil
}
