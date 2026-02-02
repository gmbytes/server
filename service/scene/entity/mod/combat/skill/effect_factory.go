package skill

import (
	"server/data/conf"
)

func CreateEffect(cfg conf.EffectCfg) skillEffect {
	switch cfg.Type {
	case conf.EffectType_Damage:
		return NewDamageEffect(cfg)
	case conf.EffectType_Heal:
		return NewHealEffect(cfg)
	case conf.EffectType_ApplyAura:
		return NewAuraEffect(cfg)
	case conf.EffectType_Dispel:
		return NewDispelEffect(cfg)
	case conf.EffectType_Steal:
		return NewStealEffect(cfg)
	case conf.EffectType_Move:
		return NewMoveEffect(cfg)
	case conf.EffectType_Interrupt:
		return NewInterruptEffect(cfg)
	case conf.EffectType_Summon:
		return NewSummonEffect(cfg)
	case conf.EffectType_Threat:
		return NewThreatEffect(cfg)
	case conf.EffectType_SpawnArea:
		return NewSpawnAreaEffect(cfg)
	default:
		return nil
	}
}
