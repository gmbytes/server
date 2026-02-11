# SkillContext 共享数据使用指南

## 概述

`SkillContext.SharedData` 用于在同一技能的多个 `ScheduledEffect` 之间传递中间数据，解决后续 Effect 需要依赖前面 Effect 产生数据的问题。

## 核心 API

### 基础方法

```go
// 设置任意类型数据
ctx.SetData(key string, value interface{})

// 获取任意类型数据
value, ok := ctx.GetData(key string)
```

### 类型安全方法

```go
// 获取 int64 类型数据
value, ok := ctx.GetDataInt64(key string)

// 获取 bool 类型数据
value, ok := ctx.GetDataBool(key string)

// 获取 float64 类型数据
value, ok := ctx.GetDataFloat64(key string)

// 获取实体列表
entities, ok := ctx.GetDataEntities(key string)

// 递增数值（常用于计数器）
newValue := ctx.IncrementData(key string, delta int64)

// 清空所有共享数据
ctx.ClearData()
```

## 使用场景

### 场景1：连击技能 - 后续段依赖前面段的结果

**需求**：第一段攻击如果暴击，第二段伤害翻倍

```go
// 第一段攻击
type ComboEffect1 struct{}

func (e *ComboEffect1) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    damage := int64(100)
    isCrit := calculateCrit(caster) // 暴击判定
    
    if isCrit {
        damage *= 2
        ctx.SetData("combo_crit", true)  // 记录暴击状态
    }
    
    ctx.SetData("combo_damage", damage)
    applyDamage(targets, damage)
}

// 第二段攻击（延迟500ms执行）
type ComboEffect2 struct{}

func (e *ComboEffect2) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    baseDamage := int64(150)
    
    // 读取第一段是否暴击
    if isCrit, ok := ctx.GetDataBool("combo_crit"); ok && isCrit {
        baseDamage *= 2  // 第一段暴击，第二段伤害翻倍
    }
    
    applyDamage(targets, baseDamage)
}
```

**配置示例**：
```json
{
  "Effects": {
    "OnHit": [
      {"Type": "ComboEffect1", "Times": 1, "IntervalMs": 0},
      {"Type": "ComboEffect2", "Times": 1, "IntervalMs": 500}
    ]
  }
}
```

### 场景2：标记引爆 - 累积层数后引爆

**需求**：施加标记，引爆时根据标记层数造成伤害

```go
// 施加标记（可多次施放）
type MarkEffect struct{}

func (e *MarkEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 递增标记层数
    stacks := ctx.IncrementData("mark_stacks", 1)
    
    // 记录被标记的目标
    ctx.SetData("marked_targets", targets)
    
    // 最多5层
    if stacks > 5 {
        ctx.SetData("mark_stacks", int64(5))
    }
}

// 引爆标记
type DetonateEffect struct{}

func (e *DetonateEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 读取标记层数
    stacks, ok := ctx.GetDataInt64("mark_stacks")
    if !ok || stacks == 0 {
        return // 没有标记，不造成伤害
    }
    
    // 每层增加50%伤害
    baseDamage := int64(200)
    finalDamage := baseDamage * (100 + stacks*50) / 100
    
    // 对被标记的目标造成伤害
    if markedTargets, ok := ctx.GetDataEntities("marked_targets"); ok {
        applyDamage(markedTargets, finalDamage)
    }
    
    // 清除标记
    ctx.SetData("mark_stacks", int64(0))
}
```

### 场景3：链式闪电 - 每次弹跳依赖上次命中的目标

**需求**：闪电从主目标弹跳到附近目标，每次弹跳伤害衰减，不重复命中

```go
type ChainLightningEffect struct{}

func (e *ChainLightningEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 读取已命中的目标（避免重复）
    hitTargets := make(map[int64]bool)
    if prevHit, ok := ctx.GetData("chain_hit_targets"); ok {
        if m, ok := prevHit.(map[int64]bool); ok {
            hitTargets = m
        }
    }
    
    // 过滤掉已命中的目标
    newTargets := []score.IEntity{}
    for _, target := range targets {
        if target != nil && !hitTargets[int64(target.GetId())] {
            newTargets = append(newTargets, target)
            hitTargets[int64(target.GetId())] = true
        }
    }
    
    if len(newTargets) == 0 {
        return // 没有新目标，停止弹跳
    }
    
    // 更新已命中列表
    ctx.SetData("chain_hit_targets", hitTargets)
    
    // 递增弹跳次数
    chainCount := ctx.IncrementData("chain_count", 1)
    
    // 伤害随弹跳次数衰减（每次减少20%）
    baseDamage := int64(300)
    damage := baseDamage * (100 - (chainCount-1)*20) / 100
    
    applyDamage(newTargets, damage)
    
    // 记录最后命中的目标，作为下次弹跳的起点
    if len(newTargets) > 0 {
        ctx.SetData("chain_last_target", newTargets[0])
    }
}
```

