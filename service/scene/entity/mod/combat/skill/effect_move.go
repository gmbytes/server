package skill

import (
	"server/data/conf"
	"server/service/scene/score"
	"time"
)

type MoveEffect struct {
	cfg conf.EffectCfg
}

func NewMoveEffect(cfg conf.EffectCfg) *MoveEffect {
	return &MoveEffect{cfg: cfg}
}

func (e *MoveEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *MoveEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *MoveEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *MoveEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
