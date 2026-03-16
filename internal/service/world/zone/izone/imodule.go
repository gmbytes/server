package izone

import "server/internal/data"

type IModule interface {
	Init(owner IEntity, initData data.EntityInitData)
	Update(duration int64)
}
