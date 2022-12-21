package auth

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"time"
)

type TokenService struct {
	key string
}

type authClaims struct {
	Id uuid.UUID `json:"id"`
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

func (s *TokenService) Extract(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.key), nil
	})

	if claims, ok := token.Claims.(*authClaims); ok && token.Valid {
		return claims.Id, nil
	}
	return uuid.Nil, err
	//if token.Valid {
	//	fmt.Println("You look nice today")
	//} else if errors.Is(err, jwt.ErrTokenMalformed) {
	//	fmt.Println("That's not even a token")
	//} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
	//	// Token is either expired or not active yet
	//	fmt.Println("Timing is everything")
	//} else {
	//	fmt.Println("Couldn't handle this token:", err)
	//}
}
