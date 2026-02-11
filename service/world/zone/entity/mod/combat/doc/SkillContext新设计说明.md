# SkillContext 新设计说明

## 设计问题回顾

### 旧设计的问题

1. **CurrentEffectIndex 不够精确**
   - 只能区分同一个 EffectCfg 配置的多次执行
   - 无法区分不同配置的相同类型 Effect
   - 无法区分不同阶段的 Effect

2. **map[string]interface{} 不够类型安全**
   - 字符串 key 容易拼写错误
   - 类型断言容易出错
   - 没有代码提示和编译检查

## 新设计方案

### 1. 使用全局序列号（Seq）

```go
type ScheduledEffect struct {
    At     int64
    Stage  Stage
    Effect conf.EffectCfg
    Index  int32 // 同一 EffectCfg 配置的第几次执行（0, 1, 2, ...）
    Seq    int32 // 全局序列号（整个技能所有 Effect 的执行顺序）
}
```

**Seq 的优势**：
- 唯一标识整个技能中的每个 Effect 实例
- 不受配置、阶段、类型的影响
- 严格按执行顺序递增

### 2. 类型化的 EffectResult

```go
type EffectResult struct {
    Seq int32 // Effect 序列号

    // 常用数据字段（覆盖大部分场景）
    Damage    int64           // 造成的伤害
    Heal      int64           // 治疗量
    IsCrit    bool            // 是否暴击
    Targets   []score.IEntity // 命中的目标
    HitCount  int32           // 命中次数
    KilledAny bool            // 是否击杀了目标

    // 扩展字段（特殊情况使用，但仍然类型安全）
    ExtraInt64  map[string]int64
    ExtraBool   map[string]bool
    ExtraEntity map[string][]score.IEntity
}
```

**优势**：
- 常用字段直接访问，无需类型断言
- 字段名有代码提示
- 编译期检查，避免拼写错误
- 扩展字段仍然类型安全

### 3. SkillContext 新结构

```go
type SkillContext struct {
    // ...
    CurrentEffectSeq int32        // 当前执行的 Effect 全局序列号
    effectResults    []*EffectResult // 每个 Effect 的执行结果

    // 全局共享数据（具体类型字段）
    TotalDamage int64 // 累计总伤害
    TotalHeal   int64 // 累计总治疗
    TotalHits   int32 // 累计命中次数
    KillCount   int32 // 击杀数

    // 扩展的全局数据（如果上面的字段不够用）
    globalInt64  map[string]int64
    globalBool   map[string]bool
    globalEntity map[string][]score.IEntity
}
```

## 新 API 使用方法

### 基本用法：访问当前和上一个 Effect 的结果

```go
type ComboEffect struct{}

func (e *ComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 获取当前 Effect 的结果对象（自动创建）
    result := ctx.GetCurrentResult()
    
    // 计算伤害
    damage := int64(100)
    
    // 检查上一个 Effect 是否暴击
    if prev := ctx.GetPrevResult(); prev != nil && prev.IsCrit {
        damage *= 2 // 上一次暴击，本次伤害翻倍
    }
    
    // 暴击判定
    isCrit := calculateCrit(caster)
    if isCrit {
        damage *= 2
    }
    
    // 记录本次结果（直接赋值，类型安全）
    result.Damage = damage
    result.IsCrit = isCrit
    result.Targets = targets
    result.HitCount = int32(len(targets))
    
    // 更新全局统计
    ctx.TotalDamage += damage
    ctx.TotalHits++
    
    applyDamage(targets, damage)
}
```

### 访问任意 Effect 的结果

```go
// 获取第一个 Effect 的结果
first := ctx.GetResultBySeq(0)
if first != nil {
    log.Infof("第一段伤害: %d", first.Damage)
}

// 获取所有 Effect 的结果
allResults := ctx.GetAllResults()
for _, r := range allResults {
    log.Infof("Seq=%d, Damage=%d, Crit=%v", r.Seq, r.Damage, r.IsCrit)
}
```

### 使用扩展字段

```go
result := ctx.GetCurrentResult()

// 使用扩展字段存储特殊数据
result.ExtraInt64["mark_stacks"] = 5
result.ExtraBool["triggered_passive"] = true
result.ExtraEntity["chain_targets"] = chainTargets

// 读取扩展字段
if stacks, ok := result.ExtraInt64["mark_stacks"]; ok {
    // 使用 stacks
}
```

