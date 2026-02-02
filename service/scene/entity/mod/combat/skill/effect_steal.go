package skill

import (
	"server/data/conf"
	"server/service/scene/score"
	"time"
)

type StealEffect struct {
	cfg conf.EffectCfg
}

func NewStealEffect(cfg conf.EffectCfg) *StealEffect {
	return &StealEffect{cfg: cfg}
}

func (e *StealEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *StealEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *StealEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *StealEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
