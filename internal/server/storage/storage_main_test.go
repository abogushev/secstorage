package storage

import (
	"context"
	"os"
	"secstorage/internal/server/storage/auth"
	"secstorage/internal/server/testutils"
	"testing"

	_ "github.com/lib/pq"
)

var AuthStorage *auth.Storage

func TestMain(m *testing.M) {
	var code int
	testutils.RunWithDBContainer(func(dbUrl string) {
		db := MustInitDB(context.Background(), dbUrl)
		AuthStorage = auth.NewStorage(context.Background(), db)
		code = m.Run()
	})

	os.Exit(code)
}