### 全局数据 API

```go
// 使用内置的全局统计字段
ctx.TotalDamage += damage
ctx.TotalHeal += heal
ctx.TotalHits++
ctx.KillCount++

// 使用扩展的全局数据
ctx.SetGlobalInt64("combo_count", 3)
ctx.SetGlobalBool("has_crit", true)
ctx.SetGlobalEntities("all_targets", allTargets)

// 递增全局数据
comboCount := ctx.IncrementGlobalInt64("combo_count", 1)
```

## 完整示例

### 示例1：递增连击

```go
type IncrementalComboEffect struct{}

func (e *IncrementalComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    result := ctx.GetCurrentResult()
    
    // 基础伤害（可以通过 Seq 知道是第几次执行）
    baseDamage := int64(100 + ctx.CurrentEffectSeq*50)
    
    // 检查上一次是否暴击
    if prev := ctx.GetPrevResult(); prev != nil && prev.IsCrit {
        baseDamage = baseDamage * 150 / 100 // 上次暴击，本次+50%
    }
    
    // 暴击判定
    isCrit := calculateCrit(caster)
    if isCrit {
        baseDamage *= 2
    }
    
    // 记录结果
    result.Damage = baseDamage
    result.IsCrit = isCrit
    result.Targets = targets
    result.HitCount = int32(len(targets))
    
    // 更新全局统计
    ctx.TotalDamage += baseDamage
    ctx.TotalHits++
    
    applyDamage(targets, baseDamage)
    
    log.Infof("第%d段连击: 伤害=%d, 暴击=%v, 累计=%d",
        ctx.CurrentEffectSeq+1, baseDamage, isCrit, ctx.TotalDamage)
}
```

### 示例2：标记引爆

```go
// 施加标记
type MarkEffect struct{}

func (e *MarkEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    result := ctx.GetCurrentResult()
    result.Targets = targets
    
    // 递增标记层数（使用全局数据）
    stacks := ctx.IncrementGlobalInt64("mark_stacks", 1)
    if stacks > 5 {
        ctx.SetGlobalInt64("mark_stacks", 5) // 最多5层
    }
}

// 引爆标记
type DetonateEffect struct{}

func (e *DetonateEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    result := ctx.GetCurrentResult()
    
    // 读取标记层数
    stacks, ok := ctx.GetGlobalInt64("mark_stacks")
    if !ok || stacks == 0 {
        return // 没有标记
    }
    
    // 根据层数计算伤害
    baseDamage := int64(200)
    finalDamage := baseDamage * (100 + stacks*50) / 100
    
    result.Damage = finalDamage
    result.Targets = targets
    
    ctx.TotalDamage += finalDamage
    
    applyDamage(targets, finalDamage)
    
    // 清除标记
    ctx.SetGlobalInt64("mark_stacks", 0)
}
```

### 示例3：链式闪电

```go
type ChainLightningEffect struct{}

func (e *ChainLightningEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    result := ctx.GetCurrentResult()
    
    // 读取已命中的目标（使用全局数据）
    hitTargets, _ := ctx.GetGlobalEntities("chain_hit_targets")
    hitMap := make(map[int64]bool)
    for _, t := range hitTargets {
        hitMap[int64(t.GetId())] = true
    }
    
    // 过滤掉已命中的目标
    newTargets := []score.IEntity{}
    for _, target := range targets {
        if target != nil && !hitMap[int64(target.GetId())] {
            newTargets = append(newTargets, target)
        }
    }
    
    if len(newTargets) == 0 {
        return // 没有新目标
    }
    
    // 更新已命中列表
    hitTargets = append(hitTargets, newTargets...)
    ctx.SetGlobalEntities("chain_hit_targets", hitTargets)
    
    // 伤害随弹跳次数衰减
    baseDamage := int64(300)
    damage := baseDamage * (100 - ctx.CurrentEffectSeq*20) / 100
    
    result.Damage = damage
    result.Targets = newTargets
    result.HitCount = int32(len(newTargets))
    
    ctx.TotalDamage += damage
    ctx.TotalHits += int32(len(newTargets))
    
    applyDamage(newTargets, damage)
}
```

### 示例4：击杀触发爆炸

