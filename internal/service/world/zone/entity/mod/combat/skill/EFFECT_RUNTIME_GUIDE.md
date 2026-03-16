# Effect 运行数据处理指南

## 核心概念

Effect 执行时的运行数据处理分为两种模式：

### 1. 瞬时 Effect（Instant Effect）
**特点**：一次性结算，不需要维护运行时状态

**处理方式**：
- 直接在 `Begin()` 中完成所有逻辑
- 不需要 `Update()` / `End()` / `Revert()`
- 不需要 EffectRuntime

**适用类型**：
- Damage（瞬时伤害）
- Heal（瞬时治疗）
- Move（位移）
- Interrupt（打断）
- Dispel（驱散）
- Steal（偷取）
- Threat（仇恨修改）
- Summon（召唤）

### 2. 持续性 Effect（Persistent Effect）
**特点**：需要持续一段时间，周期性执行逻辑

**处理方式**：
- `Begin()` - 初始化效果
- `Update()` - 每个 Tick 执行
- `End()` - 正常结束时清理
- `Revert()` - 被打断/驱散时回滚
- 需要 EffectRuntime 管理状态

**适用类型**：
- ApplyAura（Buff/Debuff）
- SpawnArea（区域效果）

## 运行数据管理架构

```
EffectRuntime (运行时数据)
    ├── Id: 效果实例唯一ID
    ├── Effect: 效果实例
    ├── Ctx: SkillContext
    ├── Caster/Targets: 施法者和目标
    ├── StartMs/EndMs: 开始/结束时间
    ├── TickIntervalMs: Tick间隔
    ├── TickCount/MaxTicks: Tick计数
    └── IsActive: 激活状态

EffectManager (效果管理器)
    ├── runningEffects: map[uid.Uid]*EffectRuntime
    ├── AddEffect(): 添加持续效果
    ├── RemoveEffect(): 正常移除
    ├── CancelEffect(): 取消并回滚
    └── Update(): 每帧更新所有效果
```

## 实现示例

### 示例1：瞬时伤害（DamageEffect）

```go
type DamageEffect struct {
    cfg conf.EffectCfg
}

func (e *DamageEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
    // 从配置读取基础伤害
    baseDamage := e.cfg.P1
    
    // 获取战斗管理器
    combatMgr := getCombatManager(causer)
    if combatMgr == nil {
        return
    }
    
    // 对每个目标造成伤害
    for _, target := range targets {
        // 计算最终伤害
        damage := combatMgr.CalculateDamage(causer, target, baseDamage)
        
        // 应用伤害
        targetCombat := getCombatManager(target)
        if targetCombat != nil {
            targetCombat.ApplyDamage(target, damage)
        }
    }
}

// 瞬时效果不需要实现 Update/End/Revert
func (e *DamageEffect) Update(ctx *SkillContext, delta time.Duration) {}
func (e *DamageEffect) End(ctx *SkillContext) {}
func (e *DamageEffect) Revert(ctx *SkillContext) {}
```

### 示例2：持续伤害 DoT（通过 AuraEffect 实现）

```go
type AuraEffect struct {
    cfg conf.EffectCfg
    
    // 运行时数据（由 EffectRuntime 管理）
    // 这里只存储配置，实际运行数据在 EffectRuntime 中
}

func (e *AuraEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
    // 初始化 Buff/Debuff
    // RefId 指向 Buff 配置ID
    buffId := e.cfg.RefId
    
    for _, target := range targets {
        // 施加 Buff 到目标
        // 这里可以创建 Buff 实例并添加到目标的 BuffManager
        applyBuff(target, buffId, ctx)
    }
}

func (e *AuraEffect) Update(ctx *SkillContext, delta time.Duration) {
    // 每个 Tick 执行
    // 例如：DoT 每秒造成伤害
    
    // 从配置读取每跳伤害
    tickDamage := e.cfg.P1
    
    // 这里可以访问 EffectRuntime 中的 Targets
    // 对每个目标造成 DoT 伤害
    for _, target := range ctx.Owner.GetZone().GetEntities() {
        // 检查目标是否有此 Buff
        if hasBuff(target, e.cfg.RefId) {
            applyDotDamage(target, tickDamage)
        }
    }
}

func (e *AuraEffect) End(ctx *SkillContext) {
    // 正常结束时清理
    // 移除 Buff
    buffId := e.cfg.RefId
    
    // 从所有目标移除 Buff
    // 这里需要从 EffectRuntime 获取目标列表
}

func (e *AuraEffect) Revert(ctx *SkillContext) {
    // 被驱散/打断时回滚
    // 移除 Buff 并可能回滚已造成的效果
    e.End(ctx)
}
```

### 示例3：区域持续效果（SpawnAreaEffect）

