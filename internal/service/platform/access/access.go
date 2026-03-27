package access

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[Access, *Access]("Access")
}

type Option struct {
	HttpHost string `snow:"HttpHost"`
	HttpPort int    `snow:"HttpPort"`
}

type GateMeta struct {
	GateId string
	Host   string
	Port   int
	WsHost string
	WsPort int
	WsPath string
}

type GateMetrics struct {
	Connections int64
	Players     int64
}

type gateEntry struct {
	Meta      GateMeta
	Metrics   GateMetrics
	LastBeat  int64
}

type Access struct {
	node.Service

	opt       *Option
	sAuth     node.IProxy
	sAccount  node.IProxy
	sDB       node.IProxy
	httpSrv   *http.Server
	gates     map[string]*gateEntry
	gatesMu   sync.RWMutex
}

func (ss *Access) Construct(opt *Option) {
	if opt == nil {
		opt = &Option{}
	}
	if opt.HttpPort == 0 {
		opt.HttpPort = 9000
	}
	ss.opt = opt
	ss.gates = make(map[string]*gateEntry)
}

func (ss *Access) Start(_ any) {
	ss.sAuth = ss.CreateProxy("Auth")
	ss.sAccount = ss.CreateProxy("Account")
	ss.sDB = ss.CreateProxy("DB")

	mux := http.NewServeMux()
	mux.HandleFunc("/bootstrap", ss.handleBootstrap)
	mux.HandleFunc("/realms", ss.handleRealms)
	mux.HandleFunc("/auth/login", ss.handleLogin)
	mux.HandleFunc("/auth/reconnect", ss.handleReconnect)
	mux.HandleFunc("/account/create-role", ss.handleCreateRole)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := net.JoinHostPort(ss.opt.HttpHost, strconv.Itoa(ss.opt.HttpPort))
	ss.httpSrv = &http.Server{Addr: addr, Handler: mux}
	go func() {
		ss.Infof("Access HTTP listening on %s", addr)
		if err := ss.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ss.Errorf("Access HTTP error: %v", err)
		}
	}()

	ss.EnableRpc()
	ss.Infof("Access service started")
}

func (ss *Access) Stop(_ *sync.WaitGroup) {
	if ss.httpSrv != nil {
		_ = ss.httpSrv.Close()
	}
}

func (ss *Access) RpcStatus(ctx node.IRpcContext) {
	ss.gatesMu.RLock()
	gateCount := len(ss.gates)
	ss.gatesMu.RUnlock()
	ctx.Return(map[string]any{"status": "Access.OK", "gates": gateCount})
}

// --------------- 对内 RPC ---------------

func (ss *Access) RpcVerifyGateTicket(ctx node.IRpcContext, ticket string) {
	if !ss.sAuth.Avail() {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	ss.sAuth.Call("VerifyGateTicket", ticket).
		Then(func(ok bool, account string, realmId int64, roleId int64) {
			ctx.Return(ok, account, realmId, roleId)
		}).
		Catch(func(err error) {
			ss.Errorf("VerifyGateTicket failed: %v", err)
			ctx.Return(false, "", int64(0), int64(0))
		}).Done()
}

func (ss *Access) RpcRegisterGate(_ node.IRpcContext, meta *GateMeta) {
	ss.gatesMu.Lock()
	ss.gates[meta.GateId] = &gateEntry{
		Meta:     *meta,
		LastBeat: time.Now().Unix(),
	}
	ss.gatesMu.Unlock()
	ss.Infof("Gate registered: %s (%s:%d)", meta.GateId, meta.Host, meta.Port)
}

func (ss *Access) RpcHeartbeatGate(_ node.IRpcContext, gateId string, metrics *GateMetrics) {
	ss.gatesMu.Lock()
	if g, ok := ss.gates[gateId]; ok {
		g.Metrics = *metrics
		g.LastBeat = time.Now().Unix()
	}
	ss.gatesMu.Unlock()
}

func (ss *Access) RpcUnregisterGate(_ node.IRpcContext, gateId string) {
	ss.gatesMu.Lock()
	delete(ss.gates, gateId)
	ss.gatesMu.Unlock()
	ss.Infof("Gate unregistered: %s", gateId)
}

// --------------- HTTP handlers ---------------

func (ss *Access) handleBootstrap(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"version":     "1.0.0",
		"maintenance": false,
		"cdnUrl":      "",
		"patchList":   []string{},
	})
}

