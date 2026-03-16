# SharedData 数据覆盖问题解决方案

## 问题描述

当一个技能配置了多次执行同一个 Effect 时（通过 `Times` 和 `IntervalMs`），所有执行实例共享同一个 `SkillContext`。如果它们都使用相同的 key 写入 `SharedData`，会导致数据互相覆盖。

### 问题场景

```go
// 配置：3次连击
{"Type": "ComboEffect", "Times": 3, "IntervalMs": 500}

// 问题代码
func (e *ComboEffect) Begin(ctx *SkillContext, ...) {
    damage := int64(100)
    ctx.SetData("damage", damage)  // ❌ 每次都覆盖同一个 key
}

// 结果：SharedData["damage"] 只保留最后一次的值
```

## 解决方案：执行索引机制

### 核心思路

为每个 `ScheduledEffect` 添加执行索引（Index），让 Effect 知道自己是第几次执行，从而可以：
1. 使用带索引的 key 存储数据（如 `damage_0`, `damage_1`, `damage_2`）
2. 访问上一次执行的数据
3. 判断当前是第几次执行

### 实现细节

#### 1. ScheduledEffect 添加 Index 字段

```go
// @skill.go:27-34
type ScheduledEffect struct {
    At     int64
    Stage  Stage
    Effect conf.EffectCfg
    Index  int32 // 执行索引（同一 Effect 多次执行时，从 0 开始递增）
}
```

#### 2. scheduleEffect 设置索引

```go
// @skill.go:265-276
for i := int32(0); i < times; i++ {
    at := startAt + int64(i)*int64(interval)
    if endAt > 0 && at > endAt {
        break
    }
    s.Pending = append(s.Pending, ScheduledEffect{
        At:     at,
        Stage:  stage,
        Effect: eff,
        Index:  i, // 设置执行索引
    })
}
```

#### 3. Update 执行前设置 CurrentEffectIndex

```go
// @skill.go:179-190
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

#### 4. SkillContext 添加 CurrentEffectIndex

```go
// @skill_context.go:23-25
type SkillContext struct {
    // ...
    CurrentEffectIndex int32  // 当前执行的 Effect 索引
    SharedData map[string]interface{}
}
```

#### 5. 提供索引化 API

```go
// @skill_context.go:136-171

// 生成带索引的 key
func (c *SkillContext) MakeIndexedKey(baseKey string) string {
    return fmt.Sprintf("%s_%d", baseKey, c.CurrentEffectIndex)
}

// 设置带索引的数据
func (c *SkillContext) SetIndexedData(baseKey string, value interface{}) {
    c.SetData(c.MakeIndexedKey(baseKey), value)
}

// 获取当前索引的数据
func (c *SkillContext) GetIndexedData(baseKey string) (interface{}, bool) {
    return c.GetData(c.MakeIndexedKey(baseKey))
}

// 获取上一次执行的数据
func (c *SkillContext) GetPrevIndexedData(baseKey string) (interface{}, bool) {
    if c.CurrentEffectIndex <= 0 {
        return nil, false
    }
    prevKey := fmt.Sprintf("%s_%d", baseKey, c.CurrentEffectIndex-1)
    return c.GetData(prevKey)
}

// 获取上一次执行的 int64 数据
func (c *SkillContext) GetPrevIndexedDataInt64(baseKey string) (int64, bool) {
    val, ok := c.GetPrevIndexedData(baseKey)
    if !ok {
        return 0, false
    }
    if v, ok := val.(int64); ok {
        return v, true
    }
    return 0, false
}
```

## 使用方法

### 方法1：使用索引化 API（推荐）

适用于**同一 Effect 多次执行**的场景。

```go
type ComboEffect struct{}