```go
type SpawnAreaEffect struct {
    cfg conf.EffectCfg
}

func (e *SpawnAreaEffect) Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity) {
    // 创建区域实体
    // P1: 区域半径
    // P2: 持续时间
    // RefId: 区域配置ID
    
    areaId := e.cfg.RefId
    radius := float32(e.cfg.P1)
    
    // 在场景中创建区域实体
    createAreaEntity(ctx.Zone, ctx.Req.Pos, radius, areaId)
}

func (e *SpawnAreaEffect) Update(ctx *SkillContext, delta time.Duration) {
    // 每个 Tick 检测区域内的单位
    // 对区域内的单位施加效果
    
    entitiesInArea := findEntitiesInArea(ctx.Zone, ctx.Req.Pos, e.cfg.P1)
    
    for _, entity := range entitiesInArea {
        // 施加区域效果（例如每秒造成伤害）
        applyAreaDamage(entity, e.cfg.P3)
    }
}

func (e *SpawnAreaEffect) End(ctx *SkillContext) {
    // 移除区域实体
    removeAreaEntity(ctx.Zone, e.cfg.RefId)
}

func (e *SpawnAreaEffect) Revert(ctx *SkillContext) {
    // 立即移除区域
    e.End(ctx)
}
```

## 配置参数说明

### EffectCfg 参数用途

```go
type EffectCfg struct {
    Type       EffectType  // 效果类型
    Times      int32       // 执行次数（持续效果的 Tick 次数）
    IntervalMs int32       // 间隔时间（持续效果的 Tick 间隔）
    RefId      int64       // 引用ID（BuffId/AreaId/SummonId等）
    
    // 通用参数（具体含义由 EffectType 决定）
    P1 int64  // 伤害：基础伤害值 | Buff：BuffId | 区域：半径
    P2 int64  // 持续时间（毫秒）| 治疗：基础治疗值
    P3 int64  // 扩展参数1
    P4 int64  // 扩展参数2
    
    Args []int64  // 额外参数数组
}
```

### 配置示例

#### 瞬时伤害
```go
EffectCfg{
    Type: EffectType_Damage,
    P1: 500,  // 基础伤害
}
```

#### DoT（持续伤害）
```go
EffectCfg{
    Type: EffectType_ApplyAura,
    RefId: 1001,        // Buff ID
    Times: 5,           // 5次Tick
    IntervalMs: 1000,   // 每秒一次
    P1: 100,            // 每跳伤害
    P2: 5000,           // 持续5秒
}
```

#### 区域持续伤害
```go
EffectCfg{
    Type: EffectType_SpawnArea,
    RefId: 2001,        // 区域配置ID
    Times: 10,          // 10次Tick
    IntervalMs: 500,    // 每0.5秒一次
    P1: 300,            // 半径300
    P2: 5000,           // 持续5秒
    P3: 50,             // 每跳伤害
}
```

## 运行数据访问

### 在 Effect 中访问运行数据

由于 Effect 接口只接收 `SkillContext`，如果需要访问 `EffectRuntime` 的数据，有两种方式：

#### 方式1：通过 SkillContext 扩展
```go
// 在 SkillContext 中添加 Runtime 引用
type SkillContext struct {
    // ... 现有字段
    Runtime *EffectRuntime  // 关联的运行时数据（可选）
}
```

#### 方式2：在 Effect 内部维护状态
```go
type AuraEffect struct {
    cfg conf.EffectCfg
    
    // 内部状态（仅用于 Begin 到 End 之间的临时数据）
    appliedTargets []score.IEntity
    buffInstances  map[uid.Uid]interface{}
}
```

**推荐使用方式1**，因为 EffectRuntime 已经包含了所有需要的运行时信息。

## 最佳实践

### 1. 瞬时效果
- ✅ 在 `Begin()` 中完成所有逻辑
- ✅ 保持无状态设计
- ✅ 使用配置参数驱动

### 2. 持续效果
- ✅ 使用 EffectRuntime 管理状态
- ✅ 在 `Begin()` 中初始化
- ✅ 在 `Update()` 中执行周期逻辑
- ✅ 在 `End()` 中清理资源
- ✅ 在 `Revert()` 中回滚状态

### 3. 数据隔离
- ✅ 配置数据（EffectCfg）只读
- ✅ 运行时数据（EffectRuntime）由管理器维护
- ✅ 临时状态在 Effect 实例内部

### 4. 生命周期管理
- ✅ 由 EffectManager 统一管理
- ✅ 自动检测过期并清理
- ✅ 支持主动移除和取消

## 调试技巧

### 查看运行中的效果
```go
effectMgr := combatMgr.GetEffectManager()

// 获取目标身上的所有效果
effects := effectMgr.GetEffectsByTarget(targetId)
for _, runtime := range effects {
    fmt.Printf("Effect ID: %v, TickCount: %d/%d, Active: %v\n",
        runtime.Id, runtime.TickCount, runtime.MaxTicks, runtime.IsActive)
}
```

### 手动移除效果
```go
// 正常移除
effectMgr.RemoveEffect(effectId)

// 取消并回滚
effectMgr.CancelEffect(effectId)
```

## 总结

**运行数据处理的核心原则**：
1. **瞬时效果**：无状态，直接执行
2. **持续效果**：有状态，由 EffectRuntime + EffectManager 管理
3. **数据分离**：配置、运行时、临时状态三者分离
4. **生命周期**：Begin → Update(循环) → End/Revert
