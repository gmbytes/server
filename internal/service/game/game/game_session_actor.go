package game

import (
	"fmt"
	"server/internal/pb"
	"server/pkg/uid"
	"time"

	"github.com/gmbytes/snow/routines/node"
)

type sessionState int32

const (
	sessionStateInit             sessionState = iota // 等待 Login 请求
	sessionStateAfterGetRoleList                     // 可创建/删除/登录角色
	sessionStateBeforeRoleLogin                      // 角色正在登录，禁止顶号
	sessionStateAfterRoleLogin                       // 转发消息给 Actor
)

type actorSession struct {
	gs *Game

	connId    uint64
	describer string
	ipAddr    string

	account string
	roleId  uid.Uid
	sAddr   int32
	proxy   node.IProxy

	state   sessionState
	spArgs  *pb.ReqLogin
	roles   []int64
	loginTs int64

	pkgUpdateTs int64
	pkgCount    int32

	bClosing bool
	bClosed  bool

	metricMs int64
}

func newActorSession(gs *Game, connId uint64, ipAddr string) *actorSession {
	now := gs.GetSecond()
	as := &actorSession{
		gs:          gs,
		connId:      connId,
		state:       sessionStateInit,
		ipAddr:      ipAddr,
		loginTs:     now,
		pkgUpdateTs: now,
	}
	as.updateDescriptor()
	return as
}

func (ss *actorSession) descriptor() string {
	return ss.describer
}

func (ss *actorSession) updateDescriptor() {
	switch {
	case ss.account != "" && ss.roleId > 0:
		ss.describer = fmt.Sprintf("sess[%d|%s|%d|%s]", ss.connId, ss.account, ss.roleId, ss.ipAddr)
	case ss.account != "":
		ss.describer = fmt.Sprintf("sess[%d|%s|?|%s]", ss.connId, ss.account, ss.ipAddr)
	default:
		ss.describer = fmt.Sprintf("sess[%d|?|?|%s]", ss.connId, ss.ipAddr)
	}
}

func (ss *actorSession) avail() bool {
	return !ss.bClosing && !ss.bClosed
}

// onClientPackage 根据当前 session 状态分发客户端消息
func (ss *actorSession) onClientPackage(key pb.EKey_T, serialNumber uint32, content []byte) {
	if !ss.avail() {
		return
	}

	now := ss.gs.GetSecond()
	if now != ss.pkgUpdateTs {
		ss.pkgUpdateTs = now
		ss.pkgCount = 0
	}
	ss.pkgCount++
	if ss.gs.opt.MaxMessageCountPerSecond > 0 && ss.pkgCount > ss.gs.opt.MaxMessageCountPerSecond {
		ss.gs.Warnf("%s msg rate exceeded", ss.descriptor())
		ss.gs.removeActor(ss, "msg_rate_exceeded")
		return
	}

	if key == pb.EKey_Ping {
		ss.handlePing(serialNumber)
		return
	}

	switch ss.state {
	case sessionStateInit:
		ss.onStateInit(key, serialNumber, content)
	case sessionStateAfterGetRoleList:
		ss.onStateRoleList(key, serialNumber, content)
	case sessionStateBeforeRoleLogin:
		ss.gs.Debugf("%s msg during role login, ignoring key=%d", ss.descriptor(), key)
	case sessionStateAfterRoleLogin:
		ss.onStateForward(key, serialNumber, content)
	}
}

// ---- State: Init → 等待 Login 请求 ----

func (ss *actorSession) onStateInit(key pb.EKey_T, serialNumber uint32, content []byte) {
	if key != pb.EKey_Login {
		ss.gs.Warnf("%s unexpected key=%d in init state", ss.descriptor(), key)
		return
	}
	ss.addMetric("SignInPlayer")
	ss.signInPlayer(serialNumber, content)
}

