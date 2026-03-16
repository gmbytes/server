# ScheduledEffect 索引机制说明

## 问题背景

当一个技能配置了多次执行同一个 Effect 时（通过 `Times` 和 `IntervalMs`），如果所有执行实例都使用相同的 key 写入 `SharedData`，会导致数据互相覆盖。

### 问题示例

```go
// 配置：3次连击，每次间隔500ms
{"Type": "ComboEffect", "Times": 3, "IntervalMs": 500}

// 如果 Effect 这样写：
func (e *ComboEffect) Begin(ctx *SkillContext, ...) {
    damage := int64(100)
    ctx.SetData("damage", damage)  // 问题：每次都覆盖同一个 key！
}

// 结果：第3次执行后，SharedData["damage"] 只保留最后一次的值
```

## 解决方案：执行索引机制

### 核心改动

1. **ScheduledEffect 添加 Index 字段**
   - 同一 Effect 多次执行时，Index 从 0 开始递增
   - 例如：`Times=3` 会生成 3 个 ScheduledEffect，Index 分别为 0, 1, 2

2. **SkillContext 添加 CurrentEffectIndex 字段**
   - 执行每个 Effect 前，自动设置为当前 ScheduledEffect 的 Index
   - Effect 内部可通过 `ctx.CurrentEffectIndex` 获取自己是第几次执行

3. **提供索引化的数据读写方法**
   - `SetIndexedData(baseKey, value)` - 自动生成 `baseKey_Index` 作为 key
   - `GetIndexedData(baseKey)` - 读取当前索引的数据
   - `GetPrevIndexedData(baseKey)` - 读取上一次执行的数据

## 使用方法

### 方法1：使用索引化 API（推荐）

```go
type ComboEffect struct{}

func (e *ComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    damage := int64(100)
    
    // 使用索引化 API，自动生成 "damage_0", "damage_1", "damage_2"
    ctx.SetIndexedData("damage", damage)
    
    // 读取上一次的伤害（如果存在）
    if prevDamage, ok := ctx.GetPrevIndexedDataInt64("damage"); ok {
        // 如果上一次暴击，本次伤害翻倍
        if prevCrit, ok := ctx.GetPrevIndexedData("crit"); ok && prevCrit.(bool) {
            damage *= 2
        }
    }
    
    // 记录本次是否暴击
    isCrit := calculateCrit(caster)
    ctx.SetIndexedData("crit", isCrit)
    
    applyDamage(targets, damage)
}
```

**数据存储结果**：
```
SharedData = {
    "damage_0": 100,
    "crit_0": false,
    "damage_1": 100,
    "crit_1": true,
    "damage_2": 200,  // 因为第1次暴击，第2次翻倍
    "crit_2": false,
}
```

### 方法2：手动使用 CurrentEffectIndex

```go
type ComboEffect struct{}

func (e *ComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 手动生成带索引的 key
    damageKey := fmt.Sprintf("damage_%d", ctx.CurrentEffectIndex)
    
    damage := int64(100)
    ctx.SetData(damageKey, damage)
    
    // 读取上一次的数据
    if ctx.CurrentEffectIndex > 0 {
        prevKey := fmt.Sprintf("damage_%d", ctx.CurrentEffectIndex-1)
        if prevDamage, ok := ctx.GetDataInt64(prevKey); ok {
            // 使用上一次的数据
            _ = prevDamage
        }
    }
}
```

### 方法3：使用全局累加（不需要索引）

如果只需要累加数据，不需要区分每次执行，可以直接使用全局 key：

```go
type ComboEffect struct{}

func (e *ComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    damage := int64(100)
    applyDamage(targets, damage)
    
    // 累加总伤害（所有执行实例共享）
    totalDamage := ctx.IncrementData("total_damage", damage)
    
    // 累加命中次数
    hitCount := ctx.IncrementData("hit_count", 1)
    
    // 最后一次执行时显示统计
    if ctx.CurrentEffectIndex == 2 { // 假设 Times=3，最后一次是索引2
        log.Infof("总伤害: %d, 命中: %d次", totalDamage, hitCount)
    }
}
```

## 完整示例

### 示例1：递增伤害的连击

每次连击伤害递增，且依赖上一次是否暴击

```go
type IncrementalComboEffect struct{}

func (e *IncrementalComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 基础伤害随连击次数递增
    baseDamage := int64(100 + ctx.CurrentEffectIndex*50)
    
    // 如果上一次暴击，本次伤害额外+50%
    if prevCrit, ok := ctx.GetPrevIndexedData("crit"); ok && prevCrit.(bool) {
        baseDamage = baseDamage * 150 / 100
    }
    
    // 暴击判定
    isCrit := calculateCrit(caster)
    if isCrit {
        baseDamage *= 2
    }
    
    // 记录本次数据
    ctx.SetIndexedData("damage", baseDamage)
    ctx.SetIndexedData("crit", isCrit)
    
    // 累加总伤害
    totalDamage := ctx.IncrementData("total_damage", baseDamage)
    
    applyDamage(targets, baseDamage)
    
    log.Infof("第%d段连击: 伤害=%d, 暴击=%v, 累计=%d",
        ctx.CurrentEffectIndex+1, baseDamage, isCrit, totalDamage)
}
```

**配置**：
```json
{
  "Effects": {
    "OnHit": [
      {"Type": "IncrementalCombo", "Times": 3, "IntervalMs": 500}
    ]
  }
}
```

**执行结果**：
```
第1段连击: 伤害=100, 暴击=false, 累计=100
第2段连击: 伤害=150, 暴击=true, 累计=400   (150*2=300)
第3段连击: 伤害=300, 暴击=false, 累计=700  (200*150%=300)
```

### 示例2：链式目标传递

