package testutils

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"log"
	"path"
	"runtime"
)

func RunWithDBContainer(testRun func(dbUrl string)) {
	dbUrl, container := SetupTestDatabase()
	defer func() { _ = container.Terminate(context.Background()) }()
	testRun(dbUrl)
}

func SetupTestDatabase() (string, testcontainers.Container) {
	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := path.Join(path.Dir(filename), "../../migrations")

	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_USER":     "postgres",
		},

		Mounts: testcontainers.Mounts(testcontainers.ContainerMount{
			Source:   testcontainers.GenericBindMountSource{HostPath: migrationsPath},
			Target:   "/docker-entrypoint-initdb.d",
			ReadOnly: true,
		}),
	}

	dbContainer, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})

	if err != nil {
		log.Fatal(err)
	}

	host, _ := dbContainer.Host(context.Background())
	port, _ := dbContainer.MappedPort(context.Background(), "5432")

	url := fmt.Sprintf("postgres://postgres:postgres@%v:%v/testdb?sslmode=disable", host, port.Port())

	return url, dbContainer
}
