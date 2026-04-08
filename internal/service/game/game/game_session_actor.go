package game

import (
	"fmt"
	"server/internal/pb"
	"server/pkg/uid"
	"time"

	"github.com/gmbytes/snow/pkg/routines/node"
	"google.golang.org/protobuf/proto"
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
	roles   []*pb.RoleSummaryData
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

// onClientMessage 根据当前 session 状态分发客户端消息
func (ss *actorSession) onClientMessage(key pb.EKey_T, msg any, body []byte) {
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

	if key == pb.EKey_ReqPing {
		if ss.state == sessionStateAfterRoleLogin {
			ss.onStateForward(key, body)
		} else {
			ss.handlePing()
		}
		return
	}

	switch ss.state {
	case sessionStateInit:
		ss.onStateInit(key, msg)
	case sessionStateAfterGetRoleList:
		ss.onStateRoleList(key, msg)
	case sessionStateBeforeRoleLogin:
		ss.gs.Debugf("%s msg during role login, ignoring key=%d", ss.descriptor(), key)
	case sessionStateAfterRoleLogin:
		ss.onStateForward(key, body)
	}
}

// ---- State: Init → 等待 Login 请求 ----

func (ss *actorSession) onStateInit(key pb.EKey_T, msg any) {
	if key != pb.EKey_ReqLogin {
		ss.gs.Warnf("%s unexpected key=%d in init state", ss.descriptor(), key)
		return
	}
	ss.addMetric("SignInPlayer")
	ss.signInPlayer(msg)
}

// signInPlayer 处理登录请求：验证 → 通过 Account 获取角色列表 → 切换到 RoleList 状态
func (ss *actorSession) signInPlayer(msg any) {
	if req, ok := msg.(*pb.ReqLogin); ok {
		ss.spArgs = req
	}
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

	if !ss.gs.sAccount.Avail() {
		ss.gs.Warnf("%s Account unavailable for signInPlayer", ss.descriptor())
		ss.sendErrorPkg(&pb.RspLogin{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
		return
	}

	ss.gs.sAccount.Call("GetRoles", account).
		Then(func(roles []*pb.RoleSummaryData) {
			ss.gs.Fork("signInPlayer.gotRoles", func() {
				if !ss.avail() {
					return
				}
				ss.roles = roles
				ss.state = sessionStateAfterGetRoleList

				ss.sendPackage(pb.NewPackage(&pb.RspLogin{
					Err:        pb.EErrorCode_Ok,
					Roles:      roles,
					Account:    ss.account,
					ServerTime: time.Now().UnixMilli(),
				}))
				ss.gs.Debugf("%s login success, roleCount=%d", ss.descriptor(), len(roles))
			})
		}).
		Catch(func(err error) {
			ss.gs.Errorf("%s Account.GetRoles failed: %v", ss.descriptor(), err)
			ss.sendErrorPkg(&pb.RspLogin{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
		}).Done()
}

// ---- State: AfterGetRoleList → 创建/登录角色 ----

func (ss *actorSession) onStateRoleList(key pb.EKey_T, msg any) {
	switch key {
	case pb.EKey_ReqCreateRole:
		ss.addMetric("NewRole")
		ss.actionCreateRole(msg)
	case pb.EKey_ReqLoginRole:
		ss.addMetric("SignInRole")
		ss.actionSignInRole(msg)
	default:
		ss.gs.Warnf("%s unexpected key=%d in role_list state", ss.descriptor(), key)
	}
}

func (ss *actorSession) actionCreateRole(msg any) {
	if !ss.gs.sAccount.Avail() {
		ss.sendErrorPkg(&pb.RspCreateRole{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
		return
	}
	req, ok := msg.(*pb.ReqCreateRole)
	if !ok {
		ss.sendErrorPkg(&pb.RspCreateRole{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
		return
	}

	ss.gs.sAccount.Call("CreateRole", ss.account, req.GetCid(), req.GetName()).
		Then(func(role *pb.RoleSummaryData) {
			ss.gs.Fork("createRole.done", func() {
				if !ss.avail() || role == nil {
					return
				}
				ss.roles = append(ss.roles, role)
				ss.sendPackage(pb.NewPackage(&pb.RspCreateRole{
					Err:  pb.EErrorCode_Ok,
					Role: role,
				}))
				ss.gs.Debugf("%s created role %d", ss.descriptor(), role.GetId())
			})
		}).
		Catch(func(err error) {
			ss.gs.Errorf("%s Account.CreateRole failed: %v", ss.descriptor(), err)
			ss.sendErrorPkg(&pb.RspCreateRole{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
		}).Done()
}

func (ss *actorSession) actionSignInRole(msg any) {
	if len(ss.roles) == 0 {
		ss.sendErrorPkg(&pb.RspLoginRole{Err: pb.EErrorCode_RoleNotFound}, pb.EErrorCode_RoleNotFound)
		return
	}

	roleId := uid.Uid(ss.roles[0].GetId())
	if req, ok := msg.(*pb.ReqLoginRole); ok && req.GetRoleId() != 0 {
		roleId = uid.Uid(req.GetRoleId())
	}
	if ss.findRole(roleId) == nil {
		ss.sendErrorPkg(&pb.RspLoginRole{Err: pb.EErrorCode_RoleNotFound}, pb.EErrorCode_RoleNotFound)
		return
	}
	ss.state = sessionStateBeforeRoleLogin
	ss.gs.Debugf("%s signing in role %d", ss.descriptor(), roleId)

	if !ss.gs.sAccount.Avail() {
		ss.state = sessionStateAfterGetRoleList
		ss.sendErrorPkg(&pb.RspLoginRole{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
		return
	}

	ss.gs.sAccount.Call("GetRole", ss.account, int64(roleId)).
		Then(func(roleData *pb.RoleSummaryData) {
			ss.gs.Fork("signInRole.gotData", func() {
				if !ss.avail() {
					return
				}
				ss.createActor(roleId, roleData)
			})
		}).
		Catch(func(err error) {
			ss.gs.Fork("signInRole.err", func() {
				ss.gs.Errorf("%s Account.GetRole failed: %v", ss.descriptor(), err)
				ss.state = sessionStateAfterGetRoleList
				ss.sendErrorPkg(&pb.RspLoginRole{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
			})
		}).Done()
}

// createActor 创建 Actor 服务实例并绑定到 session
func (ss *actorSession) createActor(roleId uid.Uid, roleData *pb.RoleSummaryData) {
	sAddr, err := node.NewService("Actor")
	if err != nil {
		ss.gs.Errorf("%s create Actor failed: %v", ss.descriptor(), err)
		ss.state = sessionStateAfterGetRoleList
		ss.sendErrorPkg(&pb.RspLoginRole{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
		return
	}

	ss.roleId = roleId
	ss.sAddr = sAddr
	ss.proxy = ss.gs.CreateProxyByNodeAddr(node.AddrLocal, sAddr)
	ss.updateDescriptor()

	ss.gs.actors[roleId] = ss
	ss.gs.bindRole(int64(roleId), ss.connId)

	node.StartService(sAddr, nil)

	ss.proxy.Call("Init", int64(roleId), ss.connId, ss.account, roleData).
		Then(func() {
			ss.gs.Fork("actor.inited", func() {
				if !ss.avail() {
					return
				}
				ss.state = sessionStateAfterRoleLogin
				ss.sendPackage(pb.NewPackage(&pb.RspLoginRole{Err: pb.EErrorCode_Ok}))
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
				ss.sendErrorPkg(&pb.RspLoginRole{Err: pb.EErrorCode_Failed}, pb.EErrorCode_Failed)
			})
		}).Done()
}

// ---- State: AfterRoleLogin → 转发给 Actor ----

func (ss *actorSession) onStateForward(key pb.EKey_T, body []byte) {
	if ss.proxy == nil {
		return
	}
	ss.proxy.Call("ClientRequest", uint16(key), body).
		Catch(func(err error) {
			ss.gs.Errorf("%s forward to actor failed key=%d: %v", ss.descriptor(), key, err)
		}).Done()
}

// ---- Helpers ----

func (ss *actorSession) handlePing() {
	ss.sendPackage(pb.NewPackage(&pb.RspPing{}))
}

func (ss *actorSession) sendPackage(p *pb.Package) {
	if ss.bClosed {
		return
	}
	ss.gs.dispatchPkg(ss, p)
}

func (ss *actorSession) sendErrorPkg(rsp proto.Message, errCode pb.EErrorCode_T) {
	ss.sendPackage(pb.NewPackage(rsp, errCode))
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

func (ss *actorSession) findRole(roleId uid.Uid) *pb.RoleSummaryData {
	for _, role := range ss.roles {
		if role != nil && role.GetId() == int64(roleId) {
			return role
		}
	}
	return nil
}

func (ss *actorSession) actionSignInRoleById(roleId uid.Uid) {
	ss.state = sessionStateBeforeRoleLogin
	ss.gs.Debugf("%s signing in role %d (from GateAuthed)", ss.descriptor(), roleId)

	if !ss.gs.sAccount.Avail() {
		ss.state = sessionStateAfterGetRoleList
		return
	}

	ss.gs.sAccount.Call("GetRole", ss.account, int64(roleId)).
		Then(func(roleData *pb.RoleSummaryData) {
			ss.gs.Fork("gateAuth.signInRole.gotData", func() {
				if !ss.avail() {
					return
				}
				ss.createActor(roleId, roleData)
			})
		}).
		Catch(func(err error) {
			ss.gs.Fork("gateAuth.signInRole.err", func() {
				ss.gs.Errorf("%s Account.GetRole failed (gateAuth): %v", ss.descriptor(), err)
				ss.state = sessionStateInit
			})
		}).Done()
}

func (ss *actorSession) autoLogin() {
	if !ss.gs.sAccount.Avail() {
		return
	}

	ss.gs.sAccount.Call("GetRoles", ss.account).
		Then(func(roles []*pb.RoleSummaryData) {
			ss.gs.Fork("autoLogin.gotRoles", func() {
				if !ss.avail() {
					return
				}
				ss.roles = roles
				ss.state = sessionStateAfterGetRoleList
				ss.gs.Debugf("%s autoLogin got %d roles", ss.descriptor(), len(roles))
			})
		}).
		Catch(func(err error) {
			ss.gs.Errorf("%s autoLogin GetRoles failed: %v", ss.descriptor(), err)
		}).Done()
}
