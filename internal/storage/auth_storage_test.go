package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"secstorage/internal/storage/auth/model"
	"testing"
)

func prepare() {
	db.MustExec("truncate table users")
}

func TestRegister_Success(t *testing.T) {
	prepare()
	id, err := AuthStorage.Register(context.TODO(), model.User{Login: "login", Password: "password"})
	assert.NoError(t, err)
	c := 0
	assert.NoError(t, db.Get(&c, "select count(*) from users where id = $1", id))
	assert.Equal(t, 1, c)
}

func TestRegister_DuplicateException(t *testing.T) {
	prepare()
	_, err1 := AuthStorage.Register(context.TODO(), model.User{Login: "login", Password: "password"})
	_, err2 := AuthStorage.Register(context.TODO(), model.User{Login: "login", Password: "password"})
	assert.NoError(t, err1)
	assert.ErrorIs(t, err2, model.ErrUserAlreadyExist)
}

func TestLogin_Success(t *testing.T) {
	prepare()
	user := model.User{Id: uuid.New(), Login: "login", Password: "password"}
	db.MustExec("insert into users(id, login, password) values ($1,$2,$3)", user.Id, user.Login, user.Password)
	id, err := AuthStorage.Login(context.TODO(), user)
	assert.NoError(t, err)
	assert.Equal(t, user.Id, id)
}

func TestLogin_UserNotFoundException(t *testing.T) {
	prepare()
	user := model.User{Id: uuid.New(), Login: "login", Password: "password"}
	_, err := AuthStorage.Login(context.TODO(), user)
	assert.ErrorIs(t, err, model.ErrUserNotFound)
}
