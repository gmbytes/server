package score

import (
	"server/data"
	"server/lib/matrix"
	"server/lib/uid"
)

type IEntity interface {
	Init(scene IScene, initData data.EntityInitData)

	GetScene() IScene

	GetId() uid.Uid
	GetPos() *matrix.Vector3D
	SetPos(pos *matrix.Vector3D)
	GetDir() int32
	SetDir(dir int32)
}
