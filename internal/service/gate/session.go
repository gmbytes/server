package gate

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// pktHeaderLen 客户端包头长度: 2 bytes key + 2 bytes err + 4 bytes body length
const pktHeaderLen = 8

type session struct {
	gate          *Gate
	id            uint64
	conn          net.Conn
	remoteIP      string
	sendQ         chan []byte
	sendMu        sync.RWMutex
	roleId        atomic.Int64
	closed        atomic.Bool
	readQPS       int32
	readQPSSecond int64
}

// serve 是长连接（TCP/WebSocket）的主循环：
// 通知 Game 新连接 → 读取客户端包 → 转发给 Game → 连接关闭时通知 Game。
func (s *session) serve() {
	defer s.gate.wg.Done()
	defer s.close()

	s.gate.onSessionOpened(s)

	s.gate.wg.Add(1)
	go s.writeLoop()

	header := make([]byte, pktHeaderLen)
	for {
		if _, err := io.ReadFull(s.conn, header); err != nil {
			if err != io.EOF && !s.closed.Load() {
				s.gate.Errorf("session %d read header: %v", s.id, err)
			}
			return
		}
		if !s.allowRead() {
			s.gate.Warnf("session %d read over limit", s.id)
			return
		}

		bodyLen := binary.LittleEndian.Uint32(header[4:8])
		pkt := make([]byte, pktHeaderLen+bodyLen)
		copy(pkt[:pktHeaderLen], header)
		if bodyLen > 0 {
			if _, err := io.ReadFull(s.conn, pkt[pktHeaderLen:]); err != nil {
				if err != io.EOF && !s.closed.Load() {
					s.gate.Errorf("session %d read body: %v", s.id, err)
				}
				return
			}
		}

		s.gate.forwardToGame(s.id, s.remoteIP, pkt)
	}
}

func (s *session) writeLoop() {
	defer s.gate.wg.Done()
	for data := range s.sendQ {
		if _, err := s.conn.Write(data); err != nil {
			if !s.closed.Load() {
				s.close()
			}
			return
		}
	}
}

func (s *session) send(data []byte) {
	if len(data) == 0 {
		return
	}
	payload := append([]byte(nil), data...)
	s.sendMu.RLock()
	defer s.sendMu.RUnlock()
	if s.closed.Load() {
		return
	}
	select {
	case s.sendQ <- payload:
	default:
		go s.close()
	}
}

func (s *session) close() {
	if s.closed.Swap(true) {
		return
	}
	s.sendMu.Lock()
	close(s.sendQ)
	s.sendMu.Unlock()
	_ = s.conn.Close()
	s.gate.onSessionClosed(s)
}

func (s *session) allowRead() bool {
	limit := s.gate.opt.MaxReadPerSec
	if limit <= 0 {
		return true
	}
	nowSec := time.Now().Unix()
	if s.readQPSSecond != nowSec {
		s.readQPSSecond = nowSec
		s.readQPS = 0
	}
	s.readQPS++
	return s.readQPS <= limit
}
