package scene

import (
	"fmt"
	"sync"

	"server/internal/service/realm/scenemgr"

	"github.com/gmbytes/snow/pkg/routines/node"
)

func init() {
	node.Register[SceneService, *SceneService]("Scene")
}

type SceneService struct {
	node.Service

	sceneId   int64
	sceneType int
	mapId     int32
	impl      IScene
}

func (ss *SceneService) Start(arg any) {
	if initData, ok := arg.(*scenemgr.SceneInitData); ok && initData != nil {
		ss.sceneId = initData.SceneId
		ss.sceneType = initData.SceneType
		ss.mapId = initData.MapId
	}

	ss.impl = NewScene(ss.sceneId, ss.sceneType, ss.mapId)
	ss.EnableRpc()
	ss.Infof("Scene service started: id=%d type=%d map=%d", ss.sceneId, ss.sceneType, ss.mapId)
}

func (ss *SceneService) Stop(_ *sync.WaitGroup) {
	if ss.impl != nil {
		ss.impl.Destroy()
	}
	ss.Infof("Scene service stopped: id=%d", ss.sceneId)
}

func (ss *SceneService) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{
		"status":  "Scene.OK",
		"sceneId": ss.sceneId,
		"type":    ss.sceneType,
		"mapId":   ss.mapId,
	})
}

func (ss *SceneService) RpcJoinScene(ctx node.IRpcContext, sceneId int64, roleId int64, snapshot []byte) {
	if ss.impl == nil {
		ctx.Return(&JoinResult{Err: fmt.Errorf("scene not initialized")})
		return
	}
	initData, err := ss.impl.Join(roleId, snapshot)
	if err != nil {
		ctx.Return(&JoinResult{Err: err})
		return
	}
	ctx.Return(&JoinResult{InitData: initData})
}

func (ss *SceneService) RpcLeaveScene(ctx node.IRpcContext, sceneId int64, roleId int64, reason int) {
	if ss.impl == nil {
		ctx.Return(&LeaveResult{Err: fmt.Errorf("scene not initialized")})
		return
	}
	carry, err := ss.impl.Leave(roleId, reason)
	if err != nil {
		ctx.Return(&LeaveResult{Err: err, Reason: reason})
		return
	}
	ctx.Return(&LeaveResult{CarryData: carry, Reason: reason})
}

func (ss *SceneService) RpcSceneMessage(_ node.IRpcContext, sceneId int64, roleId int64, key int, data []byte) {
	if ss.impl != nil {
		ss.impl.OnMessage(roleId, key, data)
	}
}
