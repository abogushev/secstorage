package auth

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "secstorage/internal/api/proto"
	. "secstorage/internal/logger"
	"secstorage/internal/storage/auth/model"
	"time"
)

type Service interface {
	Register(ctx context.Context, info model.User) (uuid.UUID, error)
	Login(ctx context.Context, info model.User) (uuid.UUID, error)
}

type Server struct {
	pb.UnimplementedAuthServer
	authService  Service
	tokenService *TokenService
}

func NewServer(authService Service, tokenService *TokenService) *Server {
	return &Server{authService: authService, tokenService: tokenService}
}

func (s *Server) Register(ctx context.Context, authData *pb.AuthData) (*pb.TokenData, error) {
	if err := validateAuthData(authData); err != nil {
		return nil, err
	}

	User := model.User{Login: authData.Login, Password: authData.Password}
	id, err := s.authService.Register(ctx, User)

	if err == nil {
		return s.genToken(id)
	}
	if errors.Is(err, model.ErrUserAlreadyExist) {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return nil, status.Error(codes.Internal, "internal error")
}

func (s *Server) Login(ctx context.Context, authData *pb.AuthData) (*pb.TokenData, error) {
	if err := validateAuthData(authData); err != nil {
		return nil, err
	}
	User := model.User{Login: authData.Login, Password: authData.Password}
	id, err := s.authService.Login(ctx, User)
	if err == nil {
		return s.genToken(id)
	}
	if errors.Is(err, model.ErrUserNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return nil, status.Error(codes.Internal, "internal error")
}

func validateAuthData(authData *pb.AuthData) error {
	if len(authData.Login) == 0 || len(authData.Password) == 0 {
		return status.Error(codes.InvalidArgument, "invalid login/password format: must be nonempty")
	}
	return nil
}

func (s *Server) genToken(id uuid.UUID) (*pb.TokenData, error) {
	expireAt := time.Now().UTC().Add(time.Hour)
	token, err := s.tokenService.Generate(id, expireAt)
	if err != nil {
		Log.Error("error on register", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	Log.Info("generate token successfully", zap.Time("expireAt", expireAt))
	return &pb.TokenData{Token: token, ExpireAt: timestamppb.New(expireAt)}, nil
}
