package mod

import (
	"server/data"
	"server/data/conf"
	"server/lib/matrix"
	"server/lib/uid"
	"server/service/scene/entity/mod/skill"
	"server/service/scene/score"
)

var _ score.IModule = (*SkillManager)(nil)

type SkillManager struct {
	owner score.IEntity

	skills map[int64]*skill.Skill

	NowMs int64
}

func (m *SkillManager) Init(owner score.IEntity, initData data.EntityInitData) {
	m.owner = owner
	m.skills = make(map[int64]*skill.Skill)
}

func (m *SkillManager) Update() {
	m.NowMs += 50

	for _, s := range m.skills {
		skillInst := s
		s.Update(m.NowMs, func(stage skill.Stage, eff conf.EffectCfg, ctx skill.CastContext) {
			m.execEffect(skillInst, stage, eff, ctx)
		})
	}
}

func (m *SkillManager) AddSkill(cfg *conf.CSkill) {
	if cfg == nil {
		return
	}
	if m.skills == nil {
		m.skills = make(map[int64]*skill.Skill)
	}
	m.skills[cfg.Cid] = skill.NewSkill(cfg)
}

func (m *SkillManager) Cast(skillId int64, ctx skill.CastContext) bool {
	rt := m.skills[skillId]
	if rt == nil {
		return false
	}
	return rt.StartCast(m.NowMs, ctx)
}

func (m *SkillManager) Cancel(skillId int64) {
	rt := m.skills[skillId]
	if rt == nil {
		return
	}
	rt.Cancel(m.NowMs)
}

func (m *SkillManager) execEffect(s *skill.Skill, stage skill.Stage, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m.owner

	selector := m.selectTargetCfg(s, stage)
	targets := m.resolveTargets(selector, ctx)

	switch eff.Type {
	case conf.EffectType_Damage:
		m.applyDamage(targets, eff, ctx)
	case conf.EffectType_Heal:
		m.applyHeal(targets, eff, ctx)
	case conf.EffectType_ApplyAura:
		m.applyAura(targets, eff, ctx)
	case conf.EffectType_Dispel:
		m.applyDispel(targets, eff, ctx)
	case conf.EffectType_Steal:
		m.applySteal(targets, eff, ctx)
	case conf.EffectType_Move:
		m.applyMove(targets, eff, ctx)
	case conf.EffectType_Interrupt:
		m.applyInterrupt(targets, eff, ctx)
	case conf.EffectType_Summon:
		m.applySummon(targets, eff, ctx)
	case conf.EffectType_Threat:
		m.applyThreat(targets, eff, ctx)
	case conf.EffectType_SpawnArea:
		m.applySpawnArea(targets, eff, ctx)
	default:
		_ = stage
	}
}

func (m *SkillManager) selectTargetCfg(s *skill.Skill, stage skill.Stage) *conf.TargetCfg {
	if s == nil || s.Cfg == nil {
		return nil
	}

	selectors := s.Cfg.Selectors
	switch stage {
	case skill.Stage_CastStart:
		if selectors.OnCastStart != nil {
			return selectors.OnCastStart
		}
	case skill.Stage_CastFinish:
		if selectors.OnCastFinish != nil {
			return selectors.OnCastFinish
		}
	case skill.Stage_Channel:
		if selectors.OnChannelTick != nil {
			return selectors.OnChannelTick
		}
	case skill.Stage_Hit:
		if selectors.OnHit != nil {
			return selectors.OnHit
		}
	case skill.Stage_Cancel:
		if selectors.OnCancel != nil {
			return selectors.OnCancel
		}
	}

	return &s.Cfg.Target
}

func (m *SkillManager) resolveTargets(cfg *conf.TargetCfg, ctx skill.CastContext) []uid.Uid {
	if cfg == nil {
		return nil
	}
	if m.owner == nil {
		return nil
	}

	if cfg.Relation == conf.TargetRelation_Self {
		if m.owner == nil {
			return nil
		}
		return []uid.Uid{m.owner.GetId()}
	}

	sc := m.owner.GetScene()
	if sc == nil {
		return nil
	}

	if cfg.Mode == conf.TargetMode_Unit {
		if !ctx.TargetId.IsValid() {
			return nil
		}
		if _, ok := sc.GetEntity(ctx.TargetId); !ok {
			return nil
		}
		return []uid.Uid{ctx.TargetId}
	}

	if cfg.Mode == conf.TargetMode_NoTarget {
		if cfg.Shape == conf.ShapeType_Single {
			if m.owner == nil {
				return nil
			}
			return []uid.Uid{m.owner.GetId()}
		}
		if cfg.Shape == conf.ShapeType_Circle {
			pos := m.owner.GetPos()
			if pos == nil {
				return nil
			}
			ctx.X = float32(pos.X)
			ctx.Y = float32(pos.Y)
			cfg = &conf.TargetCfg{Relation: cfg.Relation, Mode: conf.TargetMode_Point, Shape: conf.ShapeType_Circle, Radius: cfg.Radius}
		} else {
			return nil
		}
	}

	if cfg.Mode == conf.TargetMode_Point {
		if cfg.Shape != conf.ShapeType_Circle {
			return nil
		}
		r := float64(cfg.Radius)
		if r <= 0 {
			return nil
		}
		center := matrix.Vector3D{X: float64(ctx.X), Y: float64(ctx.Y), Z: 0}
		r2 := r * r
		res := make([]uid.Uid, 0)
		sc.ForEachEntity(func(id uid.Uid, e score.IEntity) {
			if e == nil {
				return
			}
			p := e.GetPos()
			if p == nil {
				return
			}
			dx := p.X - center.X
			dy := p.Y - center.Y
			if dx*dx+dy*dy <= r2 {
				res = append(res, id)
			}
		})
		return res
	}

	_ = ctx
	return nil
}

func (m *SkillManager) applyDamage(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applyHeal(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applyAura(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applyDispel(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applySteal(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applyMove(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applyInterrupt(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applySummon(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applyThreat(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}

func (m *SkillManager) applySpawnArea(targets []uid.Uid, eff conf.EffectCfg, ctx skill.CastContext) {
	_ = m
	_ = targets
	_ = eff
	_ = ctx
}
