package scenemgr

import (
	"fmt"
	"sync"

	"server/pkg/uid"

	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[SceneMgr, *SceneMgr]("SceneMgr")
}

type AllocSceneReq struct {
	SceneType int
	MapId     int32
	CrossNode bool
	Extra     []byte
}

type FindSceneReq struct {
	SceneId int64
}

type SceneProxyInfo struct {
	SceneId   int64
	NodeAddr  node.Addr
	SAddr     int32
	SceneType int
	Err       error
}

type sceneRecord struct {
	SceneId   int64
	SceneType int
	MapId     int32
	SAddr     int32
	NodeAddr  node.Addr
	Players   int
}

type SceneMgr struct {
	node.Service

	scenes map[int64]*sceneRecord // sceneId -> record
}

func (sm *SceneMgr) Start(_ any) {
	sm.scenes = make(map[int64]*sceneRecord)
	sm.EnableRpc()
	sm.Infof("SceneMgr service started")
}

func (sm *SceneMgr) Stop(_ *sync.WaitGroup) {
	sm.Infof("SceneMgr service stopping")
	for _, rec := range sm.scenes {
		node.StopService(rec.SAddr)
	}
}

func (sm *SceneMgr) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{
		"status": "SceneMgr.OK",
		"scenes": len(sm.scenes),
	})
}

func (sm *SceneMgr) RpcAllocScene(ctx node.IRpcContext, req *AllocSceneReq) {
	for _, rec := range sm.scenes {
		if rec.SceneType == req.SceneType && rec.MapId == req.MapId {
			ctx.Return(&SceneProxyInfo{
				SceneId:   rec.SceneId,
				NodeAddr:  rec.NodeAddr,
				SAddr:     rec.SAddr,
				SceneType: rec.SceneType,
			})
			return
		}
	}

	sAddr, err := node.NewService("Scene")
	if err != nil {
		ctx.Return(&SceneProxyInfo{Err: fmt.Errorf("create scene failed: %w", err)})
		return
	}

	sceneId := int64(uid.Gen())
	rec := &sceneRecord{
		SceneId:   sceneId,
		SceneType: req.SceneType,
		MapId:     req.MapId,
		SAddr:     sAddr,
		NodeAddr:  node.AddrLocal,
	}
	sm.scenes[sceneId] = rec

	node.StartService(sAddr, &SceneInitData{
		SceneId:   sceneId,
		SceneType: req.SceneType,
		MapId:     req.MapId,
	})

	sm.Infof("Scene created: id=%d type=%d map=%d sAddr=%d", sceneId, req.SceneType, req.MapId, sAddr)

	ctx.Return(&SceneProxyInfo{
		SceneId:   sceneId,
		NodeAddr:  rec.NodeAddr,
		SAddr:     rec.SAddr,
		SceneType: rec.SceneType,
	})
}

func (sm *SceneMgr) RpcFindScene(ctx node.IRpcContext, req *FindSceneReq) {
	rec, ok := sm.scenes[req.SceneId]
	if !ok {
		ctx.Return(&SceneProxyInfo{Err: fmt.Errorf("scene %d not found", req.SceneId)})
		return
	}
	ctx.Return(&SceneProxyInfo{
		SceneId:   rec.SceneId,
		NodeAddr:  rec.NodeAddr,
		SAddr:     rec.SAddr,
		SceneType: rec.SceneType,
	})
}

func (sm *SceneMgr) RpcDeallocScene(_ node.IRpcContext, sceneId int64, reason int) {
	rec, ok := sm.scenes[sceneId]
	if !ok {
		return
	}
	sm.Infof("Scene dealloc: id=%d reason=%d", sceneId, reason)
	node.StopService(rec.SAddr)
	delete(sm.scenes, sceneId)
}

type SceneInitData struct {
	SceneId   int64
	SceneType int
	MapId     int32
}
