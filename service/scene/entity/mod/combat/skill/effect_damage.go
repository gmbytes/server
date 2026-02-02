package skill

import (
	"server/data/conf"
	"server/service/scene/score"
	"time"
)

type DamageEffect struct {
	cfg conf.EffectCfg
}

func NewDamageEffect(cfg conf.EffectCfg) *DamageEffect {
	return &DamageEffect{cfg: cfg}
}

func (e *DamageEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *DamageEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *DamageEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *DamageEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
