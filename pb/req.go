package pb

import "server/lib/matrix"

type ReqCastSkill struct {
	Cid        int64
	SubCid     int64
	Pos        *matrix.Vector3D
	Dir        *matrix.Vector3D
	LockTarget int64
}
