package scene

import (
	"encoding/json"
)

type DefaultScene struct {
	id        int64
	sceneType SceneType
	mapId     int32
	players   map[int64]*playerEntry
}

type playerEntry struct {
	RoleId   int64
	Snapshot []byte
}

func NewScene(id int64, sceneType SceneType, mapId int32) IScene {
	return &DefaultScene{
		id:        id,
		sceneType: sceneType,
		mapId:     mapId,
		players:   make(map[int64]*playerEntry),
	}
}

func (s *DefaultScene) Id() int64      { return s.id }
func (s *DefaultScene) Type() SceneType { return s.sceneType }

func (s *DefaultScene) Join(roleId int64, snapshot []byte) ([]byte, error) {
	s.players[roleId] = &playerEntry{
		RoleId:   roleId,
		Snapshot: snapshot,
	}
	initData, _ := json.Marshal(map[string]any{
		"sceneId":   s.id,
		"sceneType": s.sceneType,
		"mapId":     s.mapId,
		"players":   len(s.players),
	})
	return initData, nil
}

func (s *DefaultScene) Leave(roleId int64, reason int) (*EntityCarryData, error) {
	delete(s.players, roleId)
	return &EntityCarryData{}, nil
}

func (s *DefaultScene) OnMessage(roleId int64, key int, data []byte) {
}

func (s *DefaultScene) Update(deltaMs int64) {
}

func (s *DefaultScene) Destroy() {
	s.players = nil
}
