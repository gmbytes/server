package score

import "server/lib/uid"

type IScene interface {
	Init()
	AddEntity(e IEntity)
	RemoveEntity(id uid.Uid)
	GetEntity(id uid.Uid) (IEntity, bool)
	ForEach(fn func(e IEntity))
}
