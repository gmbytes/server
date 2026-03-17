package game

import (
	"encoding/binary"
	"server/internal/pb"
	"server/pkg/uid"
	"sync"

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

// Game 管理所有 Actor（玩家角色）的服务，通过 RPC 与无状态 Gate 通信。
type Game struct {
	node.Service

	opt    *Option
	closed bool

	actors       map[uid.Uid]*actorSession // roleId -> session
	accActors    map[string]*actorSession  // account -> session
	connIdActors map[uint64]*actorSession  // connId -> session

	sDB   node.IProxy
	sGate node.IProxy
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

	ss.sDB = ss.CreateProxy("DB")
	ss.sGate = ss.CreateProxy("Gate")

	if ss.opt.MetricInterval > 0 {
		ss.startMetric()
	}

	ss.EnableRpc()
	ss.Infof("Game service started")
}

func (ss *Game) Stop(wg *sync.WaitGroup) {
	ss.Infof("Game service stopping")
	ss.closed = true

	for _, sess := range ss.actors {
		ss.removeActor(sess, "server_shutdown")
	}

	_ = wg
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

// RpcOnClientConnect Gate 通知：新客户端连接
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

// RpcOnClientDisconnect Gate 通知：客户端断开
func (ss *Game) RpcOnClientDisconnect(_ node.IRpcContext, connId uint64) {
	sess, ok := ss.connIdActors[connId]
	if !ok {
		return
	}
	ss.Debugf("client disconnected: %s", sess.descriptor())
	ss.closeActorSession(sess, "client_disconnect")
}

// RpcHandleClientMsg Gate 转发的客户端消息（一个完整的协议包）
func (ss *Game) RpcHandleClientMsg(ctx node.IRpcContext, connId uint64, _ string, payload []byte) {
	if len(payload) < pktHeaderLen {
		ctx.Return([]byte(nil))
		return
	}

	sess, ok := ss.connIdActors[connId]
	if !ok {
		ss.Debugf("msg for unknown connId=%d, telling gate to close", connId)
		ss.kickByConnId(connId)
		ctx.Return([]byte(nil))
		return
	}

	key := pb.EKey_T(binary.LittleEndian.Uint16(payload[0:2]))
	// errCode := pb.EErrorCode_T(binary.LittleEndian.Uint16(payload[2:4])) // 客户端请求一般为 0
	bodyLen := binary.LittleEndian.Uint32(payload[4:8])
	body := payload[8 : 8+bodyLen]

	msg := pb.Unmarshal(key, body)
	sess.onClientMessage(key, msg, body)
	ctx.Return([]byte(nil))
}

// RpcActorResponse Actor 通过 Game 向客户端发送数据
func (ss *Game) RpcActorResponse(_ node.IRpcContext, roleId int64, data []byte) {
	ss.sendToClient(uid.Uid(roleId), data)
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
	if sess.bClosed {
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
		sess.proxy.Call("Kick").
			Catch(func(err error) {
				ss.Debugf("kick actor failed roleId=%d: %v", sess.roleId, err)
			}).
			Final(func() {
				node.StopService(sess.sAddr)
			}).Done()
	}

	sess.bClosed = true
}

func (ss *Game) removeActor(sess *actorSession, reason string) {
	ss.closeActorSession(sess, reason)
	ss.kickByConnId(sess.connId)
}
