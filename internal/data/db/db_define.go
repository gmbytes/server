package db

import (
	"server/pkg/uid"
)

type RoleBase struct {
	Id      uid.Uid `json:",omitempty" bson:",omitempty"`
	Cid     int64   `json:",omitempty" bson:",omitempty"`
	Lv      int64   `json:",omitempty" bson:",omitempty"`
	Account string  `json:",omitempty" bson:",omitempty"` // 角色账户
	Name    string  `json:",omitempty" bson:",omitempty"`
	Version int64   `json:",omitempty" bson:",omitempty"`
}
