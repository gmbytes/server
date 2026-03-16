# Combat System - 战斗系统架构文档

## 概述

基于 `SkillContext` 和 `skillEffect` 接口的战斗系统，实现了完整的技能释放流程和效果执行机制。

## 核心组件

### 1. CombatManager - 战斗管理器

**职责**:
- 统一管理战斗相关的所有子系统
- 协调技能管理器 (SkillManager)
- 执行技能效果 (ExecuteEffect)
- 提供战斗计算接口 (伤害/治疗计算)
- 管理生命值状态

**关键方法**:
```go
func (m *CombatManager) Init(owner score.IEntity, initData data.EntityInitData)
func (m *CombatManager) Update(duration int64)
func (m *CombatManager) ExecuteEffect(eff conf.EffectCfg, ctx *skill.SkillContext, caster score.IEntity, targets []score.IEntity)
func (m *CombatManager) CalculateDamage(attacker, target score.IEntity, baseDamage int64) int64
func (m *CombatManager) CalculateHeal(healer, target score.IEntity, baseHeal int64) int64
func (m *CombatManager) ApplyDamage(target score.IEntity, damage int64)
func (m *CombatManager) ApplyHeal(target score.IEntity, heal int64)
```

### 2. SkillManager - 技能管理器

**职责**:
- 管理技能实例的生命周期
- 处理技能施放请求
- 选择技能目标
- 调度技能效果执行

**关键方法**:
```go
func (m *SkillManager) AddSkill(cfg *conf.CSkill)
func (m *SkillManager) Cast(skillId int64, req *pb.ReqCastSkill) bool
func (m *SkillManager) Cancel(skillId int64)
func (m *SkillManager) Update(deltaMs int64)
```

**目标选择**:
- `selectUnit` - 单体目标选择
- `selectNoTarget` - 无目标/自身目标
- `selectPoint` - 点选范围目标（圆形范围）

### 3. Skill - 技能运行时

**职责**:
- 管理技能状态 (空闲/吟唱/引导)
- 处理 CD/GCD
- 调度效果执行时机
- 支持多段结算

**技能阶段 (Stage)**:
- `Stage_CastStart` - 开始施法
- `Stage_CastFinish` - 吟唱结束/释放成功
- `Stage_Channel` - 引导阶段 Tick
- `Stage_Hit` - 命中阶段
- `Stage_Cancel` - 取消/被打断

**关键方法**:
```go
func NewSkill(cfg *conf.CSkill) *Skill
func (s *Skill) CanCast(now int64) bool
func (s *Skill) StartCast(now int64, ctx *SkillContext) bool
func (s *Skill) Cancel(now int64)
func (s *Skill) TriggerHit(now int64, ctx *SkillContext)
func (s *Skill) Update(now int64, exec func(Stage, conf.EffectCfg, *SkillContext))
```

### 4. SkillContext - 技能上下文

**职责**:
- 携带技能执行所需的完整上下文信息
- 提供唯一 ID 用于效果追踪
- 管理技能执行状态

**字段**:
```go
type SkillContext struct {
    id            uid.Uid          // 唯一ID，方便清理ctx
    destructionMs int64            // 对象销毁时间
    isReset       bool
    
    Zone       score.IZone        // 当前区域
    Owner      score.IEntity      // 技能拥有者
    Req        *pb.ReqCastSkill   // 技能请求
    SkillLevel int64              // 技能等级
    IsFinished bool               // 技能已结束
}
```

**工厂方法**:
```go
func NewSkillContext(owner score.IEntity, req *pb.ReqCastSkill, skillLevel int64) *SkillContext
```

### 5. skillEffect - 效果接口

所有技能效果都实现此接口:
```go
type skillEffect interface {
    Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity)
    Update(ctx *SkillContext, delta time.Duration)
    End(ctx *SkillContext)
    Revert(ctx *SkillContext)
}
```

**已实现的效果类型**:
- `DamageEffect` - 伤害
- `HealEffect` - 治疗
- `AuraEffect` - 光环/Buff
- `DispelEffect` - 驱散
- `StealEffect` - 偷取
- `MoveEffect` - 位移
- `InterruptEffect` - 打断
- `SummonEffect` - 召唤
- `ThreatEffect` - 仇恨
- `SpawnAreaEffect` - 生成区域

## 技能释放流程

