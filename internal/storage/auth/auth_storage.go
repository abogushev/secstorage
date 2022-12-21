package auth

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"secstorage/internal/storage/auth/model"
)

type Storage struct {
	ctx context.Context
	db  *sqlx.DB
}

func NewAuthStorage(ctx context.Context, db *sqlx.DB) *Storage {
	return &Storage{ctx: ctx, db: db}
}

func (s *Storage) Register(ctx context.Context, user model.User) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.db.QueryRowContext(ctx, "insert into users (id, login, password) values (gen_random_uuid(), $1, $2) returning id", user.Login, user.Password).Scan(&id)

	if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == "23505" {
		return uuid.Nil, model.ErrUserAlreadyExist
	}
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (s *Storage) Login(ctx context.Context, user model.User) (uuid.UUID, error) {
	id := uuid.Nil
	err := s.db.GetContext(ctx, &id, "select id from users where login = $1 and password = $2", user.Login, user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, model.ErrUserNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
