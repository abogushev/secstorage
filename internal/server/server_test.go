package server

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
	"os"
	pb "secstorage/internal/api/proto"
	"secstorage/internal/server/auth"
	"secstorage/internal/server/auth/services"
	"secstorage/internal/storage"
	authStorage "secstorage/internal/storage/auth"
	"secstorage/internal/testutils"
	"testing"
)

var client pb.AuthClient
var db *sqlx.DB

func initServerAndClient(db *sqlx.DB) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	storage := authStorage.NewAuthStorage(context.Background(), db)
	authService := services.NewAuthService(storage)
	tokenService := auth.NewTokenService("7+P+BBqjUvY6NF0jGU9JVWurFULGLbDWPWBRVK6MCpvCHkU1aPAA/gm4t0xKTNGxbQdJvUXMa89rGQCur1z5rw==")

	go Run(context.Background(), authService, tokenService, insecure.NewCredentials(), lis)

	conn, err := grpc.DialContext(context.Background(), "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("error connecting to server: %v", err)
	}

	go func() {
		<-context.Background().Done()
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
	}()

	client = pb.NewAuthClient(conn)
}

func TestMain(m *testing.M) {
	var code int
	testutils.RunWithDBContainer(func(dbUrl string) {
		db = storage.MustInitDB(context.Background(), dbUrl)
		initServerAndClient(db)
		code = m.Run()
	})

	os.Exit(code)
}

func prepare() {
	db.MustExec("truncate table users")
}

var testAuthData = &pb.AuthData{
	Login:    "login",
	Password: "password",
}

func TestAuthServer_Register_Success(t *testing.T) {
	prepare()

	token, err := client.Register(context.Background(), testAuthData)
	assert.NoError(t, err)
	assert.True(t, len(token.Token) != 0)
}

func TestAuthServer_Register_UserAlreadyExist(t *testing.T) {
	prepare()

	_, _ = client.Register(context.Background(), testAuthData)
	_, err := client.Register(context.Background(), testAuthData)

	assert.ErrorIs(t, err, status.Error(codes.AlreadyExists, "user already exist"))
}

func TestAuthServer_Login_Success(t *testing.T) {
	prepare()

	_, _ = client.Register(context.Background(), testAuthData)
	token, err := client.Login(context.Background(), testAuthData)
	assert.NoError(t, err)
	assert.NotEmpty(t, token.Token)
}

func TestAuthServer_Login_UserNotFound(t *testing.T) {
	prepare()

	_, err := client.Login(context.Background(), testAuthData)
	assert.ErrorIs(t, err, status.Error(codes.NotFound, "user not found"))
}
