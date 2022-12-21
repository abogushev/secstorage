package server

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		db := storage.MustInitDB(context.Background(), dbUrl)
		initServerAndClient(db)
		code = m.Run()
	})

	os.Exit(code)
}

func TestAuthServer_Register(t *testing.T) {
	ctx := context.Background()
	token, err := client.Register(ctx, &pb.AuthData{
		Login:    "login",
		Password: "password",
	})
	token2, err2 := client.Register(ctx, &pb.AuthData{
		Login:    "login",
		Password: "password",
	})
	fmt.Println(token, err, token2, err2)
	fmt.Println(token2, err2)
	//fmt.Println(token, err)
}