**配置示例**（5次弹跳，每次间隔200ms）：
```json
{
  "Effects": {
    "OnHit": [
      {"Type": "ChainLightning", "Times": 5, "IntervalMs": 200}
    ]
  }
}
```

### 场景4：累计伤害统计

**需求**：多段攻击后统计总伤害、命中次数、平均伤害

```go
type MultiHitEffect struct{}

func (e *MultiHitEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    damage := calculateDamage(caster, targets)
    applyDamage(targets, damage)
    
    // 累计总伤害
    totalDamage := ctx.IncrementData("total_damage", damage)
    
    // 累计命中次数
    hitCount := ctx.IncrementData("hit_count", 1)
    
    // 计算平均伤害
    avgDamage := totalDamage / hitCount
    ctx.SetData("avg_damage", avgDamage)
}

// 最后一段显示统计信息
type ShowStatsEffect struct{}

func (e *ShowStatsEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    totalDamage, _ := ctx.GetDataInt64("total_damage")
    hitCount, _ := ctx.GetDataInt64("hit_count")
    avgDamage, _ := ctx.GetDataInt64("avg_damage")
    
    // 显示统计信息给玩家
    showMessage(caster, fmt.Sprintf(
        "总伤害: %d, 命中: %d次, 平均: %d",
        totalDamage, hitCount, avgDamage,
    ))
}
```

### 场景5：条件触发 - 根据前面 Effect 的结果决定是否执行

**需求**：只有第一段攻击击杀目标时，才触发第二段范围爆炸

```go
// 第一段攻击
type AssassinateEffect struct{}

func (e *AssassinateEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    damage := int64(500)
    
    for _, target := range targets {
        applyDamage([]score.IEntity{target}, damage)
        
        // 检查是否击杀
        if isTargetDead(target) {
            ctx.SetData("assassination_kill", true)
            ctx.SetData("kill_position", target.GetPos())
            break
        }
    }
}

// 第二段爆炸（仅在击杀时触发）
type ExplosionEffect struct{}

func (e *ExplosionEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 检查是否有击杀
    hasKill, ok := ctx.GetDataBool("assassination_kill")
    if !ok || !hasKill {
        return // 没有击杀，不触发爆炸
    }
    
    // 在击杀位置造成范围伤害
    if killPos, ok := ctx.GetData("kill_position"); ok {
        nearbyTargets := findTargetsInRadius(killPos, 5.0)
        applyDamage(nearbyTargets, 200)
    }
}
```

## 最佳实践

### 1. 命名规范

使用清晰的前缀区分不同技能的数据：

```go
// 推荐
ctx.SetData("fireball_crit", true)
ctx.SetData("combo_stacks", 3)
ctx.SetData("chain_targets", targets)

// 不推荐（容易冲突）
ctx.SetData("crit", true)
ctx.SetData("stacks", 3)
ctx.SetData("targets", targets)
```

### 2. 类型安全

优先使用类型安全的方法：

```go
// 推荐
if value, ok := ctx.GetDataInt64("damage"); ok {
    // 使用 value
}

// 不推荐
if value, ok := ctx.GetData("damage"); ok {
    damage := value.(int64) // 可能 panic
}
```

### 3. 数据清理

技能结束后，SharedData 会随 SkillContext 一起销毁，无需手动清理。

但如果需要在技能执行中途清理某些数据：

```go
// 清理特定数据
ctx.SetData("temp_data", nil)

// 清空所有数据（慎用）
ctx.ClearData()
```

### 4. 调试技巧

在开发时可以打印 SharedData 内容：

```go
func (e *MyEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 调试：打印当前所有共享数据
    for key, value := range ctx.SharedData {
        log.Debugf("SharedData[%s] = %v", key, value)
    }
}
```

## 注意事项

1. **生命周期**：SharedData 的生命周期与 SkillContext 绑定，技能结束后自动销毁
2. **线程安全**：当前实现不是线程安全的，确保在同一个 goroutine 中访问
3. **类型转换**：使用 `GetData` 时需要手动类型断言，建议使用类型安全的方法
4. **性能**：SharedData 使用 map 存储，读写性能良好，但避免存储大量数据
5. **作用域**：SharedData 仅在同一次技能释放的多个 Effect 之间共享，不同技能实例之间不共享

## 与 EffectRuntime 的区别

- **SkillContext.SharedData**：用于同一技能的多个 **ScheduledEffect** 之间传递数据
- **EffectRuntime**：用于管理单个持续性 Effect 的生命周期（DoT/HoT/Buff）

两者配合使用可以实现复杂的技能逻辑。