func (ss *Access) handleRealms(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"realms": []map[string]any{
			{"id": 1, "name": "默认服", "status": "open", "load": "low"},
		},
	})
}

func (ss *Access) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Account  string `json:"account"`
		Password string `json:"password"`
		Device   string `json:"device"`
		RealmId  int64  `json:"realm_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Account == "" {
		http.Error(w, "account required", http.StatusBadRequest)
		return
	}
	if req.RealmId == 0 {
		req.RealmId = 1
	}

	if !ss.sAuth.Avail() || !ss.sAccount.Avail() {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	type loginResult struct {
		Token  string
		Ticket string
		Roles  any
		Gates  []map[string]any
	}

	ss.sAuth.Call("GenAccessToken", req.Account, req.Device).
		Then(func(token string) {
			ss.sAccount.Call("GetRoles", req.Account).
				Then(func(roles any) {
					ss.sAuth.Call("GenGateTicket", req.Account, req.RealmId, int64(0)).
						Then(func(ticket string) {
							ss.Fork("login.done", func() {
								gates := ss.getGateList()
								writeJSON(w, map[string]any{
									"token":  token,
									"ticket": ticket,
									"roles":  roles,
									"gates":  gates,
								})
							})
						}).Catch(func(err error) {
						httpError(w, err)
					}).Done()
				}).Catch(func(err error) {
				httpError(w, err)
			}).Done()
		}).Catch(func(err error) {
		httpError(w, err)
	}).Done()
}

func (ss *Access) handleReconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ReconnectTicket string `json:"reconnect_ticket"`
		RealmId         int64  `json:"realm_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if !ss.sAuth.Avail() {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	ss.sAuth.Call("VerifyAccessToken", req.ReconnectTicket).
		Then(func(ok bool, account string, _ string) {
			if !ok {
				http.Error(w, "invalid ticket", http.StatusUnauthorized)
				return
			}
			ss.sAuth.Call("GenGateTicket", account, req.RealmId, int64(0)).
				Then(func(ticket string) {
					ss.Fork("reconnect.done", func() {
						writeJSON(w, map[string]any{
							"ticket": ticket,
							"gates":  ss.getGateList(),
						})
					})
				}).Catch(func(err error) {
				httpError(w, err)
			}).Done()
		}).Catch(func(err error) {
		httpError(w, err)
	}).Done()
}

func (ss *Access) handleCreateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Account string `json:"account"`
		Cid     int64  `json:"cid"`
		Name    string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if !ss.sAccount.Avail() {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}

	ss.sAccount.Call("CreateRole", req.Account, req.Cid, req.Name).
		Then(func(role any) {
			ss.Fork("createRole.done", func() {
				writeJSON(w, map[string]any{"role": role})
			})
		}).
		Catch(func(err error) {
			httpError(w, err)
		}).Done()
}

func (ss *Access) getGateList() []map[string]any {
	ss.gatesMu.RLock()
	defer ss.gatesMu.RUnlock()
	gates := make([]map[string]any, 0, len(ss.gates))
	for _, g := range ss.gates {
		gates = append(gates, map[string]any{
			"host":        g.Meta.Host,
			"port":        g.Meta.Port,
			"ws_host":     g.Meta.WsHost,
			"ws_port":     g.Meta.WsPort,
			"ws_path":     g.Meta.WsPath,
			"connections": g.Metrics.Connections,
		})
	}
	return gates
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%v", err)})
}
