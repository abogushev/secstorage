package services

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"secstorage/internal/api"
	"secstorage/internal/server/reservederrors"
	"time"
)

type TokenService struct {
	key string
}

type authClaims struct {
	Id api.UserId `json:"id"`
	jwt.RegisteredClaims
}

func NewTokenService(key string) *TokenService {
	return &TokenService{key}
}

func (s *TokenService) Generate(id uuid.UUID, expireAt time.Time) (string, error) {
	claims := &authClaims{
		Id:               id,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expireAt)},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.key))
}

func (s *TokenService) Extract(tokenStr string) (api.UserId, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.key), nil
	})

	if claims, ok := token.Claims.(*authClaims); ok && token.Valid {
		return claims.Id, nil
	}
	return uuid.Nil, err
}

func (s *TokenService) GetUserIdGRPC(ctx context.Context) (api.UserId, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, errors.New("can't read md")
	}
	var tokenStr string
	if values := md.Get("token"); len(values) == 0 {
		return uuid.Nil, reservederrors.ErrTokenNotFound
	} else {
		tokenStr = values[0]
	}

	return s.Extract(tokenStr)
}
