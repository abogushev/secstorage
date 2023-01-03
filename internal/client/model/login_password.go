package model

import "fmt"

type LoginPassword struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

func NewLoginPassword(login string, password string) *LoginPassword {
	return &LoginPassword{Login: login, Password: password}
}

func (p *LoginPassword) Print(description string) string {
	return fmt.Sprintf("login:%v\npassword:%v\ndescription:%v", p.Login, p.Password, description)
}
