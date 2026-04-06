package robot

import (
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"server/internal/pb"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

type State int32

const (
	StateInit        State = 0
	StateConnecting  State = 1
	StateConnected   State = 2
	StateLoggedIn    State = 3
	StateRoleReady   State = 4
	StateInScene     State = 5
	StateClosed      State = 6
)

type Stats struct {
	SendCount  atomic.Int64
	RecvCount  atomic.Int64
	ErrorCount atomic.Int64
	RTTSum     atomic.Int64
	RTTCount   atomic.Int64
}

type Robot struct {
	index   int
	account string
	wsURL   string
	mgr     *RobotMgr

	conn   *websocket.Conn
	connMu sync.Mutex

	state  State
	roleId int64
	roles  []*pb.RoleSummaryData

	stats    Stats
	stopCh   chan struct{}
	stopped  atomic.Bool
	lastPing time.Time
}

func NewRobot(index int, wsURL string, mgr *RobotMgr) *Robot {
	return &Robot{
		index:   index,
		account: fmt.Sprintf("robot_%d", index),
		wsURL:   wsURL,
		mgr:     mgr,
		stopCh:  make(chan struct{}),
	}
}

func (r *Robot) Run() {
	defer r.cleanup()

	if err := r.connect(); err != nil {
		r.logf("connect failed: %v", err)
		r.stats.ErrorCount.Add(1)
		return
	}
	r.state = StateConnected

	go r.readLoop()

	r.doLogin()

	ticker := time.NewTicker(2 * time.Second)
	pingTicker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	defer pingTicker.Stop()

	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.tick()
		case <-pingTicker.C:
			r.sendPing()
		}
	}
}

func (r *Robot) Stop() {
	if r.stopped.Swap(true) {
		return
	}
	close(r.stopCh)
}

func (r *Robot) connect() error {
	r.state = StateConnecting
	u, err := url.Parse(r.wsURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	r.conn = conn
	r.logf("connected")
	return nil
}

func (r *Robot) cleanup() {
	r.state = StateClosed
	r.connMu.Lock()
	if r.conn != nil {
		_ = r.conn.Close()
		r.conn = nil
	}
	r.connMu.Unlock()
}

func (r *Robot) readLoop() {
	for {
		select {
		case <-r.stopCh:
			return
		default:
		}

		r.connMu.Lock()
		conn := r.conn
		r.connMu.Unlock()
		if conn == nil {
			return
		}

		_, data, err := conn.ReadMessage()
		if err != nil {
			if !r.stopped.Load() {
				r.logf("read error: %v", err)
				r.stats.ErrorCount.Add(1)
			}
			r.Stop()
			return
		}

		r.stats.RecvCount.Add(1)
		r.handleMessage(data)
	}
}

func (r *Robot) handleMessage(data []byte) {
	if len(data) < 4 {
		return
	}
	key := pb.EKey_T(binary.LittleEndian.Uint16(data[0:2]))
	errCode := pb.EErrorCode_T(binary.LittleEndian.Uint16(data[2:4]))

	var body []byte
	if len(data) >= 8 {
		bodyLen := binary.LittleEndian.Uint32(data[4:8])
		if len(data) >= int(8+bodyLen) {
			body = data[8 : 8+bodyLen]
		}
	}

	keyName := pb.EKey_T_name[int32(key)]
	if errCode != pb.EErrorCode_Ok {
		r.logf("recv %s err=%d", keyName, errCode)
		r.stats.ErrorCount.Add(1)
		return
	}

	switch key {
	case pb.EKey_RspLogin:
		r.onRspLogin(body)
	case pb.EKey_RspCreateRole:
		r.onRspCreateRole(body)
	case pb.EKey_RspLoginRole:
		r.onRspLoginRole()
	case pb.EKey_RspEnterZone:
		r.onRspEnterZone()
	case pb.EKey_RspPing:
		rtt := time.Since(r.lastPing)
		r.stats.RTTSum.Add(rtt.Milliseconds())
		r.stats.RTTCount.Add(1)
		r.logf("recv RspPing rtt=%dms", rtt.Milliseconds())
	case pb.EKey_RspMove:
		// ok
	default:
		r.logf("recv unhandled key=%s(%d)", keyName, key)
	}
}

// --------------- Actions ---------------

func (r *Robot) doLogin() {
	r.sendMsg(&pb.ReqLogin{
		Account: r.account,
	})
}

func (r *Robot) onRspLogin(body []byte) {
	rsp := &pb.RspLogin{}
	if len(body) > 0 {
		_ = proto.Unmarshal(body, rsp)
	}
	r.state = StateLoggedIn
	r.roles = rsp.GetRoles()
	r.logf("logged in, account=%s roles=%d", rsp.GetAccount(), len(r.roles))

	if len(r.roles) == 0 {
		r.sendMsg(&pb.ReqCreateRole{
			Cid:  1,
			Name: fmt.Sprintf("bot_%d_%d", r.index, time.Now().UnixMilli()%10000),
		})
	} else {
		r.roleId = r.roles[0].GetId()
		r.sendMsg(&pb.ReqLoginRole{RoleId: r.roleId})
	}
}

func (r *Robot) onRspCreateRole(body []byte) {
	rsp := &pb.RspCreateRole{}
	if len(body) > 0 {
		_ = proto.Unmarshal(body, rsp)
	}
	if rsp.GetRole() != nil {
		r.roleId = rsp.GetRole().GetId()
		r.roles = append(r.roles, rsp.GetRole())
		r.logf("role created: %d", r.roleId)
		r.sendMsg(&pb.ReqLoginRole{RoleId: r.roleId})
	} else {
		r.logf("create role returned nil")
		r.stats.ErrorCount.Add(1)
	}
}

func (r *Robot) onRspLoginRole() {
	r.state = StateRoleReady
	r.logf("role ready: %d, sending test ping to actor", r.roleId)
	r.sendPing()
	r.sendMsg(&pb.ReqEnterZone{})
}

func (r *Robot) onRspEnterZone() {
	r.state = StateInScene
	r.logf("entered scene")
}

func (r *Robot) tick() {
	switch r.state {
	case StateInScene:
		r.doMove()
	case StateRoleReady:
		r.sendMsg(&pb.ReqEnterZone{})
	}
}

func (r *Robot) doMove() {
	r.sendMsg(&pb.ReqMove{
		Pos: &pb.Vector{
			X: int64(rand.IntN(1000)),
			Y: 0,
			Z: int64(rand.IntN(1000)),
		},
	})
}

func (r *Robot) sendPing() {
	if r.state < StateConnected {
		return
	}
	r.lastPing = time.Now()
	r.sendMsg(&pb.ReqPing{})
}

// --------------- Send ---------------

func (r *Robot) sendMsg(msg proto.Message) {
	data, err := pb.MarshalRequest(msg)
	if err != nil {
		r.logf("marshal error: %v", err)
		r.stats.ErrorCount.Add(1)
		return
	}

	r.connMu.Lock()
	conn := r.conn
	r.connMu.Unlock()
	if conn == nil {
		return
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		r.logf("write error: %v", err)
		r.stats.ErrorCount.Add(1)
		return
	}
	r.stats.SendCount.Add(1)
}

func (r *Robot) logf(format string, args ...any) {
	if r.mgr != nil && r.mgr.verbose {
		fmt.Printf("[Robot#%d] %s\n", r.index, fmt.Sprintf(format, args...))
	}
}