func (e *ComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 基础伤害随连击次数递增
    baseDamage := int64(100 + ctx.CurrentEffectIndex*50)
    
    // 读取上一次是否暴击
    if prevCrit, ok := ctx.GetPrevIndexedData("crit"); ok {
        if crit, ok := prevCrit.(bool); ok && crit {
            baseDamage = baseDamage * 150 / 100  // 上次暴击，本次+50%
        }
    }
    
    // 暴击判定
    isCrit := calculateCrit(caster)
    if isCrit {
        baseDamage *= 2
    }
    
    // 使用索引化 API 记录本次数据
    ctx.SetIndexedData("damage", baseDamage)  // 自动生成 "damage_0", "damage_1", ...
    ctx.SetIndexedData("crit", isCrit)
    
    applyDamage(targets, baseDamage)
}
```

**数据存储结果**：
```
SharedData = {
    "damage_0": 100,
    "crit_0": false,
    "damage_1": 150,
    "crit_1": true,
    "damage_2": 300,  // 上次暴击，本次 200*150%=300
    "crit_2": false,
}
```

### 方法2：使用全局累加（不需要索引）

适用于**只需要累加统计**的场景。

```go
type ComboEffect struct{}

func (e *ComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    damage := int64(100)
    applyDamage(targets, damage)
    
    // 累加总伤害（所有执行实例共享同一个 key）
    totalDamage := ctx.IncrementData("total_damage", damage)
    
    // 累加命中次数
    hitCount := ctx.IncrementData("hit_count", 1)
    
    // 最后一次执行时显示统计
    // 假设配置了 Times=3，最后一次是索引2
    if ctx.CurrentEffectIndex == 2 {
        log.Infof("总伤害: %d, 命中: %d次", totalDamage, hitCount)
    }
}
```

### 方法3：不同 Effect 类型之间传递数据

适用于**不同 Effect 类型**之间传递数据的场景。

```go
// 第一个 Effect
type Effect1 struct{}

func (e *Effect1) Begin(ctx *SkillContext, ...) {
    // 使用全局 key（不同 Effect 类型，不会冲突）
    ctx.SetData("effect1_result", true)
    ctx.SetData("effect1_targets", targets)
}

// 第二个 Effect（读取第一个 Effect 的数据）
type Effect2 struct{}

func (e *Effect2) Begin(ctx *SkillContext, ...) {
    // 读取第一个 Effect 的结果
    if result, ok := ctx.GetDataBool("effect1_result"); ok && result {
        // 使用第一个 Effect 的数据
    }
    
    if targets, ok := ctx.GetDataEntities("effect1_targets"); ok {
        // 使用第一个 Effect 命中的目标
    }
}
```

## 执行顺序保证

### 严格按时间顺序执行

`Skill.Update` 确保 ScheduledEffect 按以下顺序执行：

1. **排序**：按 `At` 时间排序（相同时间按 `Stage` 排序）
2. **顺序执行**：`for` 循环逐个执行，**不是并发**
3. **设置索引**：执行前设置 `ctx.CurrentEffectIndex = se.Index`

因此，**后执行的 Effect 一定能读到前面 Effect 写入的数据**。

### 执行时间线示例

配置：`{"Type": "ComboEffect", "Times": 3, "IntervalMs": 500}`

```
0ms    -> ScheduledEffect{Index: 0, At: 0}
          ctx.CurrentEffectIndex = 0
          写入: SharedData["damage_0"] = 100

500ms  -> ScheduledEffect{Index: 1, At: 500}
          ctx.CurrentEffectIndex = 1
          读取: SharedData["damage_0"] = 100 (上一次)
          写入: SharedData["damage_1"] = 150

1000ms -> ScheduledEffect{Index: 2, At: 1000}
          ctx.CurrentEffectIndex = 2
          读取: SharedData["damage_1"] = 150 (上一次)
          写入: SharedData["damage_2"] = 200