每次命中的目标作为下次的起点

```go
type ChainEffect struct{}

func (e *ChainEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    var currentTarget score.IEntity
    
    if ctx.CurrentEffectIndex == 0 {
        // 第一次：使用初始目标
        if len(targets) > 0 {
            currentTarget = targets[0]
        }
    } else {
        // 后续次数：使用上一次命中的目标作为起点
        if prevTarget, ok := ctx.GetPrevIndexedData("target"); ok {
            if t, ok := prevTarget.(score.IEntity); ok {
                // 从上一次的目标附近寻找新目标
                currentTarget = findNearestTarget(t, 5.0)
            }
        }
    }
    
    if currentTarget == nil {
        return // 没有目标，停止链式
    }
    
    // 伤害随链式次数衰减
    baseDamage := int64(300)
    damage := baseDamage * (100 - ctx.CurrentEffectIndex*20) / 100
    
    applyDamage([]score.IEntity{currentTarget}, damage)
    
    // 记录本次命中的目标
    ctx.SetIndexedData("target", currentTarget)
    ctx.SetIndexedData("damage", damage)
}
```

### 示例3：判断是否是最后一次执行

```go
type FinalBurstEffect struct{}

func (e *FinalBurstEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    damage := int64(100)
    applyDamage(targets, damage)
    
    // 累加伤害
    totalDamage := ctx.IncrementData("total_damage", damage)
    
    // 判断是否是最后一次执行
    // 方法1：通过配置获取 Times（需要传递配置）
    // 方法2：约定最后一次触发特殊效果
    
    // 这里假设配置了 Times=5，最后一次是索引4
    if ctx.CurrentEffectIndex == 4 {
        // 最后一次：造成额外爆发伤害（等于累计伤害的50%）
        burstDamage := totalDamage / 2
        applyDamage(targets, burstDamage)
        
        log.Infof("终结爆发！累计伤害: %d, 爆发伤害: %d", totalDamage, burstDamage)
    }
}
```

## 执行顺序保证

### 严格按时间顺序执行

`Skill.Update` 方法确保 ScheduledEffect 按以下顺序执行：

1. **排序**：按 `At` 时间排序（相同时间按 `Stage` 排序）
2. **顺序执行**：`for` 循环逐个执行，**不是并发**
3. **设置索引**：执行前设置 `ctx.CurrentEffectIndex = se.Index`

```go
// @skill.go:171-190
sort.Slice(s.Pending, func(i, j int) bool {
    if s.Pending[i].At == s.Pending[j].At {
        return s.Pending[i].Stage < s.Pending[j].Stage
    }
    return s.Pending[i].At < s.Pending[j].At
})

idx := 0
for idx < len(s.Pending) && s.Pending[idx].At <= now {
    se := s.Pending[idx]
    if exec != nil {
        // 设置当前 Effect 的执行索引
        if s.Ctx != nil {
            s.Ctx.CurrentEffectIndex = se.Index
        }
        exec(se.Stage, se.Effect, s.Ctx)
    }
    idx++
}
```

### 执行时间线示例

配置：`{"Type": "ComboEffect", "Times": 3, "IntervalMs": 500}`

```
时间轴：
0ms    -> 执行 ScheduledEffect{Index: 0, At: 0}
          ctx.CurrentEffectIndex = 0
          Effect 内部可访问 SharedData["damage_0"]

500ms  -> 执行 ScheduledEffect{Index: 1, At: 500}
          ctx.CurrentEffectIndex = 1
          Effect 内部可访问 SharedData["damage_0"]（上一次）
          Effect 内部可写入 SharedData["damage_1"]（本次）

1000ms -> 执行 ScheduledEffect{Index: 2, At: 1000}
          ctx.CurrentEffectIndex = 2
          Effect 内部可访问 SharedData["damage_1"]（上一次）
          Effect 内部可写入 SharedData["damage_2"]（本次）
```

## API 总结

### SkillContext 新增字段

```go
type SkillContext struct {
    CurrentEffectIndex int32  // 当前执行的 Effect 索引
    SharedData map[string]interface{}
}
```

### 索引化数据方法

```go
// 生成带索引的 key
key := ctx.MakeIndexedKey("damage")  // "damage_0", "damage_1", ...

// 设置当前索引的数据
ctx.SetIndexedData("damage", 100)  // 自动使用 CurrentEffectIndex

// 获取当前索引的数据
value, ok := ctx.GetIndexedData("damage")

// 获取上一次执行的数据
prevValue, ok := ctx.GetPrevIndexedData("damage")
prevDamage, ok := ctx.GetPrevIndexedDataInt64("damage")
```

### 全局数据方法（所有执行实例共享）

```go
// 设置全局数据
ctx.SetData("total_damage", 500)

// 获取全局数据
value, ok := ctx.GetData("total_damage")
damage, ok := ctx.GetDataInt64("total_damage")

// 递增全局数据（常用于累加）
total := ctx.IncrementData("total_damage", 100)
```

## 最佳实践

1. **需要区分每次执行**：使用 `SetIndexedData/GetIndexedData`
2. **需要累加统计**：使用 `SetData/IncrementData`（全局 key）
3. **需要访问上一次数据**：使用 `GetPrevIndexedData`
4. **需要判断是第几次执行**：使用 `ctx.CurrentEffectIndex`

## 注意事项

1. **索引从 0 开始**：第一次执行 Index=0，第二次 Index=1，以此类推
2. **索引仅在同一 Effect 配置内有效**：不同的 Effect 配置有各自独立的索引序列
3. **数据生命周期**：SharedData 随 SkillContext 销毁，技能结束后自动清理
4. **线程安全**：当前实现不是线程安全的，确保在同一 goroutine 中访问
