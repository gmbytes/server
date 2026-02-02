package skill

import (
	"server/data/conf"
	"server/service/scene/score"
	"time"
)

type SummonEffect struct {
	cfg conf.EffectCfg
}

func NewSummonEffect(cfg conf.EffectCfg) *SummonEffect {
	return &SummonEffect{cfg: cfg}
}

func (e *SummonEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
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
