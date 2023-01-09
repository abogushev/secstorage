package model

import "secstorage/internal/api"

type ShortResourceInfo struct {
	Id   api.ResourceId `db:"id"`
	Meta []byte         `db:"meta"`
}
