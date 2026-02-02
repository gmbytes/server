package entity

import (
	"server/data"
	"server/data/enum"
	"server/lib/matrix"
	"server/lib/uid"
	"server/service/scene/entity/mod"
	"server/service/scene/score"
)

var _ score.IEntity = (*EntityBase)(nil)

type ManagerType = int

const (
	BattleManager ManagerType = iota
	SkillManager
	Max
)

type EntityBase struct {
	id    uid.Uid
	scene score.IScene
	ety   enum.EntityType
	pos   *matrix.Vector3D
	dir   int32

	managers [Max]score.IModule
}

func (e *EntityBase) Init(scene score.IScene, initData data.EntityInitData) {
	e.scene = scene
	e.id = uid.Gen()
	if e.scene != nil {
		e.scene.AddEntity(e)
	}
	if e.ety == enum.EntityType_Role || e.ety == enum.EntityType_Npc {
		e.managers[BattleManager] = &mod.BattleManager{}
		e.managers[SkillManager] = &mod.SkillManager{}
	}

	for _, m := range e.managers {
		if m == nil {
			continue
		}
		m.Init(e, initData)
	}
}

func (e *EntityBase) Update() {
	for _, m := range e.managers {
		if m == nil {
			continue
		}
		m.Update()
	}
}

func (e *EntityBase) GetId() uid.Uid {
	return e.id
}

func (e *EntityBase) GetScene() score.IScene {
	return e.scene
}

func (e *EntityBase) GetPos() *matrix.Vector3D {
	return e.pos
}

func (e *EntityBase) SetPos(pos *matrix.Vector3D) {
	e.pos = pos
}

func (e *EntityBase) GetDir() int32 {
	return e.dir
}

func (e *EntityBase) SetDir(dir int32) {
	e.dir = dir
}
