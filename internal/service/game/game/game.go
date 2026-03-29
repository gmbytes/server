package game

import (
	"context"
	"encoding/binary"
	"server/internal/pb"
	"server/pkg/uid"
	"sync"
	"time"

	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[Game, *Game]("Game")
}

type Option struct {
	MetricInterval           int   `snow:"MetricInterval"`
	MaxOnlinePlayerCount     int   `snow:"MaxOnlinePlayerCount"`
	MaxMessageCountPerSecond int32 `snow:"MaxMessageCountPerSecond"`
	MaxBackstageSecond       int   `snow:"MaxBackstageSecond"`
}

type Game struct {
	node.Service

	opt    *Option
	closed bool

	actors       map[uid.Uid]*actorSession // roleId -> session
	accActors    map[string]*actorSession  // account -> session
	connIdActors map[uint64]*actorSession  // connId -> session

	sAccount  node.IProxy
	sDB       node.IProxy
	sGate     node.IProxy
	sSceneMgr node.IProxy
}

func (ss *Game) Construct(opt *Option) {
	if opt == nil {
		opt = &Option{}
	}
	if opt.MaxOnlinePlayerCount <= 0 {
		opt.MaxOnlinePlayerCount = 3000
	}
	if opt.MaxMessageCountPerSecond <= 0 {
		opt.MaxMessageCountPerSecond = 60
	}
	ss.opt = opt
	ss.actors = make(map[uid.Uid]*actorSession)
	ss.accActors = make(map[string]*actorSession)
	ss.connIdActors = make(map[uint64]*actorSession)
}

func (ss *Game) Start(_ any) {
	ss.Infof("Game service starting")

	ss.sAccount = ss.CreateProxy("Account")
	ss.sDB = ss.CreateProxy("DB")
	ss.sGate = ss.CreateProxy("Gate")
	ss.sSceneMgr = ss.CreateProxy("SceneMgr")

	if ss.opt.MetricInterval > 0 {
		ss.startMetric()
	}

	ss.EnableRpc()
	ss.Infof("Game service started")
}

func (ss *Game) Stop(wg *sync.WaitGroup) {
	ss.Infof("Game service stopping")
	ss.closed = true

	kickWg := &sync.WaitGroup{}
	for _, sess := range ss.actors {
		kickWg.Add(1)
		ss.closeActorSessionWithWg(sess, "server_shutdown", kickWg)
		ss.kickByConnId(sess.connId)
	}

	wg.Add(1)
	go func() {
		kickWg.Wait()
		wg.Done()
	}()
}

func (ss *Game) AfterStop() {
	ss.Infof("Game service stopped")
}

// --------------- RPC: Gate → Game ---------------

func (ss *Game) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{
		"status":  "Game.OK",
		"online":  len(ss.actors),
		"connIds": len(ss.connIdActors),
	})
}

func (ss *Game) RpcOnClientConnect(_ node.IRpcContext, connId uint64, remoteAddr string) {
	if ss.closed {
		return
	}
	if _, exists := ss.connIdActors[connId]; exists {
		ss.Warnf("duplicate connId=%d from %s, ignoring", connId, remoteAddr)
		return
	}
	sess := newActorSession(ss, connId, remoteAddr)
	ss.connIdActors[connId] = sess
	ss.Debugf("new client connected: connId=%d addr=%s", connId, remoteAddr)
}

func (ss *Game) RpcOnClientDisconnect(_ node.IRpcContext, connId uint64) {
	sess, ok := ss.connIdActors[connId]
	if !ok {
		return
	}
	ss.Debugf("client disconnected: %s", sess.descriptor())
	ss.closeActorSession(sess, "client_disconnect")
}

func (ss *Game) RpcOnGateAuthedClient(ctx node.IRpcContext, connId uint64, account string, roleId int64) {
	if ss.closed {
		ctx.Error(nil)
		return
	}

	if existing, ok := ss.accActors[account]; ok && existing.connId != connId {
		if existing.proxy != nil && existing.roleId > 0 {
			ss.Debugf("rebinding actor account=%s roleId=%d to newConn=%d", account, existing.roleId, connId)
			delete(ss.connIdActors, existing.connId)
			existing.connId = connId
			existing.updateDescriptor()
			ss.connIdActors[connId] = existing
			ss.bindRole(int64(existing.roleId), connId)

			existing.proxy.Call("Rebind", connId).
				Catch(func(err error) {
					ss.Errorf("rebind actor failed: %v", err)
				}).Done()

			ctx.Return(int64(existing.roleId))
			return
		}
		ss.removeActor(existing, "duplicate_login")
	}

	sess := newActorSession(ss, connId, "")
	sess.account = account
	sess.updateDescriptor()
	ss.connIdActors[connId] = sess
	ss.accActors[account] = sess

	if roleId > 0 {
		sess.actionSignInRoleById(uid.Uid(roleId))
	} else {
		sess.autoLogin()
	}

	ctx.Return(roleId)
}

