package model

type Resource struct {
	Id     ResourceId   `db:"id"`
	UserId UserId       `db:"user_id"`
	Type   ResourceType `db:"type"`
	Data   []byte       `db:"data"`
	Meta   []byte       `db:"meta"`
}