```

## API 总结

### 索引相关

| 方法 | 说明 | 示例 |
|------|------|------|
| `ctx.CurrentEffectIndex` | 当前执行索引（0, 1, 2, ...） | `if ctx.CurrentEffectIndex == 0` |
| `ctx.MakeIndexedKey(baseKey)` | 生成带索引的 key | `"damage_0"`, `"damage_1"` |
| `ctx.SetIndexedData(baseKey, value)` | 设置当前索引的数据 | 自动使用 CurrentEffectIndex |
| `ctx.GetIndexedData(baseKey)` | 获取当前索引的数据 | 自动使用 CurrentEffectIndex |
| `ctx.GetPrevIndexedData(baseKey)` | 获取上一次的数据 | CurrentEffectIndex - 1 |
| `ctx.GetPrevIndexedDataInt64(baseKey)` | 获取上一次的 int64 数据 | 类型安全版本 |

### 全局数据（所有执行实例共享）

| 方法 | 说明 | 示例 |
|------|------|------|
| `ctx.SetData(key, value)` | 设置全局数据 | 所有 Effect 共享 |
| `ctx.GetData(key)` | 获取全局数据 | 返回 interface{} |
| `ctx.GetDataInt64(key)` | 获取 int64 数据 | 类型安全 |
| `ctx.GetDataBool(key)` | 获取 bool 数据 | 类型安全 |
| `ctx.GetDataEntities(key)` | 获取实体列表 | 类型安全 |
| `ctx.IncrementData(key, delta)` | 递增数值 | 常用于累加 |

## 最佳实践

### 1. 选择合适的数据存储方式

| 场景 | 推荐方法 | 示例 |
|------|----------|------|
| 同一 Effect 多次执行，需要区分每次 | `SetIndexedData` | 连击每段伤害不同 |
| 同一 Effect 多次执行，需要累加统计 | `IncrementData` | 累计总伤害 |
| 不同 Effect 类型之间传递数据 | `SetData` | Effect1 → Effect2 |
| 需要访问上一次执行的数据 | `GetPrevIndexedData` | 上次暴击判断 |

### 2. 命名规范

```go
// 推荐：使用清晰的前缀
ctx.SetIndexedData("combo_damage", 100)    // 生成 "combo_damage_0"
ctx.SetData("fireball_total", 500)         // 全局数据

// 不推荐：容易冲突
ctx.SetIndexedData("damage", 100)
ctx.SetData("total", 500)
```

### 3. 判断是否是最后一次执行

```go
// 方法1：通过索引判断（需要知道 Times）
if ctx.CurrentEffectIndex == 2 {  // Times=3，最后一次是索引2
    // 最后一次执行的逻辑
}

// 方法2：尝试读取下一次的数据
nextKey := fmt.Sprintf("damage_%d", ctx.CurrentEffectIndex+1)
if _, ok := ctx.GetData(nextKey); !ok {
    // 下一次不存在，说明是最后一次
}
```

## 完整示例

### 递增连击技能

```go
type IncrementalComboEffect struct{}

func (e *IncrementalComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 基础伤害随连击次数递增
    baseDamage := int64(100 + ctx.CurrentEffectIndex*50)
    
    // 如果上一次暴击，本次伤害额外+50%
    if prevCrit, ok := ctx.GetPrevIndexedData("crit"); ok {
        if crit, ok := prevCrit.(bool); ok && crit {
            baseDamage = baseDamage * 150 / 100
        }
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
第2段连击: 伤害=300, 暴击=true, 累计=400   (150*2=300)
第3段连击: 伤害=300, 暴击=false, 累计=700  (200*150%=300)

SharedData = {
    "damage_0": 100,
    "crit_0": false,
    "damage_1": 300,
    "crit_1": true,
    "damage_2": 300,
    "crit_2": false,
    "total_damage": 700,
}
```

## 总结

通过引入**执行索引机制**，我们解决了同一 Effect 多次执行时 SharedData 互相覆盖的问题：

1. ✅ **每次执行有独立的索引**：Index 从 0 开始递增
2. ✅ **自动设置 CurrentEffectIndex**：执行前自动设置
3. ✅ **提供索引化 API**：`SetIndexedData/GetIndexedData/GetPrevIndexedData`
4. ✅ **保证执行顺序**：严格按时间顺序执行，后面的 Effect 一定能读到前面的数据
5. ✅ **灵活的数据存储**：支持索引化数据和全局数据两种方式

现在你可以放心地在同一技能中多次执行同一个 Effect，每次执行都能正确地读写自己的数据，同时也能访问上一次执行的结果。
