package scene

import (
	"server/lib/container"
	"server/lib/uid"
	"server/service/scene/score"
)

var _ score.IScene = (*Scene)(nil)

type Scene struct {
	entities *container.LMap[uid.Uid, score.IEntity]
}

func (ss *Scene) Init() {
	ss.entities = container.NewLMap[uid.Uid, score.IEntity]()
}

func (ss *Scene) AddEntity(e score.IEntity) {
	if e == nil {
		return
	}
	ss.entities.Set(e.GetId(), e)
}

func (ss *Scene) RemoveEntity(id uid.Uid) {
	ss.entities.Delete(id)
}

func (ss *Scene) GetEntity(id uid.Uid) (score.IEntity, bool) {
	return ss.entities.Get(id)
}

func (ss *Scene) ForEach(fn func(e score.IEntity)) {
	if fn == nil {
		return
	}
	ss.entities.ForEach(fn)
}
