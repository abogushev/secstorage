package services

import (
	"context"
	"github.com/google/uuid"
	"secstorage/internal/server/storage/auth/model"
)

type AuthStorage interface {
	Register(context.Context, model.User) (uuid.UUID, error)
	Login(context.Context, model.User) (uuid.UUID, error)
}

type AuthService struct {
	storage AuthStorage
	ctx     context.Context
}

func NewAuthService(storage AuthStorage) *AuthService {
	return &AuthService{storage: storage}
}

func (s *AuthService) Register(ctx context.Context, info model.User) (uuid.UUID, error) {
	return s.storage.Register(ctx, info)
}

func (s *AuthService) Login(ctx context.Context, info model.User) (uuid.UUID, error) {
	return s.storage.Login(ctx, info)
}
