package model

import "secstorage/internal/api"

type Resource struct {
	Id     api.ResourceId   `db:"id"`
	UserId api.UserId       `db:"user_id"`
	Type   api.ResourceType `db:"type"`
	Data   []byte           `db:"data"`
	Meta   []byte           `db:"meta"`
}
