package storage

import (
	"context"
	"os"
	"secstorage/internal/storage/auth"
	"secstorage/internal/testutils"
	"testing"

	_ "github.com/lib/pq"
)

var AuthStorage *auth.Storage

func TestMain(m *testing.M) {
	var code int
	testutils.RunWithDBContainer(func(dbUrl string) {
		db := MustInitDB(context.Background(), dbUrl)
		AuthStorage = auth.NewAuthStorage(context.Background(), db)
		code = m.Run()
	})

	os.Exit(code)
}
