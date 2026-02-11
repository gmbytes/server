package pb

type ReqCastSkill struct {
	Cid        int64
	SubCid     int64
	Pos        *Vector
	Dir        *Vector
	LockTarget int64
}
