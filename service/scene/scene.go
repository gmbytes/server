package scene

import (
	"server/lib/uid"
	score2 "server/service/scene/score"
)

var _ score2.IScene = (*Scene)(nil)

type Scene struct {
	entities map[uid.Uid]score2.IEntity
}

func (ss *Scene) Init() {
	ss.entities = make(map[uid.Uid]score2.IEntity)
}

func (ss *Scene) AddEntity(e score2.IEntity) {
	if ss.entities == nil {
		ss.entities = make(map[uid.Uid]score2.IEntity)
	}
	if e == nil {
		return
	}
	ss.entities[e.GetId()] = e
}

func (ss *Scene) RemoveEntity(id uid.Uid) {
	if ss.entities == nil {
		return
	}
	delete(ss.entities, id)
}

func (ss *Scene) GetEntity(id uid.Uid) (score2.IEntity, bool) {
	if ss.entities == nil {
		return nil, false
	}
	e, ok := ss.entities[id]
	return e, ok
}

func (ss *Scene) ForEachEntity(fn func(id uid.Uid, e score2.IEntity)) {
	if ss.entities == nil || fn == nil {
		return
	}
	for id, e := range ss.entities {
		fn(id, e)
	}
}
