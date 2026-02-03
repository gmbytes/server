package combat

import (
	"server/data/conf"
	"server/lib/container"
	"server/lib/matrix"
	"server/lib/uid"
	"server/pb"
	"server/service/scene/entity/mod/combat/skill"
	"server/service/scene/score"
)

type SkillManager struct {
	*CombatManager

	owner score.IEntity

	skills *container.LMap[int64, *skill.Skill]

	NowMs int64
}

func newSkillManager(combatMgr *CombatManager) *SkillManager {
	ret := &SkillManager{
		CombatManager: combatMgr,
		owner:         combatMgr.owner,
		skills:        container.NewLMap[int64, *skill.Skill](),
	}
	ret.Init()

	return ret
}

func (m *SkillManager) Init() {

}

func (m *SkillManager) Update(deltaMs int64) {
	m.NowMs += deltaMs

	m.skills.ForEach(func(s *skill.Skill) {
		s.Update(m.NowMs, func(stage skill.Stage, eff conf.EffectCfg, ctx *skill.SkillContext) {
			m.execEffect(s, stage, eff, ctx)
		})
	})
}

func (m *SkillManager) AddSkill(cfg *conf.CSkill) {
	if cfg == nil {
		return
	}
	m.skills.Set(cfg.Cid, skill.NewSkill(cfg))
}

func (m *SkillManager) Cast(skillId int64, req *pb.ReqCastSkill) bool {
	rt, ok := m.skills.Get(skillId)
	if !ok {
		return false
	}

	ctx := skill.NewSkillContext(m.owner, req, 1)
	return rt.StartCast(m.NowMs, ctx)
}

func (m *SkillManager) Cancel(skillId int64) {
	rt, ok := m.skills.Get(skillId)
	if !ok {
		return
	}
	rt.Cancel(m.NowMs)
}

func (m *SkillManager) execEffect(s *skill.Skill, stage skill.Stage, eff conf.EffectCfg, ctx *skill.SkillContext) {
	if m.CombatManager == nil {
		return
	}

	targetCfg := m.selectTargetCfg(s, stage)
	if targetCfg == nil {
		return
	}

	targets := m.selectTargets(targetCfg, ctx)
	if len(targets) == 0 {
		return
	}

	m.CombatManager.ExecuteEffect(eff, ctx, m.owner, targets)
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

func (m *SkillManager) selectTargets(cfg *conf.TargetCfg, ctx *skill.SkillContext) []score.IEntity {
	if cfg == nil || ctx == nil {
		return nil
	}

	if cfg.Relation == conf.TargetRelation_Self {
		return []score.IEntity{m.owner}
	}

	sc := m.owner.GetScene()
	if sc == nil {
		return nil
	}

	switch cfg.Mode {
	case conf.TargetMode_Unit:
		return m.selectUnit(sc, ctx, cfg)
	case conf.TargetMode_NoTarget:
		return m.selectNoTarget(sc, ctx, cfg)
	case conf.TargetMode_Point:
		return m.selectPoint(sc, ctx, cfg)
	default:
		return nil
	}
}

func (m *SkillManager) selectUnit(sc score.IScene, ctx *skill.SkillContext, cfg *conf.TargetCfg) []score.IEntity {
	if ctx.Req == nil {
		return nil
	}

	targetId := uid.Uid(ctx.Req.LockTarget)
	if !targetId.IsValid() {
		return nil
	}

	entity, ok := sc.GetEntity(targetId)
	if !ok {
		return nil
	}

	return []score.IEntity{entity}
}

func (m *SkillManager) selectNoTarget(sc score.IScene, ctx *skill.SkillContext, cfg *conf.TargetCfg) []score.IEntity {
	if cfg.Shape == conf.ShapeType_Single {
		return []score.IEntity{m.owner}
	}

	if cfg.Shape == conf.ShapeType_Circle {
		pos := m.owner.GetPos()
		if pos == nil {
			return nil
		}

		newCtx := *ctx
		if newCtx.Req == nil {
			newCtx.Req = &pb.ReqCastSkill{}
		}
		newCtx.Req.Pos = &matrix.Vector3D{X: pos.X, Y: pos.Y, Z: pos.Z}

		return m.selectPoint(sc, &newCtx, cfg)
	}

	return nil
}

func (m *SkillManager) selectPoint(sc score.IScene, ctx *skill.SkillContext, cfg *conf.TargetCfg) []score.IEntity {
	if cfg.Shape != conf.ShapeType_Circle {
		return nil
	}

	r := float64(cfg.Radius)
	if r <= 0 {
		return nil
	}

	if ctx.Req == nil || ctx.Req.Pos == nil {
		return nil
	}

	center := *ctx.Req.Pos
	r2 := r * r
	result := make([]score.IEntity, 0)

	sc.ForEach(func(e score.IEntity) {
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
			result = append(result, e)
		}
	})

	return result
}
