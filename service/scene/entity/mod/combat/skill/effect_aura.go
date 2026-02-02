package skill

import (
	"server/data/conf"
	"server/service/scene/score"
	"time"
)

type AuraEffect struct {
	cfg conf.EffectCfg
}

func NewAuraEffect(cfg conf.EffectCfg) *AuraEffect {
	return &AuraEffect{cfg: cfg}
}

func (e *AuraEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *AuraEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *AuraEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *AuraEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
