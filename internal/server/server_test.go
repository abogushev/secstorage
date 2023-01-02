package server

import (
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"io"
	"log"
	"net"
	"os"
	pb "secstorage/internal/api/proto"
	"secstorage/internal/server/implementations"
	services "secstorage/internal/server/services"
	"secstorage/internal/server/storage"
	authStorage "secstorage/internal/server/storage/auth"
	resourceStorage "secstorage/internal/server/storage/resource"
	"secstorage/internal/server/storage/resource/model"
	"secstorage/internal/server/testutils"
	"testing"
)

var authClient pb.AuthClient
var resourceClient pb.ResourcesClient
var db *sqlx.DB

var TokenService = services.NewTokenService("7+P+BBqjUvY6NF0jGU9JVWurFULGLbDWPWBRVK6MCpvCHkU1aPAA/gm4t0xKTNGxbQdJvUXMa89rGQCur1z5rw==")
var testResource = &pb.Resource{
	Type: 1,
	Data: []byte("data"),
	Meta: []byte("meta"),
}

func initServerAndClient(db *sqlx.DB) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	authStore := authStorage.NewStorage(context.Background(), db)
	authService := services.NewAuthService(authStore)
	authServer := implementations.NewAuthServer(authService, TokenService)

	resourceStore := resourceStorage.NewStore(context.Background(), db)
	resourceService := services.NewResourceStoreService(resourceStore)
	resourceServer := implementations.NewResourcesServer(resourceService)

	go Run(context.Background(), authServer, resourceServer, TokenService, insecure.NewCredentials(), lis)

	con, err := grpc.DialContext(context.Background(), "",
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

	authClient = pb.NewAuthClient(con)
	resourceClient = pb.NewResourcesClient(con)
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
	db.MustExec("truncate table users cascade")
}

var testAuthData = &pb.AuthData{
	Login:    "login",
	Password: "password",
}

func TestAuthServer_Register_Success(t *testing.T) {
	prepare()

	token, err := authClient.Register(context.Background(), testAuthData)
	assert.NoError(t, err)
	assert.True(t, len(token.Token) != 0)
}

func TestAuthServer_Register_UserAlreadyExist(t *testing.T) {
	prepare()

	_, _ = authClient.Register(context.Background(), testAuthData)
	_, err := authClient.Register(context.Background(), testAuthData)

	assert.ErrorIs(t, err, status.Error(codes.AlreadyExists, "user already exist"))
}

func TestAuthServer_Login_Success(t *testing.T) {
	prepare()

	_, _ = authClient.Register(context.Background(), testAuthData)
	token, err := authClient.Login(context.Background(), testAuthData)
	assert.NoError(t, err)
	assert.NotEmpty(t, token.Token)
}

func TestAuthServer_Login_UserNotFound(t *testing.T) {
	prepare()

	_, err := authClient.Login(context.Background(), testAuthData)
	assert.ErrorIs(t, err, status.Error(codes.NotFound, "user not found"))
}

func TestResourceServer_Save_and_Get_Success(t *testing.T) {
	prepare()

	token, err := authClient.Register(context.Background(), testAuthData)
	assert.NoError(t, err)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"token": token.Token}))
	resource := &pb.Resource{
		Type: 1,
		Data: []byte("data"),
		Meta: []byte("meta"),
	}
	id, err := resourceClient.Save(ctx, resource)
	assert.NoError(t, err)

	result, err := resourceClient.Get(ctx, id)
	assert.NoError(t, err)

	assert.Equal(t, resource.Type, result.Type)
	assert.Equal(t, resource.Data, result.Data)
	assert.Equal(t, resource.Meta, result.Meta)
}

func TestResourceServer_List_And_Delete_Success(t *testing.T) {
	prepare()
	token, err := authClient.Register(context.Background(), testAuthData)
	assert.NoError(t, err)
	userId, err := TokenService.Extract(token.Token)
	assert.NoError(t, err)

	var rId1 model.ResourceId
	err = db.QueryRowContext(
		context.Background(),
		"insert into resources(id, user_id, type, data, meta) values (gen_random_uuid(), $1,$2,$3,$4) returning id",
		userId,
		testResource.Type,
		testResource.Data,
		testResource.Meta,
	).Scan(&rId1)
	assert.NoError(t, err)
	var rId2 model.ResourceId
	err = db.QueryRowContext(
		context.Background(),
		"insert into resources(id, user_id, type, data, meta) values (gen_random_uuid(), $1,$2,$3,$4) returning id",
		userId,
		testResource.Type,
		testResource.Data,
		testResource.Meta,
	).Scan(&rId2)
	assert.NoError(t, err)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"token": token.Token}))
	stream, err := resourceClient.ListByUserId(ctx, &pb.Query{ResourceType: testResource.Type})

	expecteds := map[model.ResourceId]*pb.Resource{rId1: testResource, rId2: testResource}
	result := make([]*pb.ShortResourceInfo, 0, 2)

	for {
		shortInfo, err := stream.Recv()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		result = append(result, shortInfo)

	}
	for i := 0; i < len(result); i++ {
		id, err := uuid.FromBytes(result[i].Id.Value)
		assert.NoError(t, err)
		exp := expecteds[id]
		assert.Equal(t, exp.Meta, result[i].Meta)
	}

	_, err = resourceClient.Delete(ctx, &pb.UUID{Value: rId1[:]})
	assert.NoError(t, err)
	var c int
	err = db.GetContext(ctx, &c, "select count(*) from resources where id = $1", rId1)
	assert.NoError(t, err)
	assert.Equal(t, 0, c)
}
