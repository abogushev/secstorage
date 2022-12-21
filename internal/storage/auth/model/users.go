package model

import "github.com/google/uuid"

type User struct {
	Id       uuid.UUID `db:"id"`
	Login    string    `db:"login"`
	Password string    `db:"password"`
}
