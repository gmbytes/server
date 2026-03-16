package robot

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"server/internal/pb"
	"server/pkg/net_pkg"

	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[Robot, *Robot]("Robot")
}

type robotState int32

const (
	stateInit            robotState = iota
	stateConnecting                        // TCP 连接中
	stateWaitLogin                         // 等待 Login 响应
	stateWaitCreateRole                    // 等待 CreateRole 响应
	stateWaitLoginRole                     // 等待 LoginRole 响应
	stateReady                             // 完成登录，可压测
	stateStopped                           // 已停止
)

const pktHeaderLen = 6

// Robot 代表一个模拟客户端，通过 TCP 连接到 Gate 进行压力测试。
type Robot struct {
	node.Service

	index    int
	gateAddr string
	account  string

	conn      net.Conn
	state     robotState
	serialNum uint32
	closed    atomic.Bool

	pingHandle   ITimeWheelHandle
	stressHandle ITimeWheelHandle

	sRobotMgr node.IProxy

	sendCount  int64
	recvCount  int64
	errorCount int64
}

type ITimeWheelHandle = node.ITimeWheelHandle

func (r *Robot) Start(_ any) {
	r.sRobotMgr = r.CreateProxy("RobotMgr")
	r.EnableRpc()
}

func (r *Robot) Stop(_ *sync.WaitGroup) {
	r.Infof("Robot[%d] stopping", r.index)
	r.disconnect()
}

func (r *Robot) AfterStop() {
	r.Infof("Robot[%d] stopped, sent=%d recv=%d errors=%d",
		r.index, r.sendCount, r.recvCount, r.errorCount)
}

// --------------- RPC: RobotMgr → Robot ---------------

func (r *Robot) RpcInit(ctx node.IRpcContext, index int, gateAddr string) {
	r.index = index
	r.gateAddr = gateAddr
	r.account = fmt.Sprintf("robot_%d", index)

	r.Infof("Robot[%d] init, gate=%s account=%s", r.index, r.gateAddr, r.account)
	r.connect()
	ctx.Return()
}

func (r *Robot) RpcKick(ctx node.IRpcContext) {
	r.Infof("Robot[%d] kicked", r.index)
	r.disconnect()
	ctx.Return()
}

func (r *Robot) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{
		"index":   r.index,
		"state":   int(r.state),
		"sent":    r.sendCount,
		"recv":    r.recvCount,
		"errors":  r.errorCount,
		"account": r.account,
	})
}

// --------------- TCP Connection ---------------

func (r *Robot) connect() {
	r.state = stateConnecting

	go func() {
		conn, err := net.DialTimeout("tcp", r.gateAddr, 10*time.Second)
		if err != nil {
			r.Fork("connect.fail", func() {
				r.Errorf("Robot[%d] connect to %s failed: %v", r.index, r.gateAddr, err)
				r.errorCount++
				r.state = stateStopped
			})
			return
		}

		if err := net_pkg.ClientPkgPreprocessor.Process(conn); err != nil {
			r.Fork("handshake.fail", func() {
				r.Errorf("Robot[%d] handshake failed: %v", r.index, err)
				r.errorCount++
				r.state = stateStopped
			})
			_ = conn.Close()
			return
		}

		r.Fork("connected", func() {
			if r.closed.Load() {
				_ = conn.Close()
				return
			}
			r.conn = conn
			r.Infof("Robot[%d] connected to %s", r.index, r.gateAddr)
			go r.readLoop()
			r.sendLogin()
		})
	}()
}