func (ss *Game) RpcForwardToActor(ctx node.IRpcContext, connId uint64, msgKey int, msgData []byte) {
	sess, ok := ss.connIdActors[connId]
	if !ok {
		return
	}
	if sess.proxy == nil {
		return
	}
	sess.proxy.Call("ClientRequest", uint16(msgKey), msgData).
		Catch(func(err error) {
			ss.Errorf("forward to actor failed: %v", err)
		}).Done()
	ctx.Return()
}

func (ss *Game) RpcHandleClientMsg(ctx node.IRpcContext, connId uint64, remoteIP string, payload []byte) {
	if len(payload) < pktHeaderLen {
		ctx.Return([]byte(nil))
		return
	}

	sess, ok := ss.connIdActors[connId]
	if !ok {
		if ss.closed {
			ctx.Return([]byte(nil))
			return
		}
		sess = newActorSession(ss, connId, remoteIP)
		ss.connIdActors[connId] = sess
		ss.Debugf("auto-created session for connId=%d addr=%s", connId, remoteIP)
	}

	key := pb.EKey_T(binary.LittleEndian.Uint16(payload[0:2]))
	bodyLen := binary.LittleEndian.Uint32(payload[4:8])
	body := payload[8 : 8+bodyLen]

	msg := pb.Unmarshal(key, body)
	sess.onClientMessage(key, msg, body)
	ctx.Return([]byte(nil))
}

func (ss *Game) RpcActorResponse(_ node.IRpcContext, roleId int64, data []byte) {
	ss.sendToClient(uid.Uid(roleId), data)
}

func (ss *Game) RpcOnRemoteSceneMessage(_ node.IRpcContext, roleId int64, entityId int64, key int, data []byte) {
	sess, ok := ss.actors[uid.Uid(roleId)]
	if !ok || sess.proxy == nil {
		return
	}
	sess.proxy.Call("ClientRequest", uint16(key), data).
		Catch(func(err error) {
			ss.Errorf("remote scene msg forward failed: %v", err)
		}).Done()
}

// --------------- Internal ---------------

const pktHeaderLen = 8

func (ss *Game) sendToClient(roleId uid.Uid, data []byte) {
	sess, ok := ss.actors[roleId]
	if !ok {
		return
	}
	if !ss.sGate.Avail() {
		return
	}
	ss.sGate.Call("SendToClient", sess.connId, data).
		Catch(func(err error) {
			ss.Errorf("send to client failed roleId=%d err=%v", roleId, err)
		}).Done()
}

func (ss *Game) sendToClientByConnId(connId uint64, data []byte) {
	if !ss.sGate.Avail() {
		return
	}
	ss.sGate.Call("SendToClient", connId, data).
		Catch(func(err error) {
			ss.Errorf("send to client failed connId=%d err=%v", connId, err)
		}).Done()
}

func (ss *Game) kickByConnId(connId uint64) {
	if !ss.sGate.Avail() {
		return
	}
	ss.sGate.Call("KickClient", connId).Done()
}

func (ss *Game) bindRole(roleId int64, connId uint64) {
	if !ss.sGate.Avail() {
		return
	}
	ss.sGate.Call("BindRole", roleId, connId).Done()
}

func (ss *Game) closeActorSession(sess *actorSession, reason string) {
	ss.closeActorSessionWithWg(sess, reason, nil)
}

func (ss *Game) closeActorSessionWithWg(sess *actorSession, reason string, wg *sync.WaitGroup) {
	if sess.bClosed {
		if wg != nil {
			wg.Done()
		}
		return
	}
	sess.bClosing = true

	ss.Debugf("closing session %s reason=%s", sess.descriptor(), reason)

	delete(ss.connIdActors, sess.connId)
	if sess.roleId > 0 {
		delete(ss.actors, sess.roleId)
	}
	if sess.account != "" {
		delete(ss.accActors, sess.account)
	}

	if sess.proxy != nil && sess.sAddr != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		sAddr := sess.sAddr

		sess.proxy.Call("Kick").
			WithContext(ctx).
			Catch(func(err error) {
				ss.Debugf("kick actor failed roleId=%d: %v", sess.roleId, err)
			}).
			Final(func() {
				cancel()
				node.StopService(sAddr)
				if wg != nil {
					wg.Done()
				}
			}).Done()
	} else {
		if wg != nil {
			wg.Done()
		}
	}

	sess.bClosed = true
}

func (ss *Game) removeActor(sess *actorSession, reason string) {
	ss.closeActorSession(sess, reason)
	ss.kickByConnId(sess.connId)
}
