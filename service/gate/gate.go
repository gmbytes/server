package gate

import (
	"sync"

	"github.com/gmbytes/snow/routines/node"
)

// Gate 基于 snow node.Service 的网关服务，可被 RPC/HTTP 调用。
type Gate struct {
	node.Service
}

// Start 服务启动时调用，启用 RPC 后开始处理请求。
func (s *Gate) Start(_ any) {
	s.Infof("gate service starting")
	s.EnableRpc()
	s.Infof("gate service started")
}

// Stop 服务关闭时调用，与消息处理同线程。
func (s *Gate) Stop(_ *sync.WaitGroup) {
	s.Infof("gate service stopping")
}

// AfterStop 服务完全关闭后调用，此时不再处理任何消息。
func (s *Gate) AfterStop() {
	s.Infof("gate service stopped")
}

// RpcStatus 可选：覆写默认状态 RPC，用于健康检查。
func (s *Gate) RpcStatus(ctx node.IRpcContext) {
	ctx.Return("Gate.OK")
}