func (r *Robot) disconnect() {
	if r.closed.Swap(true) {
		return
	}
	r.state = stateStopped
	if r.pingHandle != nil {
		r.pingHandle.Stop()
	}
	if r.stressHandle != nil {
		r.stressHandle.Stop()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

// --------------- Read Loop ---------------

func (r *Robot) readLoop() {
	defer func() {
		r.Fork("readLoop.done", func() {
			if !r.closed.Load() {
				r.Infof("Robot[%d] read loop ended, disconnecting", r.index)
				r.disconnect()
			}
		})
	}()

	header := make([]byte, pktHeaderLen)
	for {
		if r.closed.Load() {
			return
		}

		if _, err := io.ReadFull(r.conn, header); err != nil {
			if !r.closed.Load() {
				r.Fork("read.err", func() {
					r.Debugf("Robot[%d] read header error: %v", r.index, err)
					r.errorCount++
				})
			}
			return
		}

		key := pb.EKey_T(binary.LittleEndian.Uint16(header[0:2]))
		bodyLen := binary.LittleEndian.Uint32(header[2:6])

		var body []byte
		if bodyLen > 0 {
			body = make([]byte, bodyLen)
			if _, err := io.ReadFull(r.conn, body); err != nil {
				if !r.closed.Load() {
					r.Fork("read.body.err", func() {
						r.Debugf("Robot[%d] read body error: %v", r.index, err)
						r.errorCount++
					})
				}
				return
			}
		}

		var sn uint32
		if len(body) >= 4 {
			sn = binary.LittleEndian.Uint32(body[0:4])
		}

		r.Fork("pkt", func() {
			r.recvCount++
			r.onPacket(key, sn)
		})
	}
}

// --------------- Packet Handler ---------------

func (r *Robot) onPacket(key pb.EKey_T, sn uint32) {
	r.Debugf("Robot[%d] recv key=%d sn=%d state=%d", r.index, key, sn, r.state)

	switch key {
	case pb.EKey_Ping:
		return

	case pb.EKey_Login:
		if r.state == stateWaitLogin {
			r.Infof("Robot[%d] login ok", r.index)
			r.sendCreateRole()
		}

	case pb.EKey_CreateRole:
		if r.state == stateWaitCreateRole {
			r.Infof("Robot[%d] create role ok", r.index)
			r.sendLoginRole()
		}

	case pb.EKey_LoginRole:
		if r.state == stateWaitLoginRole {
			r.Infof("Robot[%d] login role ok → ready", r.index)
			r.state = stateReady
			r.startStressTest()
		}

	case pb.EKey_EnterZone:
		r.Debugf("Robot[%d] enter zone response", r.index)
	}
}

// --------------- Packet Writer ---------------

func (r *Robot) nextSN() uint32 {
	r.serialNum++
	return r.serialNum
}

func (r *Robot) sendPacket(key pb.EKey_T, content []byte) {
	if r.conn == nil || r.closed.Load() {
		return
	}

	p := &pb.Package{
		KeyCode:      key,
		SerialNumber: r.nextSN(),
		Content:      content,
	}

	data, err := p.Bytes()
	if err != nil {
		r.Errorf("Robot[%d] marshal failed: %v", r.index, err)
		return
	}

	if _, err := r.conn.Write(data); err != nil {
		r.Errorf("Robot[%d] write failed: %v", r.index, err)
		r.errorCount++
		return
	}
	r.sendCount++
}

func (r *Robot) sendLogin() {
	r.state = stateWaitLogin
	r.sendPacket(pb.EKey_Login, nil)
	r.Debugf("Robot[%d] → Login", r.index)
}

func (r *Robot) sendCreateRole() {
	r.state = stateWaitCreateRole
	r.sendPacket(pb.EKey_CreateRole, nil)
	r.Debugf("Robot[%d] → CreateRole", r.index)
}

func (r *Robot) sendLoginRole() {
	r.state = stateWaitLoginRole
	r.sendPacket(pb.EKey_LoginRole, nil)
	r.Debugf("Robot[%d] → LoginRole", r.index)
}

// --------------- Stress Test ---------------

func (r *Robot) startStressTest() {
	r.pingHandle = r.Tick(5*time.Second, 0, func() {
		if r.state == stateReady {
			r.sendPacket(pb.EKey_Ping, nil)
		}
	})

	r.stressHandle = r.Tick(2*time.Second, time.Second, func() {
		if r.state == stateReady {
			r.sendPacket(pb.EKey_EnterZone, nil)
		}
	})

	r.sendPacket(pb.EKey_EnterZone, nil)
}
