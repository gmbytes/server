package izone

import (
	"server/data"
	"server/lib/uid"
	"server/pb"
)

type IEntity interface {
	Init(zone IZone, initData data.EntityInitData)

	GetZone() IZone

	GetId() uid.Uid
	GetPos() *pb.Vector
	SetPos(pos *pb.Vector)
	GetDir() int32
	SetDir(dir int32)
}