// signInPlayer 处理登录请求：验证 → 从 DB 获取角色列表 → 切换到 RoleList 状态
func (ss *actorSession) signInPlayer(serialNumber uint32, _ []byte) {
	account := ""
	if ss.spArgs != nil {
		account = ss.spArgs.Account
	}

	if account == "" {
		account = fmt.Sprintf("guest_%d", ss.connId)
	}

	if existing, ok := ss.gs.accActors[account]; ok && existing != ss {
		ss.gs.Debugf("duplicate login account=%s, kicking old session", account)
		ss.gs.removeActor(existing, "duplicate_login")
	}

	ss.account = account
	ss.gs.accActors[account] = ss
	ss.updateDescriptor()

	ss.gs.Debugf("%s signInPlayer", ss.descriptor())

	if !ss.gs.sDB.Avail() {
		ss.gs.Warnf("%s DB unavailable for signInPlayer", ss.descriptor())
		ss.sendErrorPkg(serialNumber, pb.EKey_Login, 1)
		return
	}

	ss.gs.sDB.Call("GetRoles", account).
		Then(func(roles []int64) {
			ss.gs.Fork("signInPlayer.gotRoles", func() {
				if !ss.avail() {
					return
				}
				ss.roles = roles
				ss.state = sessionStateAfterGetRoleList

				rsp := &pb.Package{
					KeyCode:      pb.EKey_Login,
					SerialNumber: serialNumber,
				}
				ss.sendPackage(rsp)
				ss.gs.Debugf("%s login success, roles=%v", ss.descriptor(), roles)
			})
		}).
		Catch(func(err error) {
			ss.gs.Errorf("%s RpcGetRoles failed: %v", ss.descriptor(), err)
			ss.sendErrorPkg(serialNumber, pb.EKey_Login, 1)
		}).Done()
}

// ---- State: AfterGetRoleList → 创建/登录角色 ----

func (ss *actorSession) onStateRoleList(key pb.EKey_T, serialNumber uint32, content []byte) {
	switch key {
	case pb.EKey_CreateRole:
		ss.addMetric("NewRole")
		ss.actionCreateRole(serialNumber, content)
	case pb.EKey_LoginRole:
		ss.addMetric("SignInRole")
		ss.actionSignInRole(serialNumber, content)
	default:
		ss.gs.Warnf("%s unexpected key=%d in role_list state", ss.descriptor(), key)
	}
}

func (ss *actorSession) actionCreateRole(serialNumber uint32, _ []byte) {
	if !ss.gs.sDB.Avail() {
		ss.sendErrorPkg(serialNumber, pb.EKey_CreateRole, 1)
		return
	}

	ss.gs.sDB.Call("InsertRoleData", ss.account).
		Then(func(roleId int64) {
			ss.gs.Fork("createRole.done", func() {
				if !ss.avail() {
					return
				}
				ss.roles = append(ss.roles, roleId)
				rsp := &pb.Package{
					KeyCode:      pb.EKey_CreateRole,
					SerialNumber: serialNumber,
				}
				ss.sendPackage(rsp)
				ss.gs.Debugf("%s created role %d", ss.descriptor(), roleId)
			})
		}).
		Catch(func(err error) {
			ss.gs.Errorf("%s RpcInsertRoleData failed: %v", ss.descriptor(), err)
			ss.sendErrorPkg(serialNumber, pb.EKey_CreateRole, 1)
		}).Done()
}

func (ss *actorSession) actionSignInRole(serialNumber uint32, _ []byte) {
	if len(ss.roles) == 0 {
		ss.sendErrorPkg(serialNumber, pb.EKey_LoginRole, 1)
		return
	}

	roleId := uid.Uid(ss.roles[0])
	ss.state = sessionStateBeforeRoleLogin
	ss.gs.Debugf("%s signing in role %d", ss.descriptor(), roleId)

	if !ss.gs.sDB.Avail() {
		ss.state = sessionStateAfterGetRoleList
		ss.sendErrorPkg(serialNumber, pb.EKey_LoginRole, 1)
		return
	}

	ss.gs.sDB.Call("GetRoleData", int64(roleId)).
		Then(func(roleData *pb.RoleSummaryData) {
			ss.gs.Fork("signInRole.gotData", func() {
				if !ss.avail() {
					return
				}
				ss.createActor(roleId, roleData, serialNumber)
			})
		}).
		Catch(func(err error) {
			ss.gs.Fork("signInRole.err", func() {
				ss.gs.Errorf("%s RpcGetRoleData failed: %v", ss.descriptor(), err)
				ss.state = sessionStateAfterGetRoleList
				ss.sendErrorPkg(serialNumber, pb.EKey_LoginRole, 1)
			})
		}).Done()
}

