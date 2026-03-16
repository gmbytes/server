package izone

import (
	"server/internal/data"
	"server/internal/pb"
	"server/pkg/uid"
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
