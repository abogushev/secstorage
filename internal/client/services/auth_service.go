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

type TokenServiceSetter interface {
	Set(string)
}

type AuthService struct {
	authClient       pb.AuthClient
	refreshTokenOnce sync.Once
	tokenService     TokenServiceSetter
}

func NewAuthService(cl pb.AuthClient, tokenService TokenServiceSetter) *AuthService {
	return &AuthService{authClient: cl, tokenService: tokenService}
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
	s.tokenService.Set(tokenData.Token)

	go s.refreshToken(login, password, tokenData.ExpireAt.AsTime())

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

	s.tokenService.Set(tokenData.Token)

	go s.refreshToken(login, password, tokenData.ExpireAt.AsTime())

	return tokenData, nil
}

func (s *AuthService) refreshToken(login, password string, expiredAt time.Time) {
	calcRefreshTime := func(expiredAt time.Time) time.Duration {
		return expiredAt.Sub(time.Now().UTC()) / 2
	}

	s.refreshTokenOnce.Do(func() {
		timer := time.NewTimer(calcRefreshTime(expiredAt))
		for {
			select {
			case <-context.Background().Done():
				Log.Info("refresh token canceled")
				timer.Stop()
				return

			case <-timer.C:
				Log.Info("start refreshing token...")
				token, err := s.Login(context.Background(), login, password)
				if err != nil {
					Log.Error("failed to refresh token", zap.Error(err))
					timer.Reset(2 * time.Second)
				} else {
					s.tokenService.Set(token.Token)
					Log.Info("token refreshed successful")
					timer.Reset(calcRefreshTime(token.ExpireAt.AsTime()))
				}
			}
		}
	})
}