// createActor 创建 Actor 服务实例并绑定到 session
func (ss *actorSession) createActor(roleId uid.Uid, roleData *pb.RoleSummaryData, serialNumber uint32) {
	sAddr, err := node.NewService("Actor")
	if err != nil {
		ss.gs.Errorf("%s create Actor failed: %v", ss.descriptor(), err)
		ss.state = sessionStateAfterGetRoleList
		ss.sendErrorPkg(serialNumber, pb.EKey_LoginRole, 1)
		return
	}

	ss.roleId = roleId
	ss.sAddr = sAddr
	ss.proxy = ss.gs.CreateProxyByNodeAddr(node.AddrLocal, sAddr)
	ss.updateDescriptor()

	ss.gs.actors[roleId] = ss
	ss.gs.bindRole(int64(roleId), ss.connId)

	node.StartService(sAddr, nil)

	ss.proxy.Call("Init", int64(roleId), ss.connId, roleData).
		Then(func() {
			ss.gs.Fork("actor.inited", func() {
				if !ss.avail() {
					return
				}
				ss.state = sessionStateAfterRoleLogin
				rsp := &pb.Package{
					KeyCode:      pb.EKey_LoginRole,
					SerialNumber: serialNumber,
				}
				ss.sendPackage(rsp)
				ss.gs.Debugf("%s actor created and ready", ss.descriptor())
			})
		}).
		Catch(func(err error) {
			ss.gs.Fork("actor.initFail", func() {
				ss.gs.Errorf("%s actor init failed: %v", ss.descriptor(), err)
				ss.state = sessionStateAfterGetRoleList
				node.StopService(sAddr)
				ss.proxy = nil
				ss.sAddr = 0
				ss.sendErrorPkg(serialNumber, pb.EKey_LoginRole, 1)
			})
		}).Done()
}

// ---- State: AfterRoleLogin → 转发给 Actor ----

func (ss *actorSession) onStateForward(key pb.EKey_T, serialNumber uint32, content []byte) {
	if ss.proxy == nil {
		return
	}
	ss.proxy.Call("ClientRequest", uint16(key), serialNumber, content).
		Catch(func(err error) {
			ss.gs.Errorf("%s forward to actor failed key=%d: %v", ss.descriptor(), key, err)
		}).Done()
}

// ---- Helpers ----

func (ss *actorSession) handlePing(serialNumber uint32) {
	rsp := &pb.Package{
		KeyCode:      pb.EKey_Ping,
		SerialNumber: serialNumber,
	}
	ss.sendPackage(rsp)
}

func (ss *actorSession) sendPackage(p *pb.Package) {
	if ss.bClosed {
		return
	}
	ss.gs.dispatchPkg(ss, p)
}

func (ss *actorSession) sendErrorPkg(serialNumber uint32, key pb.EKey_T, errCode uint16) {
	errBytes := make([]byte, 2)
	errBytes[0] = byte(errCode)
	errBytes[1] = byte(errCode >> 8)
	p := &pb.Package{
		KeyCode:      key,
		SerialNumber: serialNumber,
		Content:      errBytes,
	}
	ss.sendPackage(p)
}

func (ss *actorSession) addMetric(name string) {
	if ss.gs.opt.MetricInterval <= 0 {
		return
	}
	nowMs := time.Now().UnixMilli()
	addMetric(name, nowMs-ss.metricMs)
	ss.metricMs = nowMs
}

func (ss *actorSession) resetMetricTime() {
	if ss.gs.opt.MetricInterval <= 0 {
		return
	}
	ss.metricMs = time.Now().UnixMilli()
}
