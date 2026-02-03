package skill

import (
	"server/lib/uid"
	"server/pb"
	"server/service/scene/score"
	"time"
)

// GlobalDataKey 全局数据键类型（避免字符串拼写错误）
type GlobalDataKey string

// 预定义的常用全局数据键
const (
	// 标记相关
	GlobalKey_MarkStacks  GlobalDataKey = "mark_stacks"  // 标记层数
	GlobalKey_MarkTargets GlobalDataKey = "mark_targets" // 被标记的目标

	// 连击相关
	GlobalKey_ComboCount   GlobalDataKey = "combo_count"   // 连击次数
	GlobalKey_ComboCrit    GlobalDataKey = "combo_crit"    // 连击是否暴击
	GlobalKey_ComboTargets GlobalDataKey = "combo_targets" // 连击目标

	// 链式相关
	GlobalKey_ChainHitTargets GlobalDataKey = "chain_hit_targets" // 链式已命中目标
	GlobalKey_ChainCount      GlobalDataKey = "chain_count"       // 链式弹跳次数
	GlobalKey_ChainLastTarget GlobalDataKey = "chain_last_target" // 链式最后目标

	// 触发相关
	GlobalKey_HasTriggered GlobalDataKey = "has_triggered" // 是否已触发
	GlobalKey_TriggerCount GlobalDataKey = "trigger_count" // 触发次数

	// 自定义（用户可以定义自己的键）
	// 使用方式：GlobalDataKey("my_custom_key")
)

// EffectResult 存储单个 Effect 的执行结果
// 使用类型化字段替代 map[string]interface{}，避免字符串 key 拼写错误
type EffectResult struct {
	Seq int32 // Effect 全局序列号

	// 常用数据字段（覆盖大部分场景）
	Damage    int64           // 造成的伤害
	Heal      int64           // 治疗量
	IsCrit    bool            // 是否暴击
	Targets   []score.IEntity // 命中的目标
	HitCount  int32           // 命中次数
	KilledAny bool            // 是否击杀了目标

	// 扩展字段（特殊情况使用，使用 GlobalDataKey 避免拼写错误）
	ExtraInt64  map[GlobalDataKey]int64
	ExtraBool   map[GlobalDataKey]bool
	ExtraEntity map[GlobalDataKey][]score.IEntity
}

// NewEffectResult 创建 Effect 结果
func NewEffectResult(seq int32) *EffectResult {
	return &EffectResult{
		Seq:         seq,
		ExtraInt64:  make(map[GlobalDataKey]int64),
		ExtraBool:   make(map[GlobalDataKey]bool),
		ExtraEntity: make(map[GlobalDataKey][]score.IEntity),
	}
}

type SkillContext struct {
	id            uid.Uid // 给ctx 一个唯一id 方便清理ctx
	destructionMs int64   // 对象销毁时间
	isReset       bool

	Scene score.IScene  // 当前场景
	Owner score.IEntity // 技能拥有者

	Req        *pb.ReqCastSkill // 技能请求
	SkillLevel int64            // 技能等级
	IsFinished bool             // 技能已结束

	// CurrentEffectSeq 当前执行的 Effect 全局序列号
	// 相比 CurrentEffectIndex，Seq 能唯一标识整个技能中的每个 Effect 实例
	CurrentEffectSeq int32

	// 类型化的数据存储（按 Seq 索引）
	effectResults []*EffectResult // 每个 Effect 的执行结果

	// 全局共享数据（用于跨 Effect 的累计统计）
	// 使用具体类型字段替代 map[string]interface{}
	TotalDamage int64 // 累计总伤害
	TotalHeal   int64 // 累计总治疗
	TotalHits   int32 // 累计命中次数
	KillCount   int32 // 击杀数

	// 扩展的全局数据（如果上面的字段不够用）
	// 使用 GlobalDataKey 类型作为 key，避免字符串拼写错误
	globalInt64  map[GlobalDataKey]int64
	globalBool   map[GlobalDataKey]bool
	globalEntity map[GlobalDataKey][]score.IEntity
}

