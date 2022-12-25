package services

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "secstorage/internal/api/proto"
	. "secstorage/internal/logger"
	"sync"
	"time"
)

type AuthService struct {
	authClient       pb.AuthClient
	token            *pb.TokenData
	refreshTokenOnce sync.Once
}

func NewAuthService(cl pb.AuthClient) *AuthService {
	return &AuthService{authClient: cl}
}

func (s *AuthService) Register(ctx context.Context, login, password string) (*pb.TokenData, error) {
	tokenData, err := s.authClient.Register(ctx, &pb.AuthData{
		Login:    login,
		Password: password,
	})

	if err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.AlreadyExists {
			return nil, errors.New(e.Message())
		}
		Log.Error("register failed", zap.Error(err))
		return nil, err
	}
	s.token = tokenData

	go s.refreshToken(login, password)

	return tokenData, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (*pb.TokenData, error) {
	tokenData, err := s.authClient.Login(ctx, &pb.AuthData{
		Login:    login,
		Password: password,
	})

	if err != nil {
		if e, ok := status.FromError(err); ok && e.Code() == codes.NotFound {
			return nil, errors.New(e.Message())
		}
		Log.Error("login failed", zap.Error(err))
		return nil, err
	}

	s.token = tokenData

	go s.refreshToken(login, password)

	return tokenData, nil
}

func (s *AuthService) refreshToken(login, password string) {
	s.refreshTokenOnce.Do(func() {
		ticker := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-context.Background().Done():
				Log.Info("refresh token canceled")
				return
			case <-ticker.C:
				Log.Info("start refreshing token...")
				token, err := s.Login(context.Background(), login, password)
				if err != nil {
					Log.Error("failed to refresh token", zap.Error(err))
				} else {
					s.token = token
					Log.Info("token refreshed successful")
				}
			}
		}
	})
}
