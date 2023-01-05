package modulservers

import (
	"bufio"
	"context"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"secstorage/internal/api"
	pb "secstorage/internal/api/proto"
	"secstorage/internal/server/services"
	"secstorage/internal/server/storage/resource/model"
)

type ResourceService interface {
	Save(context.Context, *model.Resource) error
	Delete(context.Context, api.ResourceId, api.UserId) error
	ListByUserId(context.Context, api.UserId, api.ResourceType) ([]model.ShortResourceInfo, error)
	Get(context.Context, api.ResourceId, api.UserId) (*model.Resource, error)
	SaveFile(context.Context, api.UserId, []byte) (api.ResourceId, *bufio.Writer, services.Close, error)
	GetFile(context.Context, api.ResourceId, api.UserId) (*bufio.Reader, []byte, services.Close, error)
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
	id := uuid.New()
	err := s.service.Save(ctx, &model.Resource{
		Id:     id,
		UserId: extractUserId(ctx),
		Type:   api.ResourceType(resource.Type),
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
	if err := s.service.Delete(ctx, rId, extractUserId(ctx)); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ResourceServer) ListByUserId(query *pb.Query, stream pb.Resources_ListByUserIdServer) error {
	t := api.ResourceType(query.ResourceType)
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
	result, err := s.service.Get(ctx, rId, extractUserId(ctx))
	if err != nil {
		return nil, err
	}
	return &pb.Resource{
		Type: pb.TYPE(result.Type),
		Data: result.Data,
		Meta: result.Meta,
	}, nil
}

func (s *ResourceServer) SaveFile(stream pb.Resources_SaveFileServer) error {
	var writer io.Writer
	var closeWriter services.Close
	defer closeWriter()

	var id *pb.UUID
	for {
		chunk, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(id)
		}
		if err != nil {
			return err
		}
		if writer == nil {
			var rId api.ResourceId
			rId, writer, closeWriter, err = s.service.SaveFile(stream.Context(), extractUserId(stream.Context()), chunk.Meta)
			id = &pb.UUID{Value: rId[:]}
			if err != nil {
				return err
			}
		}
		_, err = writer.Write(chunk.Data)
		if err != nil {
			return err
		}
	}
}

func (s *ResourceServer) GetFile(id *pb.UUID, stream pb.Resources_GetFileServer) error {
	rId, err := uuid.FromBytes(id.Value)
	if err != nil {
		return err
	}
	reader, meta, closeReader, err := s.service.GetFile(stream.Context(), rId, extractUserId(stream.Context()))
	if err != nil {
		return err
	}
	defer closeReader()
	err = stream.Send(&pb.FileChunk{
		Meta: meta,
		Data: nil,
	})
	if err != nil {
		return err
	}

	buffer := make([]byte, 4096)
	n := 0

	for {
		n, err = reader.Read(buffer)
		if err == io.EOF || n == 0 {
			return nil
		}
		if err != nil {
			return err
		}

		err = stream.Send(&pb.FileChunk{
			Meta: nil,
			Data: buffer[:n],
		})
		if err != nil {
			return err
		}
	}
}