func NewSkillContext(owner score.IEntity, req *pb.ReqCastSkill, skillLevel int64) *SkillContext {
	ctx := &SkillContext{
		id:            uid.Gen(),
		Owner:         owner,
		Req:           req,
		SkillLevel:    skillLevel,
		IsFinished:    false,
		effectResults: make([]*EffectResult, 0, 16), // 预分配一些空间
		globalInt64:   make(map[GlobalDataKey]int64),
		globalBool:    make(map[GlobalDataKey]bool),
		globalEntity:  make(map[GlobalDataKey][]score.IEntity),
	}
	if owner != nil {
		ctx.Scene = owner.GetScene()
	}
	return ctx
}

func (c *SkillContext) Finish() {
	c.IsFinished = true
}

func (c *SkillContext) GetId() uid.Uid {
	return c.id
}

// ========== Effect 结果相关 API ==========

// GetCurrentResult 获取当前 Effect 的结果（自动创建）
func (c *SkillContext) GetCurrentResult() *EffectResult {
	// 确保 effectResults 有足够的空间
	for len(c.effectResults) <= int(c.CurrentEffectSeq) {
		c.effectResults = append(c.effectResults, NewEffectResult(int32(len(c.effectResults))))
	}
	return c.effectResults[c.CurrentEffectSeq]
}

// GetPrevResult 获取上一个 Effect 的结果
func (c *SkillContext) GetPrevResult() *EffectResult {
	if c.CurrentEffectSeq <= 0 {
		return nil
	}
	prevSeq := c.CurrentEffectSeq - 1
	if int(prevSeq) >= len(c.effectResults) {
		return nil
	}
	return c.effectResults[prevSeq]
}

// GetResultBySeq 获取指定序列号的 Effect 结果
func (c *SkillContext) GetResultBySeq(seq int32) *EffectResult {
	if seq < 0 || int(seq) >= len(c.effectResults) {
		return nil
	}
	return c.effectResults[seq]
}

// GetAllResults 获取所有 Effect 结果
func (c *SkillContext) GetAllResults() []*EffectResult {
	return c.effectResults
}

// ========== 全局数据 API（扩展字段） ==========

// SetGlobalInt64 设置全局 int64 数据
func (c *SkillContext) SetGlobalInt64(key GlobalDataKey, value int64) {
	c.globalInt64[key] = value
}

// GetGlobalInt64 获取全局 int64 数据
func (c *SkillContext) GetGlobalInt64(key GlobalDataKey) (int64, bool) {
	val, ok := c.globalInt64[key]
	return val, ok
}

// IncrementGlobalInt64 递增全局 int64 数据
func (c *SkillContext) IncrementGlobalInt64(key GlobalDataKey, delta int64) int64 {
	val := c.globalInt64[key]
	val += delta
	c.globalInt64[key] = val
	return val
}

// SetGlobalBool 设置全局 bool 数据
func (c *SkillContext) SetGlobalBool(key GlobalDataKey, value bool) {
	c.globalBool[key] = value
}

// GetGlobalBool 获取全局 bool 数据
func (c *SkillContext) GetGlobalBool(key GlobalDataKey) (bool, bool) {
	val, ok := c.globalBool[key]
	return val, ok
}

// SetGlobalEntities 设置全局实体列表
func (c *SkillContext) SetGlobalEntities(key GlobalDataKey, entities []score.IEntity) {
	c.globalEntity[key] = entities
}

// GetGlobalEntities 获取全局实体列表
func (c *SkillContext) GetGlobalEntities(key GlobalDataKey) ([]score.IEntity, bool) {
	val, ok := c.globalEntity[key]
	return val, ok
}

type skillEffect interface {
	Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity)
	Update(ctx *SkillContext, delta time.Duration)
	End(ctx *SkillContext)
	Revert(ctx *SkillContext)
}
