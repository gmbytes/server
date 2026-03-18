package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"


	"sync"
	"sync/atomic"


	"github.com/gmbytes/snow/pkg/host"
	"github.com/gmbytes/snow/pkg/option"
	"github.com/gmbytes/snow/pkg/xnet/transport"
	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[Gate, *Gate]("Gate", func(b host.IBuilder) {
		host.AddOption[*Option](b, "Gate")
	})
}

type Option struct {
	TcpListenHost    string `snow:"TcpListenHost"`
	TcpListenPort    int    `snow:"TcpListenPort"`
	WsListenHost     string `snow:"WsListenHost"`
	WsListenPort     int    `snow:"WsListenPort"`
	WsPath           string `snow:"WsPath"`
	HttpListenHost   string `snow:"HttpListenHost"`
	HttpListenPort   int    `snow:"HttpListenPort"`
	SessionSendQueue int    `snow:"SessionSendQueue"`
	ReadBufferSize   int    `snow:"ReadBufferSize"`
	MaxConnPerIP     int32  `snow:"MaxConnPerIP"`
	MaxReadPerSec    int32  `snow:"MaxReadPerSec"`
	MaxHTTPBodyBytes int64  `snow:"MaxHTTPBodyBytes"`
}

type Gate struct {
	node.Service
	opt       *Option
	listeners []net.Listener
	httpSrv   *http.Server
	wg        sync.WaitGroup
	closed    atomic.Bool

	sessionsByConnId sync.Map // connId -> *session
	sessionsByRoleId sync.Map // roleId -> *session
	ipConnCount      sync.Map // ip -> *atomic.Int32
	nextConnSeq      atomic.Uint64

	gameProxy node.IProxy
}

func (s *Gate) Construct(opt *option.Option[*Option]) {
	s.opt = opt.Get()
}

func (s *Gate) Start(_ any) {
	s.gameProxy = s.CreateProxy("Game")

	cfg := &transport.Config{
		TCPHost: s.opt.TcpListenHost,
		TCPPort: s.opt.TcpListenPort,
		WSHost:  s.opt.WsListenHost,
		WSPort:  s.opt.WsListenPort,
		WSPath:  s.opt.WsPath,
	}
	listeners, err := transport.NewListeners(cfg)
	if err != nil {
		s.Fatalf("create listeners failed: %v", err)
		return
	}
	s.listeners = listeners
	for _, ln := range listeners {
		s.wg.Add(1)
		go s.acceptLoop(ln)
	}
	s.startHTTP()
	s.EnableRpc()
}

func (s *Gate) Stop(_ *sync.WaitGroup) {
	s.Infof("gate service stopping")
	s.closed.Store(true)
	for _, ln := range s.listeners {
		_ = ln.Close()
	}
	s.sessionsByConnId.Range(func(_, v any) bool {
		v.(*session).close()
		return true
	})
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = s.httpSrv.Shutdown(ctx)
		cancel()
	}
}

func (s *Gate) AfterStop() {
	s.wg.Wait()
	s.Infof("gate service stopped")
}

// --------------- RPC: Game -> Gate ---------------

func (s *Gate) RpcStatus(ctx node.IRpcContext) {
	var conns int64
	s.sessionsByConnId.Range(func(_, _ any) bool {
		conns++
		return true
	})
	ctx.Return(map[string]any{
		"status":      "Gate.OK",
		"connections": conns,
	})
}

func (s *Gate) RpcSendToClient(_ node.IRpcContext, connId uint64, data []byte) {
	if v, ok := s.sessionsByConnId.Load(connId); ok {
		v.(*session).send(data)
	}
}

func (s *Gate) RpcKickClient(_ node.IRpcContext, connId uint64) {
	if v, ok := s.sessionsByConnId.Load(connId); ok {
		v.(*session).close()
	}
}

func (s *Gate) RpcBindRole(_ node.IRpcContext, roleId int64, connId uint64) {
	if roleId == 0 {
		return
	}
	if v, ok := s.sessionsByConnId.Load(connId); ok {
		sess := v.(*session)
		if old, exists := s.sessionsByRoleId.Load(roleId); exists {
			oldSess := old.(*session)
			if oldSess != sess {
				oldSess.close()
			}
		}
		sess.roleId.Store(roleId)
		s.sessionsByRoleId.Store(roleId, sess)
	}
}

func (s *Gate) RpcSendToRole(_ node.IRpcContext, roleId int64, data []byte) {
	if v, ok := s.sessionsByRoleId.Load(roleId); ok {
		v.(*session).send(data)
	}
}

// --------------- Accept / Session ---------------

