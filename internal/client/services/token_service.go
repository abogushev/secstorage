package services

type TokenService struct {
	token string
}

func (s *TokenService) Set(token string) {
	s.token = token
}

func (s *TokenService) Get() string {
	return s.token
}
