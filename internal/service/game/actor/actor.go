package actor

import (
	"sync"

	"server/internal/pb"

	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[Actor, *Actor]("Actor")
}

// Actor 代表一个玩家角色（role），由 Game 动态创建和销毁。
type Actor struct {
	node.Service

	roleID   int64
	connId   uint64
	roleData *pb.RoleSummaryData

	sGame node.IProxy
}

func (a *Actor) Start(_ any) {
	a.sGame = a.CreateProxy("Game")
	a.EnableRpc()
}

func (a *Actor) Stop(_ *sync.WaitGroup) {
	a.Infof("Actor stopping: roleID=%d", a.roleID)
}

func (a *Actor) AfterStop() {
	a.Infof("Actor stopped: roleID=%d", a.roleID)
}

// --------------- RPC: Game → Actor ---------------

// RpcInit 由 Game 在创建 Actor 后调用，初始化角色数据
func (a *Actor) RpcInit(ctx node.IRpcContext, roleID int64, connId uint64, roleData *pb.RoleSummaryData) {
	a.roleID = roleID
	a.connId = connId
	a.roleData = roleData
	a.Infof("Actor init: roleID=%d connId=%d", roleID, connId)
	ctx.Return()
}

// RpcClientRequest 处理来自客户端的业务请求（由 Game 转发）
func (a *Actor) RpcClientRequest(ctx node.IRpcContext, key uint16, body []byte) {
	a.Debugf("Actor recv: roleID=%d key=%d len=%d", a.roleID, key, len(body))

	k := pb.EKey_T(key)
	switch k {
	case pb.EKey_ReqEnterZone:
		a.handleEnterZone()
	default:
		a.Debugf("Actor unhandled key=%d for roleID=%d", key, a.roleID)
	}

	ctx.Return()
}

// RpcKick 由 Game 调用，通知 Actor 即将被关闭
func (a *Actor) RpcKick(ctx node.IRpcContext) {
	a.Infof("Actor kicked: roleID=%d", a.roleID)
	ctx.Return()
}

func (a *Actor) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{
		"status": "Actor.OK",
		"roleID": a.roleID,
		"connId": a.connId,
	})
}

// --------------- Business Handlers ---------------

func (a *Actor) handleEnterZone() {
	a.sendToClient(pb.NewPackage(&pb.RspEnterZone{}))
}

// --------------- Response ---------------

// sendToClient 将响应包通过 Game → Gate 发送给客户端
func (a *Actor) sendToClient(p *pb.Package) {
	data, err := p.Marshal()
	if err != nil {
		a.Errorf("marshal pkg failed: %v", err)
		return
	}
	if !a.sGame.Avail() {
		a.Warnf("Game proxy unavailable for roleID=%d", a.roleID)
		return
	}
	a.sGame.Call("ActorResponse", a.roleID, data).
		Catch(func(err error) {
			a.Errorf("send to client failed roleID=%d: %v", a.roleID, err)
		}).Done()
}
