package skill

import (
	"server/data/conf"
	"server/service/scene/score"
	"time"
)

type HealEffect struct {
	cfg conf.EffectCfg
}

func NewHealEffect(cfg conf.EffectCfg) *HealEffect {
	return &HealEffect{cfg: cfg}
}

func (e *HealEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *HealEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *HealEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *HealEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
