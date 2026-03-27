package cross

import (
	"sync"

	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[Cross, *Cross]("Cross")
}

type Cross struct {
	node.Service
}

func (ss *Cross) Start(_ any) {
	ss.EnableRpc()
	ss.Infof("Cross service started (placeholder)")
}

func (ss *Cross) Stop(_ *sync.WaitGroup) {}

func (ss *Cross) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{"status": "Cross.OK"})
}
