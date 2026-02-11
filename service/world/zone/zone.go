package zone

import (
	"server/lib/container"
	"server/lib/uid"
	"server/service/world/zone/izone"
)

var _ izone.IZone = (*Zone)(nil)

type Zone struct {
	entities *container.LMap[uid.Uid, izone.IEntity]
}

func (ss *Zone) Init() {
	ss.entities = container.NewLMap[uid.Uid, izone.IEntity]()
}

func (ss *Zone) AddEntity(e izone.IEntity) {
	if e == nil {
		return
	}
	ss.entities.Set(e.GetId(), e)
}

func (ss *Zone) RemoveEntity(id uid.Uid) {
	ss.entities.Delete(id)
}

func (ss *Zone) GetEntity(id uid.Uid) (izone.IEntity, bool) {
	return ss.entities.Get(id)
}

func (ss *Zone) ForEach(fn func(e izone.IEntity)) {
	if fn == nil {
		return
	}
	ss.entities.ForEach(fn)
}
