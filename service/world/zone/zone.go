package zone

import (
	"sync"

	"server/lib/container"
	"server/lib/uid"
	"server/service/world/zone/izone"

	"github.com/gmbytes/snow/routines/node"
)

var _ izone.IZone = (*Zone)(nil)

// Zone 基于 snow node.Service 的区服逻辑服务，可被 RPC/HTTP 调用。
type Zone struct {
	node.Service
	entities *container.LMap[uid.Uid, izone.IEntity]
}

func (ss *Zone) Init() {
	ss.entities = container.NewLMap[uid.Uid, izone.IEntity]()
}

func (ss *Zone) AddEntity(e izone.IEntity) {
	if e == nil {
		return
	}
	ss.entities.Set(e.GetId(), e)
}

func (ss *Zone) RemoveEntity(id uid.Uid) {
	ss.entities.Delete(id)
}

func (ss *Zone) GetEntity(id uid.Uid) (izone.IEntity, bool) {
	return ss.entities.Get(id)
}

func (ss *Zone) ForEach(fn func(e izone.IEntity)) {
	if fn == nil {
		return
	}
	ss.entities.ForEach(fn)
}

// Start 服务启动时调用，启用 RPC 后开始处理请求。
func (ss *Zone) Start(_ any) {
	ss.Init()
	ss.Infof("zone service starting")
	ss.EnableRpc()
	ss.Infof("zone service started")
}

// Stop 服务关闭时调用，与消息处理同线程。
func (ss *Zone) Stop(_ *sync.WaitGroup) {
	ss.Infof("zone service stopping")
}

// AfterStop 服务完全关闭后调用，此时不再处理任何消息。
func (ss *Zone) AfterStop() {
	ss.Infof("zone service stopped")
}

// RpcStatus 可选：覆写默认状态 RPC，用于健康检查。
func (ss *Zone) RpcStatus(ctx node.IRpcContext) {
	ctx.Return("Zone.OK")
}
