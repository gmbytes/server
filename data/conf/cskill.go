package conf

// CSkill 为技能配置（只读数据），由配置/配表加载。
// 该结构只描述“技能是什么”，不包含运行时状态（CD/充能/吟唱进度等）。
type CSkill struct {
	Cid int64 // 技能配置ID

	Name string // 技能名称

	CastTimeMs         int32       // 吟唱时间（毫秒），0 表示瞬发
	ChannelTimeMs      int32       // 引导总时长（毫秒），0 表示非引导
	ChannelTickMs      int32       // 引导每跳间隔（毫秒），0 表示不启用阶段级 Tick（完全由 EffectCfg.Times/IntervalMs 控制）
	ChannelTickDelayMs int32       // 引导首跳延迟（毫秒），0 表示进入引导后立刻触发第一跳
	GcdMs              int32       // 公共CD（毫秒）
	CooldownMs         int32       // 技能冷却（毫秒）
	GcdStartAt         TimingPoint // GCD 起算时机（默认 CastStart）
	CooldownStartAt    TimingPoint // CD 起算时机（默认 CastStart）

	HitOnCastFinish bool  // 是否在 CastFinish 后自动触发一次 OnHit（常用于“瞬发即命中”的技能）
	HitDelayMs      int32 // HitOnCastFinish 为 true 时生效：CastFinish 到 Hit 的延迟（毫秒）

	Charges    int32 // 充能数量（0/1 表示无充能机制）
	RechargeMs int32 // 充能恢复时间（毫秒）

	CostMp int64 // 消耗MP

	RangeMin float32 // 最小施法距离
	RangeMax float32 // 最大施法距离

	Target TargetCfg // 目标选择/范围形状
	// Selectors 用于“不同阶段不同筛选器”的情况。
	// 若某阶段为 nil，则使用默认的 Target。
	Selectors SkillSelectors

	Effects SkillEffects // 分阶段效果列表（Effect）
}

// TimingPoint 表示“某个规则从技能哪个阶段开始起算”。
type TimingPoint int32

const (
	TimingPoint_Invalid    TimingPoint = 0
	TimingPoint_CastStart  TimingPoint = 1
	TimingPoint_CastFinish TimingPoint = 2
)

// SkillSelectors 为按阶段覆盖的目标筛选配置。
// 注意：这是“选目标规则”，不是效果本身；效果仍由 SkillEffects 驱动。
type SkillSelectors struct {
	OnCastStart   *TargetCfg
	OnCastFinish  *TargetCfg
	OnChannelTick *TargetCfg
	OnHit         *TargetCfg
	OnCancel      *TargetCfg
}

// TargetRelation 表示技能的目标关系（对谁生效）。
type TargetRelation int32

const (
	TargetRelation_Invalid TargetRelation = 0
	TargetRelation_Self    TargetRelation = 1
	TargetRelation_Ally    TargetRelation = 2
	TargetRelation_Enemy   TargetRelation = 3
)

// TargetMode 表示技能的目标模式（如何选目标）。
type TargetMode int32

const (
	TargetMode_Invalid  TargetMode = 0
	TargetMode_Unit     TargetMode = 1
	TargetMode_Point    TargetMode = 2
	TargetMode_NoTarget TargetMode = 3
)

// ShapeType 表示技能作用区域形状。
type ShapeType int32

const (
	ShapeType_Invalid ShapeType = 0
	ShapeType_Single  ShapeType = 1
	ShapeType_Circle  ShapeType = 2
	ShapeType_Cone    ShapeType = 3
	ShapeType_Rect    ShapeType = 4
	ShapeType_Ring    ShapeType = 5
)

// TargetCfg 描述技能目标与范围参数。
type TargetCfg struct {
	Relation TargetRelation // 目标关系（自/友/敌）
	Mode     TargetMode     // 目标模式（单位/点/无目标）
	Shape    ShapeType      // 区域形状

	Radius float32 // 圆/环半径
	Angle  float32 // 扇形角度
	Width  float32 // 矩形宽
	Length float32 // 矩形长/扇形长度
}

// EffectType 表示瞬时结算的效果类型（Effect）。
// 持续类效果应通过 EffectType_ApplyAura 施加 Buff/Debuff，由 Buff 系统维护生命周期。
type EffectType int32

const (
	EffectType_Invalid   EffectType = 0
	EffectType_Damage    EffectType = 1
	EffectType_Heal      EffectType = 2
	EffectType_ApplyAura EffectType = 3
	EffectType_Dispel    EffectType = 4
	EffectType_Steal     EffectType = 5
	EffectType_Move      EffectType = 6
	EffectType_Interrupt EffectType = 7
	EffectType_Summon    EffectType = 8
	EffectType_Threat    EffectType = 9
	EffectType_SpawnArea EffectType = 10
)

// EffectCfg 为单个效果配置。
// Times/IntervalMs 用于多段结算：同一个 Effect 可重复执行多次，间隔 IntervalMs。
type EffectCfg struct {
	Type EffectType // 效果类型

	Times      int32 // 执行次数（<=1 视为 1）
	IntervalMs int32 // 多段间隔（毫秒）

	RefId int64 // 引用ID（例如 BuffId、召唤物Id、区域Id 等，由具体 EffectType 解释）

	P1 int64 // 通用参数1（由具体 EffectType 解释）
	P2 int64 // 通用参数2
	P3 int64 // 通用参数3
	P4 int64 // 通用参数4

	Args []int64 // 扩展参数（可用于倍率、范围、标记等）
}

// SkillEffects 将效果按“技能阶段/时机”拆分。
// 运行时由 skill.Skill 在不同阶段触发对应列表。
type SkillEffects struct {
	OnCastStart   []EffectCfg // 开始施法/按下技能且校验通过
	OnCastFinish  []EffectCfg // 吟唱结束/瞬发释放成功
	OnChannelTick []EffectCfg // 引导期间的每跳
	OnHit         []EffectCfg // 命中时（弹道到达/范围生效）
	OnCancel      []EffectCfg // 取消/被打断
}
