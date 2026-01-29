package mod

import (
	"server/data"
	score2 "server/service/scene/score"
)

var _ score2.IModule = (*BattleManager)(nil)

type BattleManager struct {
	owner score2.IEntity
	attrs *data.Attrs

	hp int64
}

func (m *BattleManager) Init(owner score2.IEntity, initData data.EntityInitData) {
	m.owner = owner
	m.attrs = initData.Attrs
}

func (m *BattleManager) Update() {

}
