package score

import "server/data"

type IModule interface {
	Init(onwer IEntity, initData data.EntityInitData)
	Update(duration int64)
}
