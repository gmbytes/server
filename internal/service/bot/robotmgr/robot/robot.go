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

	"github.com/gmbytes/snow/pkg/routines/node"
	"google.golang.org/protobuf/proto"
)

func init() {
	node.Register[Robot, *Robot]("Robot")
}

type robotState int32

const (
	stateInit           robotState = iota
	stateConnecting                // TCP 连接中
	stateWaitLogin                 // 等待 Login 响应
	stateWaitCreateRole            // 等待 CreateRole 响应
	stateWaitLoginRole             // 等待 LoginRole 响应
	stateReady                     // 完成登录，可压测
	stateStopped                   // 已停止
)

const pktHeaderLen = 8

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
		errCode := pb.EErrorCode_T(binary.LittleEndian.Uint16(header[2:4]))
		bodyLen := binary.LittleEndian.Uint32(header[4:8])

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

		r.Fork("pkt", func() {
			r.recvCount++
			r.onPacket(key, errCode, body)
		})
	}
}

// --------------- Packet Handler ---------------

func (r *Robot) onPacket(key pb.EKey_T, errCode pb.EErrorCode_T, body []byte) {
	r.Debugf("Robot[%d] recv key=%d err=%d state=%d", r.index, key, errCode, r.state)

	switch key {
	case pb.EKey_RspPing:
		return

	case pb.EKey_RspLogin:
		if r.state == stateWaitLogin {
			r.Infof("Robot[%d] login ok", r.index)
			r.sendCreateRole()
		}

	case pb.EKey_RspCreateRole:
		if r.state == stateWaitCreateRole {
			r.Infof("Robot[%d] create role ok", r.index)
			r.sendLoginRole()
		}

	case pb.EKey_RspLoginRole:
		if r.state == stateWaitLoginRole {
			r.Infof("Robot[%d] login role ok → ready", r.index)
			r.state = stateReady
			r.startStressTest()
		}

	case pb.EKey_RspEnterZone:
		r.Debugf("Robot[%d] enter zone response", r.index)
	}

	_ = pb.Unmarshal(key, body) // 仅用于确保协议可反序列化（压测无需业务字段）
}

// --------------- Packet Writer ---------------

func (r *Robot) sendProto(msg proto.Message) {
	if r.conn == nil || r.closed.Load() {
		return
	}
	data, err := pb.MarshalRequest(msg)
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
	r.sendProto(&pb.ReqLogin{Account: r.account})
	r.Debugf("Robot[%d] → Login", r.index)
}

func (r *Robot) sendCreateRole() {
	r.state = stateWaitCreateRole
	r.sendProto(&pb.ReqCreateRole{})
	r.Debugf("Robot[%d] → CreateRole", r.index)
}

func (r *Robot) sendLoginRole() {
	r.state = stateWaitLoginRole
	r.sendProto(&pb.ReqLoginRole{})
	r.Debugf("Robot[%d] → LoginRole", r.index)
}

// --------------- Stress Test ---------------

func (r *Robot) startStressTest() {
	r.pingHandle = r.Tick(5*time.Second, 0, func() {
		if r.state == stateReady {
			r.sendProto(&pb.ReqPing{})
		}
	})

	r.stressHandle = r.Tick(2*time.Second, time.Second, func() {
		if r.state == stateReady {
			r.sendProto(&pb.ReqEnterZone{})
		}
	})

	r.sendProto(&pb.ReqEnterZone{})
}
