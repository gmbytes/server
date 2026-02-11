package entity

import (
	"server/data"
	"server/data/enum"
	"server/lib/uid"
	"server/pb"
	"server/service/world/zone/entity/mod/combat"
	"server/service/world/zone/izone"
)

var _ izone.IEntity = (*EntityBase)(nil)

type ManagerType = int

const (
	CombatManager ManagerType = iota
	Max
)

type EntityBase struct {
	id   uid.Uid
	zone izone.IZone
	ety  enum.EntityType
	pos  *pb.Vector
	dir  int32

	managers [Max]izone.IModule
}

func (e *EntityBase) Init(zone izone.IZone, initData data.EntityInitData) {
	e.zone = zone
	e.id = uid.Gen()
	if e.zone != nil {
		e.zone.AddEntity(e)
	}
	if e.ety == enum.EntityType_Role || e.ety == enum.EntityType_Npc {
		e.managers[CombatManager] = &combat.CombatManager{}

	}

	for _, m := range e.managers {
		if m == nil {
			continue
		}
		m.Init(e, initData)
	}
}

func (e *EntityBase) Update(duration int64) {
	for _, m := range e.managers {
		if m == nil {
			continue
		}
		m.Update(duration)
	}
}

func (e *EntityBase) GetId() uid.Uid {
	return e.id
}

func (e *EntityBase) GetZone() izone.IZone {
	return e.zone
}

func (e *EntityBase) GetPos() *pb.Vector {
	return e.pos
}

func (e *EntityBase) SetPos(pos *pb.Vector) {
	e.pos = pos
}

func (e *EntityBase) GetDir() int32 {
	return e.dir
}

func (e *EntityBase) SetDir(dir int32) {
	e.dir = dir
}
