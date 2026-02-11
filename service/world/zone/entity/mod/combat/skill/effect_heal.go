package skill

import (
	"server/data/conf"
	"server/service/world/zone/izone"
	"time"
)

type HealEffect struct {
	cfg conf.EffectCfg
}

func NewHealEffect(cfg conf.EffectCfg) *HealEffect {
	return &HealEffect{cfg: cfg}
}

func (e *HealEffect) Begin(ctx *SkillContext, causer izone.IEntity, targets []izone.IEntity) {
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
