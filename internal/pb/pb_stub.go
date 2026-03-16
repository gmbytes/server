package pb

import (
	"encoding/binary"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type EKey_T uint16

const (
	EKey_Login EKey_T = iota + 1
	EKey_CreateRole
	EKey_LoginRole
	EKey_EnterZone
	EKey_Ping
	EKey_KickRole
)

var EKey_T_name = map[int32]string{
	int32(EKey_Login):      "Login",
	int32(EKey_CreateRole): "CreateRole",
	int32(EKey_LoginRole):  "LoginRole",
	int32(EKey_EnterZone):  "EnterZone",
	int32(EKey_Ping):       "Ping",
	int32(EKey_KickRole):   "KickRole",
}

type EErrorCode_T uint16

const (
	EErrorCode_Ok EErrorCode_T = 0
)

type Vector struct {
	X float64
	Y float64
	Z float64
}

type ReqLogin struct {
	Fast      bool
	Timestamp int64
	SId       int64
	Account   string
	Username  string
	ChannelId string
	Platform  string
	Appid     string
	Version   string
}

func (x *ReqLogin) Reset()         { *x = ReqLogin{} }
func (x *ReqLogin) String() string { return "ReqLogin" }
func (*ReqLogin) ProtoMessage()    {}
func (*ReqLogin) ProtoReflect() protoreflect.Message {
	return nil
}

type ReqPing struct{}

func (x *ReqPing) Reset()         { *x = ReqPing{} }
func (x *ReqPing) String() string { return "ReqPing" }
func (*ReqPing) ProtoMessage()    {}
func (*ReqPing) ProtoReflect() protoreflect.Message {
	return nil
}

type ReqCastSkill struct {
	LockTarget int64
	Pos        *Vector
}

func (x *ReqCastSkill) Reset()         { *x = ReqCastSkill{} }
func (x *ReqCastSkill) String() string { return "ReqCastSkill" }
func (*ReqCastSkill) ProtoMessage()    {}
func (*ReqCastSkill) ProtoReflect() protoreflect.Message {
	return nil
}

type RoleSummaryData struct{}

func (x *RoleSummaryData) Reset()         { *x = RoleSummaryData{} }
func (x *RoleSummaryData) String() string { return "RoleSummaryData" }
func (*RoleSummaryData) ProtoMessage()    {}
func (*RoleSummaryData) ProtoReflect() protoreflect.Message {
	return nil
}

type Package struct {
	KeyCode      EKey_T
	SerialNumber uint32
	Content      []byte
}

func (x *Package) Reset()         { *x = Package{} }
func (x *Package) String() string { return "Package" }
func (*Package) ProtoMessage()    {}
func (*Package) ProtoReflect() protoreflect.Message {
	return nil
}

func (x *Package) Key() EKey_T {
	if x == nil {
		return 0
	}
	return x.KeyCode
}

func (x *Package) Bytes() ([]byte, error) {
	if x == nil {
		return nil, nil
	}
	bs := make([]byte, 10+len(x.Content))
	binary.LittleEndian.PutUint16(bs[0:2], uint16(x.KeyCode))
	binary.LittleEndian.PutUint32(bs[2:6], uint32(4+len(x.Content)))
	binary.LittleEndian.PutUint32(bs[6:10], x.SerialNumber)
	copy(bs[10:], x.Content)
	return bs, nil
}

func Unmarshal(key EKey_T, body []byte) *Package {
	return &Package{
		KeyCode: key,
		Content: append([]byte(nil), body...),
	}
}
