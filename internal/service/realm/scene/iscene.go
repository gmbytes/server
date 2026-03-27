package scene

type SceneType = int

const (
	SceneTypeMMOMap   SceneType = 1
	SceneTypeInstance SceneType = 2
	SceneTypeRoom     SceneType = 3
	SceneTypeSLG      SceneType = 4
	SceneTypeArena    SceneType = 5
)

type IScene interface {
	Id() int64
	Type() SceneType
	Join(roleId int64, snapshot []byte) (initData []byte, err error)
	Leave(roleId int64, reason int) (carry *EntityCarryData, err error)
	OnMessage(roleId int64, key int, data []byte)
	Update(deltaMs int64)
	Destroy()
}

type EntityCarryData struct {
	Position []float32
	HP       int64
	Buffs    []byte
	Extra    []byte
}

type EntitySnapshot struct {
	RoleId     int64
	Name       string
	Level      int32
	Position   []float32
	Attributes map[int32]int64
	Buffs      []byte
	Skills     []byte
	Appearance []byte
	Extra      []byte
}

type LeaveResult struct {
	CarryData *EntityCarryData
	Reason    int
	Err       error
}

type JoinResult struct {
	InitData []byte
	Err      error
}

const (
	LeaveReasonSwitch  = 1
	LeaveReasonLogout  = 2
	LeaveReasonTimeout = 3
	LeaveReasonCross   = 4
	LeaveReasonKick    = 5
)
