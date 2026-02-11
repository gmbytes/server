package skill

import (
	"server/data/conf"
	"server/service/world/zone/izone"
	"time"
)

type SummonEffect struct {
	cfg conf.EffectCfg
}

func NewSummonEffect(cfg conf.EffectCfg) *SummonEffect {
	return &SummonEffect{cfg: cfg}
}

func (e *SummonEffect) Begin(ctx *SkillContext, causer izone.IEntity, targets []izone.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *SummonEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *SummonEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *SummonEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
