package skill

import (
	"server/data/conf"
	"server/service/world/zone/izone"
	"time"
)

type DispelEffect struct {
	cfg conf.EffectCfg
}

func NewDispelEffect(cfg conf.EffectCfg) *DispelEffect {
	return &DispelEffect{cfg: cfg}
}

func (e *DispelEffect) Begin(ctx *SkillContext, causer izone.IEntity, targets []izone.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *DispelEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *DispelEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *DispelEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
