package model

type ShortResourceInfo struct {
	Id   ResourceId `db:"id"`
	Meta []byte     `db:"meta"`
}
