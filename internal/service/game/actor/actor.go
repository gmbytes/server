package actor

import (
	"encoding/json"
	"sync"

	"server/internal/pb"
	"server/internal/service/realm/scene"
	"server/internal/service/realm/scenemgr"

	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[Actor, *Actor]("Actor")
}

type Actor struct {
	node.Service

	roleID   int64
	connId   uint64
	roleData *pb.RoleSummaryData

	sGame    node.IProxy
	sSceneMgr node.IProxy
	sSocial  node.IProxy
	sDB      node.IProxy

	sceneProxy node.IProxy
	curSceneId int64
}

func (a *Actor) Start(_ any) {
	a.sGame = a.CreateProxy("Game")
	a.sSceneMgr = a.CreateProxy("SceneMgr")
	a.sSocial = a.CreateProxy("Social")
	a.sDB = a.CreateProxy("DB")
	a.EnableRpc()
}

func (a *Actor) Stop(_ *sync.WaitGroup) {
	a.Infof("Actor stopping: roleID=%d", a.roleID)
}

func (a *Actor) AfterStop() {
	a.Infof("Actor stopped: roleID=%d", a.roleID)
}

// --------------- RPC: Game → Actor ---------------

func (a *Actor) RpcInit(ctx node.IRpcContext, roleID int64, connId uint64, roleData *pb.RoleSummaryData) {
	a.roleID = roleID
	a.connId = connId
	a.roleData = roleData
	a.Infof("Actor init: roleID=%d connId=%d", roleID, connId)
	ctx.Return()
}

func (a *Actor) RpcClientRequest(ctx node.IRpcContext, key uint16, body []byte) {
	a.Debugf("Actor recv: roleID=%d key=%d len=%d", a.roleID, key, len(body))

	k := pb.EKey_T(key)
	switch k {
	case pb.EKey_ReqPing:
		a.handlePing()
	case pb.EKey_ReqEnterZone:
		a.handleEnterScene()
	default:
		if a.sceneProxy != nil && a.curSceneId > 0 && a.sceneProxy.Avail() {
			a.sceneProxy.Call("SceneMessage", a.curSceneId, a.roleID, int(key), body).
				Catch(func(err error) {
					a.Errorf("scene message failed: %v", err)
				}).Done()
		} else {
			a.Debugf("Actor unhandled key=%d for roleID=%d (no scene)", key, a.roleID)
		}
	}

	ctx.Return()
}

func (a *Actor) RpcKick(ctx node.IRpcContext) {
	a.Infof("Actor kicked: roleID=%d", a.roleID)
	if a.sceneProxy != nil && a.curSceneId > 0 {
		if a.sceneProxy.Avail() {
			a.sceneProxy.Call("LeaveScene", a.curSceneId, a.roleID, int(scene.LeaveReasonKick)).
				Catch(func(err error) {
					a.Warnf("LeaveScene on kick failed: roleID=%d err=%v", a.roleID, err)
				}).
				Final(func() {
					a.sceneProxy = nil
					a.curSceneId = 0
				}).Done()
		} else {
			a.Warnf("sceneProxy unavailable on kick, skip LeaveScene: roleID=%d", a.roleID)
			a.sceneProxy = nil
			a.curSceneId = 0
		}
	}
	ctx.Return()
}

func (a *Actor) RpcRebind(ctx node.IRpcContext, newConnId uint64) {
	a.Infof("Actor rebind: roleID=%d oldConn=%d newConn=%d", a.roleID, a.connId, newConnId)
	a.connId = newConnId
	ctx.Return()
}

func (a *Actor) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{
		"status":     "Actor.OK",
		"roleID":     a.roleID,
		"connId":     a.connId,
		"curSceneId": a.curSceneId,
	})
}

// --------------- Ping ---------------

func (a *Actor) handlePing() {
	a.Infof("Actor handlePing: roleID=%d", a.roleID)
	a.sendToClient(pb.NewPackage(&pb.RspPing{}))
}

// --------------- Scene ---------------

func (a *Actor) handleEnterScene() {
	req := &scenemgr.AllocSceneReq{
		SceneType: scene.SceneTypeMMOMap,
		MapId:     1001,
	}
	a.switchScene(req)
}

func (a *Actor) switchScene(newReq *scenemgr.AllocSceneReq) {
	if a.sceneProxy != nil && a.curSceneId > 0 {
		if !a.sceneProxy.Avail() {
			a.Warnf("sceneProxy unavailable on switchScene, skip LeaveScene: roleID=%d", a.roleID)
			a.sceneProxy = nil
			a.curSceneId = 0
			a.doEnterNew(newReq, nil)
			return
		}
		a.sceneProxy.Call("LeaveScene", a.curSceneId, a.roleID, int(scene.LeaveReasonSwitch)).
			Then(func(ret *scene.LeaveResult) {
				a.sceneProxy = nil
				a.curSceneId = 0
				var carry *scene.EntityCarryData
				if ret != nil {
					carry = ret.CarryData
				}
				a.doEnterNew(newReq, carry)
			}).
			Catch(func(err error) {
				a.Errorf("leave scene failed: %v", err)
				a.sceneProxy = nil
				a.curSceneId = 0
				a.doEnterNew(newReq, nil)
			}).Done()
		return
	}
	a.doEnterNew(newReq, nil)
}

func (a *Actor) doEnterNew(req *scenemgr.AllocSceneReq, carry *scene.EntityCarryData) {
	if !a.sSceneMgr.Avail() {
		a.Warnf("SceneMgr unavailable for roleID=%d", a.roleID)
		return
	}

	a.sSceneMgr.Call("AllocScene", req).
		Then(func(info *scenemgr.SceneProxyInfo) {
			if info.Err != nil {
				a.Errorf("AllocScene failed: %v", info.Err)
				return
			}
			proxy := a.CreateProxyByNodeAddr(info.NodeAddr, info.SAddr)
			snapshot := a.buildSnapshotWithCarry(carry)
			proxy.Call("JoinScene", info.SceneId, a.roleID, snapshot).
				Then(func(join *scene.JoinResult) {
					if join != nil && join.Err != nil {
						a.Errorf("JoinScene failed: %v", join.Err)
						return
					}
					a.sceneProxy = proxy
					a.curSceneId = info.SceneId
					a.Infof("Actor %d entered scene %d", a.roleID, info.SceneId)
					a.sendToClient(pb.NewPackage(&pb.RspEnterZone{}))
				}).
				Catch(func(err error) {
					a.Errorf("JoinScene RPC failed: %v", err)
				}).Done()
		}).
		Catch(func(err error) {
			a.Errorf("AllocScene RPC failed: %v", err)
		}).Done()
}

func (a *Actor) buildSnapshotWithCarry(carry *scene.EntityCarryData) []byte {
	snap := &scene.EntitySnapshot{
		RoleId: a.roleID,
		Level:  int32(a.roleData.GetLv()),
		Name:   a.roleData.GetName(),
	}
	if carry != nil {
		snap.Position = carry.Position
		snap.Buffs = carry.Buffs
		snap.Extra = carry.Extra
	}
	data, _ := json.Marshal(snap)
	return data
}

// --------------- Response ---------------

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
