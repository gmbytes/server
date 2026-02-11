package skill

import (
	"server/data/conf"
	"server/service/world/zone/izone"
	"time"
)

type SpawnAreaEffect struct {
	cfg conf.EffectCfg
}

func NewSpawnAreaEffect(cfg conf.EffectCfg) *SpawnAreaEffect {
	return &SpawnAreaEffect{cfg: cfg}
}

func (e *SpawnAreaEffect) Begin(ctx *SkillContext, causer izone.IEntity, targets []izone.IEntity) {
	_ = ctx
	_ = causer
	_ = targets
	_ = e.cfg
}

func (e *SpawnAreaEffect) Update(ctx *SkillContext, delta time.Duration) {
	_ = ctx
	_ = delta
}

func (e *SpawnAreaEffect) End(ctx *SkillContext) {
	_ = ctx
}

func (e *SpawnAreaEffect) Revert(ctx *SkillContext) {
	_ = ctx
}
