package config

type AllConfig struct {
	Buffs []Buff `json:"buffs"`
	BuffEffects []BuffEffect `json:"buffEffects"`
	DamageFormulas []DamageFormula `json:"damageFormulas"`
	Selectors []Selector `json:"selectors"`
	Skills []Skill `json:"skills"`
	SkillEffects []SkillEffect `json:"skillEffects"`
}

type Buff struct {
	ID int `json:"ID"`
	Name string `json:"Name"`
	Description string `json:"Description"`
	Icon string `json:"Icon"`
	BuffType int `json:"BuffType"`
	DurationMs int `json:"DurationMs"`
	MaxStacks int `json:"MaxStacks"`
	StackRule int `json:"StackRule"`
	DispelType int `json:"DispelType"`
	CanDispel bool `json:"CanDispel"`
	CanSteal bool `json:"CanSteal"`
	Priority int `json:"Priority"`
	EffectIDs []int `json:"EffectIDs"`
	ImmunityMask int `json:"ImmunityMask"`
	Tags string `json:"Tags"`
}

type BuffEffect struct {
	ID int `json:"ID"`
	Name string `json:"Name"`
	EffectType int `json:"EffectType"`
	TriggerType int `json:"TriggerType"`
	TickIntervalMs int `json:"TickIntervalMs"`
	MaxTicks int `json:"MaxTicks"`
	EventType int `json:"EventType"`
	TriggerChance float64 `json:"TriggerChance"`
	CooldownMs int `json:"CooldownMs"`
	AttributeType int `json:"AttributeType"`
	ModType int `json:"ModType"`
	ModValue int `json:"ModValue"`
	DamageFormulaID int `json:"DamageFormulaID"`
	HealFormulaID int `json:"HealFormulaID"`
	ShieldAmount int `json:"ShieldAmount"`
	CCType int `json:"CCType"`
	MoveSpeedPct float64 `json:"MoveSpeedPct"`
	AttackSpeedPct float64 `json:"AttackSpeedPct"`
	CastSpeedPct float64 `json:"CastSpeedPct"`
	P1 int `json:"P1"`
	P2 int `json:"P2"`
}

type DamageFormula struct {
	ID int `json:"ID"`
	Name string `json:"Name"`
	DamageType int `json:"DamageType"`
	School int `json:"School"`
	BaseDamage int `json:"BaseDamage"`
	BaseDamagePerLevel int `json:"BaseDamagePerLevel"`
	APCoefficient float64 `json:"APCoefficient"`
	SPCoefficient float64 `json:"SPCoefficient"`
	TargetHPCoefficient float64 `json:"TargetHPCoefficient"`
	TargetMissingHPCoefficient float64 `json:"TargetMissingHPCoefficient"`
	CasterHPCoefficient float64 `json:"CasterHPCoefficient"`
	ExecuteThreshold float64 `json:"ExecuteThreshold"`
	ExecuteBonus float64 `json:"ExecuteBonus"`
	CanCrit bool `json:"CanCrit"`
	CritMultiplier float64 `json:"CritMultiplier"`
	CanDodge bool `json:"CanDodge"`
	CanBlock bool `json:"CanBlock"`
	CanParry bool `json:"CanParry"`
	IgnoreArmorPct float64 `json:"IgnoreArmorPct"`
	SplashRadius float64 `json:"SplashRadius"`
	SplashDamagePct float64 `json:"SplashDamagePct"`
	MinDamage int `json:"MinDamage"`
	MaxDamage int `json:"MaxDamage"`
}

type Selector struct {
	ID int `json:"ID"`
	Name string `json:"Name"`
	Mode int `json:"Mode"`
	Shape int `json:"Shape"`
	Radius float64 `json:"Radius"`
	Angle float64 `json:"Angle"`
	Width float64 `json:"Width"`
	Length float64 `json:"Length"`
	InnerRadius float64 `json:"InnerRadius"`
	Relation int `json:"Relation"`
	MinHP int `json:"MinHP"`
	MaxHP int `json:"MaxHP"`
	MinHPPct float64 `json:"MinHPPct"`
	MaxHPPct float64 `json:"MaxHPPct"`
	RequireBuffID int `json:"RequireBuffID"`
	ExcludeBuffID int `json:"ExcludeBuffID"`
	Sort int `json:"Sort"`
	MaxCount int `json:"MaxCount"`
	IncludeCaster bool `json:"IncludeCaster"`
	IncludeDead bool `json:"IncludeDead"`
}

type Skill struct {
	ID int `json:"ID"`
	Name string `json:"Name"`
	Description string `json:"Description"`
	Icon string `json:"Icon"`
	SkillType int `json:"SkillType"`
	TargetType int `json:"TargetType"`
	MaxLevel int `json:"MaxLevel"`
	CooldownMs int `json:"CooldownMs"`
	CooldownStartStage int `json:"CooldownStartStage"`
	GcdMs int `json:"GcdMs"`
	GcdStartStage int `json:"GcdStartStage"`
	CastTimeMs int `json:"CastTimeMs"`
	ChannelTimeMs int `json:"ChannelTimeMs"`
	ChannelTickMs int `json:"ChannelTickMs"`
	Range float64 `json:"Range"`
	ResourceType int `json:"ResourceType"`
	ResourceCost int `json:"ResourceCost"`
	TargetSelectorID int `json:"TargetSelectorID"`
	EffectIDs []int `json:"EffectIDs"`
	RequireBuffID int `json:"RequireBuffID"`
	ConsumeBuffID int `json:"ConsumeBuffID"`
	CanCastWhileMoving bool `json:"CanCastWhileMoving"`
	CanCastWhileStunned bool `json:"CanCastWhileStunned"`
	InterruptibleByDamage bool `json:"InterruptibleByDamage"`
	InterruptibleByCC bool `json:"InterruptibleByCC"`
	SchoolMask int `json:"SchoolMask"`
	Tags string `json:"Tags"`
}

type SkillEffect struct {
	ID int `json:"ID"`
	Name string `json:"Name"`
	EffectType int `json:"EffectType"`
	Stage int `json:"Stage"`
	DelayMs int `json:"DelayMs"`
	Times int `json:"Times"`
	IntervalMs int `json:"IntervalMs"`
	DamageFormulaID int `json:"DamageFormulaID"`
	HealFormulaID int `json:"HealFormulaID"`
	BuffID int `json:"BuffID"`
	BuffDurationMs int `json:"BuffDurationMs"`
	BuffStacks int `json:"BuffStacks"`
	DispelType int `json:"DispelType"`
	DispelCount int `json:"DispelCount"`
	MoveType int `json:"MoveType"`
	MoveDistance float64 `json:"MoveDistance"`
	SummonID int `json:"SummonID"`
	ThreatValue int `json:"ThreatValue"`
	AreaID int `json:"AreaID"`
	P1 int `json:"P1"`
	P2 int `json:"P2"`
	P3 int `json:"P3"`
	P4 int `json:"P4"`
}