```
1. 客户端请求施放技能
   ↓
2. SkillManager.Cast(skillId, req)
   - 创建 SkillContext
   - 调用 Skill.StartCast()
   ↓
3. Skill.StartCast(now, ctx)
   - 检查 CanCast (CD/GCD/状态)
   - 设置 GCD/CD
   - 调度 OnCastStart 效果
   - 判断是否需要吟唱
     - 有吟唱: 进入 Casting 状态
     - 瞬发: 直接调用 finishCast()
   ↓
4. Skill.Update(now, exec) [每帧调用]
   - 检查吟唱是否结束 → finishCast()
   - 检查引导是否结束
   - 执行 Pending 队列中到期的效果
   ↓
5. finishCast()
   - 调度 OnCastFinish 效果
   - 如果 HitOnCastFinish=true，调度 OnHit 效果
   - 如果有引导，进入 Channeling 状态并调度 OnChannelTick
   ↓
6. 效果执行回调 execEffect(stage, eff, ctx)
   - 选择目标配置 (selectTargetCfg)
   - 选择目标实体 (selectTargets)
   - 调用 CombatManager.ExecuteEffect()
   ↓
7. CombatManager.ExecuteEffect()
   - 通过工厂创建效果实例 (CreateEffect)
   - 调用 effect.Begin(ctx, caster, targets)
   ↓
8. 具体效果执行
   - DamageEffect: 计算并应用伤害
   - HealEffect: 计算并应用治疗
   - AuraEffect: 施加 Buff/Debuff
   - 其他效果...
```

## 目标选择机制

### 目标关系 (TargetRelation)
- `Self` - 自身
- `Ally` - 友方
- `Enemy` - 敌方

### 目标模式 (TargetMode)
- `Unit` - 单体目标（锁定目标）
- `Point` - 点选位置
- `NoTarget` - 无目标（通常作用于自身或自身周围）

### 范围形状 (ShapeType)
- `Single` - 单体
- `Circle` - 圆形范围
- `Cone` - 扇形
- `Rect` - 矩形
- `Ring` - 环形

## 效果调度机制

### ScheduledEffect - 延迟执行
```go
type ScheduledEffect struct {
    At     int64           // 执行时间戳
    Stage  Stage           // 执行阶段
    Effect conf.EffectCfg  // 效果配置
}
```

### 多段结算
通过 `EffectCfg.Times` 和 `EffectCfg.IntervalMs` 实现:
- `Times` - 执行次数
- `IntervalMs` - 间隔时间（毫秒）

示例：火球术每秒造成伤害，持续5秒
```go
EffectCfg{
    Type: EffectType_Damage,
    Times: 5,
    IntervalMs: 1000,
    P1: 100, // 基础伤害
}
```

## 使用示例

### 初始化战斗系统
```go
combatMgr := &CombatManager{}
combatMgr.Init(entity, initData)
```

### 添加技能
```go
skillCfg := &conf.CSkill{
    Cid: 1001,
    Name: "火球术",
    CastTimeMs: 1500,  // 1.5秒吟唱
    CooldownMs: 5000,  // 5秒CD
    GcdMs: 1500,       // 1.5秒GCD
    HitOnCastFinish: true,
    Target: conf.TargetCfg{
        Relation: conf.TargetRelation_Enemy,
        Mode: conf.TargetMode_Unit,
        Shape: conf.ShapeType_Single,
    },
    Effects: conf.SkillEffects{
        OnHit: []conf.EffectCfg{
            {
                Type: conf.EffectType_Damage,
                P1: 500, // 基础伤害
            },
        },
    },
}

skillMgr := combatMgr.GetSkillManager()
skillMgr.AddSkill(skillCfg)
```

### 施放技能
```go
req := &pb.ReqCastSkill{
    Cid: 1001,
    LockTarget: targetId.ToInt64(),
    Pos: &matrix.Vector3D{X: 100, Y: 200, Z: 0},
}

success := skillMgr.Cast(1001, req)
```

### 更新战斗系统
```go
// 每帧调用，deltaMs 为帧间隔（毫秒）
combatMgr.Update(50) // 50ms = 20fps
```

## 扩展指南

### 添加新的效果类型

1. 在 `conf.EffectType` 中添加新类型常量
2. 在 `skill/` 目录创建 `effect_xxx.go`
3. 实现 `skillEffect` 接口
4. 在 `skill/effect_factory.go` 的 `CreateEffect()` 中添加映射

### 自定义伤害计算
重写 `CombatManager.CalculateDamage()` 方法，加入属性计算、暴击、防御等逻辑。

### 添加 Buff 系统
创建 `BuffManager` 管理持续性效果，通过 `AuraEffect` 施加 Buff。

## 设计原则

1. **职责分离**: Skill 只负责阶段推进，不处理具体效果逻辑
2. **接口驱动**: 通过 skillEffect 接口实现效果的多态
3. **上下文传递**: SkillContext 携带完整的执行上下文
4. **配置驱动**: 技能行为由配置表 (CSkill) 定义
5. **时间驱动**: 基于时间戳的调度系统，支持延迟和多段结算

## 参考文档

详细的 Effect 和 Buff 设计参考见: `skill/doc/README.md`