func (s *Gate) acceptLoop(ln net.Listener) {
	defer s.wg.Done()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if transport.IsListenerClosedError(err) {
				return
			}
			s.Errorf("accept error: %v", err)
			continue
		}
		if s.closed.Load() {
			_ = conn.Close()
			return
		}
		remoteIP := remoteIPFromAddr(conn.RemoteAddr())
		if !s.tryAddConnByIP(remoteIP) {
			_ = conn.Close()
			continue
		}
		sess := s.newSession(conn, remoteIP)
		s.wg.Add(1)
		go sess.serve()
	}
}

// --------------- HTTP: stateless forward ---------------

func (s *Gate) startHTTP() {
	if s.opt.HttpListenPort <= 0 {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/forward", s.handleForward)
	s.httpSrv = &http.Server{
		Addr:    net.JoinHostPort(s.opt.HttpListenHost, strconv.Itoa(s.opt.HttpListenPort)),
		Handler: mux,
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Errorf("http serve error: %v", err)
		}
	}()
}

func (s *Gate) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// handleForward 处理 HTTP 无状态转发：客户端通过 HTTP POST 发送单个请求包，
// Gate 为其分配临时 connId 转发给 Game，Game 处理后通过 RPC 将结果写回。
// 支持通过 X-Conn-Id 复用已有连接 ID（用于需保持上下文的场景）。
func (s *Gate) handleForward(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, s.opt.MaxHTTPBodyBytes))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	connID := s.nextConnID()
	if h := r.Header.Get("X-Conn-Id"); h != "" {
		if parsed, e := strconv.ParseUint(h, 10, 64); e == nil {
			connID = parsed
		}
	}
	remoteIP := r.Header.Get("X-Forwarded-For")
	if remoteIP == "" {
		host, _, splitErr := net.SplitHostPort(r.RemoteAddr)
		if splitErr == nil {
			remoteIP = host
		} else {
			remoteIP = r.RemoteAddr
		}
	}

	if !s.gameProxy.Avail() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	s.gameProxy.Call("HandleClientMsg", connID, remoteIP, body).
		Then(func(resp []byte) {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(resp)
		}).
		Catch(func(err error) {
			s.Errorf("http forward failed connId=%d err=%v", connID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}).
		Done()
}

// --------------- Helpers ---------------

func (s *Gate) newSession(conn net.Conn, remoteIP string) *session {
	connID := s.nextConnID()
	sess := &session{
		gate:     s,
		id:       connID,
		conn:     conn,
		remoteIP: remoteIP,
		sendQ:    make(chan []byte, s.opt.SessionSendQueue),
	}
	s.sessionsByConnId.Store(connID, sess)
	return sess
}

func (s *Gate) nextConnID() uint64 {
	seq := s.nextConnSeq.Add(1) & ((1 << 20) - 1)
	return (uint64(time.Now().UnixMilli()) << 20) | seq
}

func (s *Gate) tryAddConnByIP(ip string) bool {
	v, _ := s.ipConnCount.LoadOrStore(ip, &atomic.Int32{})
	counter := v.(*atomic.Int32)
	if counter.Add(1) <= s.opt.MaxConnPerIP {
		return true
	}
	counter.Add(-1)
	return false
}

func (s *Gate) decConnByIP(ip string) {
	v, ok := s.ipConnCount.Load(ip)
	if !ok {
		return
	}
	if v.(*atomic.Int32).Add(-1) <= 0 {
		s.ipConnCount.Delete(ip)
	}
}

func (s *Gate) onSessionOpened(sess *session) {
	if !s.gameProxy.Avail() {
		return
	}
	s.gameProxy.Call("OnClientConnect", sess.id, sess.remoteIP).
		Catch(func(err error) {
			s.Errorf("notify client connect failed connId=%d err=%v", sess.id, err)
		}).Done()
}

func (s *Gate) onSessionClosed(sess *session) {
	if roleID := sess.roleId.Load(); roleID != 0 {
		s.sessionsByRoleId.Delete(roleID)
	}
	s.sessionsByConnId.Delete(sess.id)
	s.decConnByIP(sess.remoteIP)

	if !s.gameProxy.Avail() {
		return
	}
	s.gameProxy.Call("OnClientDisconnect", sess.id).
		Catch(func(err error) {
			s.Errorf("notify client disconnect failed connId=%d err=%v", sess.id, err)
		}).Done()
}

func (s *Gate) forwardToGame(connID uint64, remoteIP string, payload []byte) {
	if !s.gameProxy.Avail() {
		s.Warnf("game proxy unavailable, connId=%d", connID)
		return
	}
	s.gameProxy.Call("HandleClientMsg", connID, remoteIP, payload).
		Catch(func(err error) {
			s.Errorf("forward client msg failed, connId=%d err=%v", connID, err)
		}).Done()
}

func remoteIPFromAddr(addr net.Addr) string {
	if addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err == nil {
		return host
	}
	return fmt.Sprintf("%v", addr)
}
