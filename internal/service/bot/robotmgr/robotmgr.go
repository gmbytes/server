package robotmgr

import (
	"sync"

	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[RobotMgr, *RobotMgr]("RobotMgr")
}

type Option struct {
	GateAddr   string `snow:"GateAddr"`
	RobotCount int    `snow:"RobotCount"`
}

type robotSession struct {
	index int
	sAddr int32
	proxy node.IProxy
}

type RobotMgr struct {
	node.Service

	opt     *Option
	robots  map[int]*robotSession
	nextIdx int
	closed  bool
}

func (ss *RobotMgr) Construct(opt *Option) {
	if opt == nil {
		opt = &Option{}
	}
	if opt.GateAddr == "" {
		opt.GateAddr = "127.0.0.1:61101"
	}
	ss.opt = opt
	ss.robots = make(map[int]*robotSession)
}

func (ss *RobotMgr) Start(_ any) {
	ss.Infof("RobotMgr starting, gate=%s robotCount=%d", ss.opt.GateAddr, ss.opt.RobotCount)
	ss.EnableRpc()

	if ss.opt.RobotCount > 0 {
		ss.startRobots(ss.opt.RobotCount)
	}
}

func (ss *RobotMgr) Stop(_ *sync.WaitGroup) {
	ss.Infof("RobotMgr stopping, active robots=%d", len(ss.robots))
	ss.closed = true
	ss.stopAllRobots()
}

func (ss *RobotMgr) AfterStop() {
	ss.Infof("RobotMgr stopped")
}

// --------------- RPC ---------------

func (ss *RobotMgr) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{
		"status":   "RobotMgr.OK",
		"robots":   len(ss.robots),
		"gateAddr": ss.opt.GateAddr,
	})
}

func (ss *RobotMgr) RpcStartRobots(ctx node.IRpcContext, count int) {
	if ss.closed {
		ctx.Return(0)
		return
	}
	started := ss.startRobots(count)
	ss.Infof("RobotMgr started %d/%d robots", started, count)
	ctx.Return(started)
}

func (ss *RobotMgr) RpcStopRobots(ctx node.IRpcContext, count int) {
	stopped := ss.stopRobots(count)
	ss.Infof("RobotMgr stopped %d/%d robots", stopped, count)
	ctx.Return(stopped)
}

func (ss *RobotMgr) RpcStopAll(ctx node.IRpcContext) {
	ss.stopAllRobots()
	ss.Infof("RobotMgr stopped all robots")
	ctx.Return()
}

// --------------- Internal ---------------

func (ss *RobotMgr) startRobots(count int) int {
	started := 0
	for i := 0; i < count; i++ {
		if ss.startOneRobot() {
			started++
		}
	}
	return started
}

func (ss *RobotMgr) startOneRobot() bool {
	ss.nextIdx++
	idx := ss.nextIdx

	sAddr, err := node.NewService("Robot")
	if err != nil {
		ss.Errorf("create Robot service failed: %v", err)
		return false
	}

	proxy := ss.CreateProxyByNodeAddr(node.AddrLocal, sAddr)
	sess := &robotSession{
		index: idx,
		sAddr: sAddr,
		proxy: proxy,
	}
	ss.robots[idx] = sess

	node.StartService(sAddr, nil)

	proxy.Call("Init", idx, ss.opt.GateAddr).
		Catch(func(err error) {
			ss.Fork("robot.initFail", func() {
				ss.Errorf("Robot[%d] init failed: %v", idx, err)
				delete(ss.robots, idx)
				node.StopService(sAddr)
			})
		}).Done()

	return true
}

func (ss *RobotMgr) stopRobots(count int) int {
	stopped := 0
	for idx, sess := range ss.robots {
		if stopped >= count {
			break
		}
		ss.closeRobotSession(sess)
		delete(ss.robots, idx)
		stopped++
	}
	return stopped
}

func (ss *RobotMgr) stopAllRobots() {
	for idx, sess := range ss.robots {
		ss.closeRobotSession(sess)
		delete(ss.robots, idx)
	}
}

func (ss *RobotMgr) closeRobotSession(sess *robotSession) {
	if sess.proxy == nil || sess.sAddr == 0 {
		return
	}
	sess.proxy.Call("Kick").
		Catch(func(err error) {
			ss.Debugf("kick Robot[%d] failed: %v", sess.index, err)
		}).
		Final(func() {
			node.StopService(sess.sAddr)
		}).Done()
}
