package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"sync"
)

var once = sync.Once{}
var db *sqlx.DB

func MustInitDB(ctx context.Context, url string) *sqlx.DB {
	var err error
	once.Do(func() {
		db, err = sqlx.ConnectContext(ctx, "postgres", url)
	})
	if err != nil {
		panic(err)
	}
	return db
}

func RunInTx(fs ...func(tx *sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	for i := 0; i < len(fs); i++ {
		if err := fs[i](tx); err != nil {
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}
	}
	return tx.Commit()
}
