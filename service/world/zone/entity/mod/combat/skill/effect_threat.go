package skill

import (
	"server/data/conf"
	"server/service/world/zone/izone"
	"time"
)

type ThreatEffect struct {
	cfg conf.EffectCfg
}

func NewThreatEffect(cfg conf.EffectCfg) *ThreatEffect {
	return &ThreatEffect{cfg: cfg}
}

func (e *ThreatEffect) Begin(ctx *SkillContext, causer izone.IEntity, targets []izone.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *ThreatEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *ThreatEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *ThreatEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
