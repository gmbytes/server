package combat

import (
	"server/data"
	"server/data/conf"
	"server/data/enum"
	"server/service/world/zone/entity/mod/combat/skill"
	"server/service/world/zone/izone"
)

var _ izone.IModule = (*CombatManager)(nil)

type CombatManager struct {
	owner izone.IEntity
	attrs *data.Attrs

	skillMgr  *SkillManager
	effectMgr *EffectManager

	hp    int64
	maxHp int64
}

func (m *CombatManager) Init(owner izone.IEntity, initData data.EntityInitData) {
	m.owner = owner
	m.attrs = initData.Attrs
	m.skillMgr = newSkillManager(m)
	m.effectMgr = newEffectManager(m)

	if m.attrs != nil {
		m.maxHp = m.attrs.GetValue(enum.AttrType_MaxHp)
		m.hp = m.maxHp
	}
}

func (m *CombatManager) Update(duration int64) {
	m.skillMgr.Update(duration)
	m.effectMgr.Update(duration)
}

func (m *CombatManager) ExecuteEffect(eff conf.EffectCfg, ctx *skill.SkillContext, caster izone.IEntity, targets []izone.IEntity) {
	if len(targets) == 0 {
		return
	}

	effect := skill.CreateEffect(eff)
	if effect == nil {
		return
	}

	// 判断是否为持续性效果
	if m.isInstantEffect(eff.Type) {
		// 瞬时效果：直接执行Begin，不需要运行时数据
		effect.Begin(ctx, caster, targets)
	} else {
		// 持续性效果：创建运行时数据并加入管理器
		runtime := skill.NewEffectRuntime(effect, ctx, caster, targets)

		// 设置持续时间和Tick参数
		if eff.P2 > 0 {
			runtime.EndMs = runtime.StartMs + eff.P2 // P2作为持续时间（毫秒）
		}
		if eff.IntervalMs > 0 {
			runtime.TickIntervalMs = int64(eff.IntervalMs)
		}
		if eff.Times > 0 {
			runtime.MaxTicks = eff.Times
		}

		// 执行Begin初始化
		effect.Begin(ctx, caster, targets)

		// 加入效果管理器
		m.effectMgr.AddEffect(runtime)
	}
}

// isInstantEffect 判断是否为瞬时效果
func (m *CombatManager) isInstantEffect(effectType conf.EffectType) bool {
	switch effectType {
	case conf.EffectType_Damage: // 瞬时伤害
		return true
	case conf.EffectType_Heal: // 瞬时治疗
		return true
	case conf.EffectType_Move: // 瞬时位移
		return true
	case conf.EffectType_Interrupt: // 瞬时打断
		return true
	case conf.EffectType_Dispel: // 瞬时驱散
		return true
	case conf.EffectType_Steal: // 瞬时偷取
		return true
	case conf.EffectType_Threat: // 瞬时仇恨修改
		return true
	case conf.EffectType_Summon: // 召唤（瞬时创建）
		return true
	case conf.EffectType_ApplyAura: // Buff/Debuff（持续）
		return false
	case conf.EffectType_SpawnArea: // 区域效果（持续）
		return false
	default:
		return true
	}
}

func (m *CombatManager) CalculateDamage(attacker izone.IEntity, target izone.IEntity, baseDamage int64) int64 {
	if attacker == nil || target == nil {
		return 0
	}

	damage := baseDamage

	return damage
}

func (m *CombatManager) CalculateHeal(healer izone.IEntity, target izone.IEntity, baseHeal int64) int64 {
	if healer == nil || target == nil {
		return 0
	}

	heal := baseHeal

	return heal
}

func (m *CombatManager) ApplyDamage(target izone.IEntity, damage int64) {
	if target == nil || damage <= 0 {
		return
	}

	if target.GetId() == m.owner.GetId() {
		m.hp -= damage
		if m.hp < 0 {
			m.hp = 0
		}
	}
}

func (m *CombatManager) ApplyHeal(target izone.IEntity, heal int64) {
	if target == nil || heal <= 0 {
		return
	}

	if target.GetId() == m.owner.GetId() {
		m.hp += heal
		if m.hp > m.maxHp {
			m.hp = m.maxHp
		}
	}
}

func (m *CombatManager) GetHp() int64 {
	return m.hp
}

func (m *CombatManager) GetMaxHp() int64 {
	return m.maxHp
}

func (m *CombatManager) GetSkillManager() *SkillManager {
	return m.skillMgr
}

func (m *CombatManager) GetEffectManager() *EffectManager {
	return m.effectMgr
}
