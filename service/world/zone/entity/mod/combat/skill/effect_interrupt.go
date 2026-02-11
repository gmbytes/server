package skill

import (
	"server/data/conf"
	"server/service/world/zone/izone"
	"time"
)

type InterruptEffect struct {
	cfg conf.EffectCfg
}

func NewInterruptEffect(cfg conf.EffectCfg) *InterruptEffect {
	return &InterruptEffect{cfg: cfg}
}

func (e *InterruptEffect) Begin(ctx *SkillContext, causer izone.IEntity, targets []izone.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *InterruptEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *InterruptEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *InterruptEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