```go
// 第一段攻击
type AssassinateEffect struct{}

func (e *AssassinateEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    result := ctx.GetCurrentResult()
    
    damage := int64(500)
    result.Damage = damage
    result.Targets = targets
    
    for _, target := range targets {
        applyDamage([]score.IEntity{target}, damage)
        
        if isTargetDead(target) {
            result.KilledAny = true
            result.ExtraEntity["kill_position"] = []score.IEntity{target}
            ctx.KillCount++
            break
        }
    }
    
    ctx.TotalDamage += damage
}

// 第二段爆炸（仅在击杀时触发）
type ExplosionEffect struct{}

func (e *ExplosionEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 检查上一个 Effect 是否击杀了目标
    prev := ctx.GetPrevResult()
    if prev == nil || !prev.KilledAny {
        return // 没有击杀，不触发爆炸
    }
    
    result := ctx.GetCurrentResult()
    
    // 在击杀位置造成范围伤害
    if killPos, ok := prev.ExtraEntity["kill_position"]; ok && len(killPos) > 0 {
        nearbyTargets := findTargetsInRadius(killPos[0].GetPos(), 5.0)
        damage := int64(200)
        
        result.Damage = damage
        result.Targets = nearbyTargets
        result.HitCount = int32(len(nearbyTargets))
        
        ctx.TotalDamage += damage
        
        applyDamage(nearbyTargets, damage)
    }
}
```

## API 总结

### Effect 结果 API

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `ctx.GetCurrentResult()` | 获取当前 Effect 的结果（自动创建） | `*EffectResult` |
| `ctx.GetPrevResult()` | 获取上一个 Effect 的结果 | `*EffectResult` (可能为 nil) |
| `ctx.GetResultBySeq(seq)` | 获取指定序列号的结果 | `*EffectResult` (可能为 nil) |
| `ctx.GetAllResults()` | 获取所有 Effect 的结果 | `[]*EffectResult` |

### EffectResult 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `Seq` | `int32` | Effect 序列号 |
| `Damage` | `int64` | 造成的伤害 |
| `Heal` | `int64` | 治疗量 |
| `IsCrit` | `bool` | 是否暴击 |
| `Targets` | `[]score.IEntity` | 命中的目标 |
| `HitCount` | `int32` | 命中次数 |
| `KilledAny` | `bool` | 是否击杀了目标 |
| `ExtraInt64` | `map[string]int64` | 扩展 int64 数据 |
| `ExtraBool` | `map[string]bool` | 扩展 bool 数据 |
| `ExtraEntity` | `map[string][]score.IEntity` | 扩展实体数据 |

### 全局数据 API

| 方法 | 说明 |
|------|------|
| `ctx.TotalDamage` | 累计总伤害（直接访问） |
| `ctx.TotalHeal` | 累计总治疗（直接访问） |
| `ctx.TotalHits` | 累计命中次数（直接访问） |
| `ctx.KillCount` | 击杀数（直接访问） |
| `ctx.SetGlobalInt64(key, value)` | 设置全局 int64 数据 |
| `ctx.GetGlobalInt64(key)` | 获取全局 int64 数据 |
| `ctx.IncrementGlobalInt64(key, delta)` | 递增全局 int64 数据 |
| `ctx.SetGlobalBool(key, value)` | 设置全局 bool 数据 |
| `ctx.GetGlobalBool(key)` | 获取全局 bool 数据 |
| `ctx.SetGlobalEntities(key, entities)` | 设置全局实体列表 |
| `ctx.GetGlobalEntities(key)` | 获取全局实体列表 |

## 优势总结

### 1. 类型安全
- 常用字段直接访问，无需类型断言
- 编译期检查，避免拼写错误
- 代码提示和自动补全

### 2. 精确标识
- Seq 唯一标识每个 Effect 实例
- 不受配置、阶段、类型的影响

### 3. 简洁易用
- `result.Damage = 100` 比 `ctx.SetData("damage", 100)` 更直观
- `prev.IsCrit` 比 `ctx.GetDataBool("crit")` 更简洁

### 4. 可扩展
- 常用字段覆盖大部分场景
- Extra 字段处理特殊情况
- 仍然保持类型安全

### 5. 性能更好
- 直接字段访问比 map 查找更快
- 减少类型断言开销
