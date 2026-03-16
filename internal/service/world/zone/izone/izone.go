package izone

import "server/pkg/uid"

type IZone interface {
	Init()
	AddEntity(e IEntity)
	RemoveEntity(id uid.Uid)
	GetEntity(id uid.Uid) (IEntity, bool)
	ForEach(fn func(e IEntity))
}
