# GlobalDataKey 使用指南

## 设计目标

使用**类型化的键（GlobalDataKey）**替代字符串 key，避免拼写错误，提供编译期检查和代码提示。

## 核心设计

### 1. GlobalDataKey 类型定义

```go
// GlobalDataKey 全局数据键类型（避免字符串拼写错误）
type GlobalDataKey string
```

### 2. 预定义的常用键

```go
const (
    // 标记相关
    GlobalKey_MarkStacks    GlobalDataKey = "mark_stacks"
    GlobalKey_MarkTargets   GlobalDataKey = "mark_targets"
    
    // 连击相关
    GlobalKey_ComboCount    GlobalDataKey = "combo_count"
    GlobalKey_ComboCrit     GlobalDataKey = "combo_crit"
    GlobalKey_ComboTargets  GlobalDataKey = "combo_targets"
    
    // 链式相关
    GlobalKey_ChainHitTargets GlobalDataKey = "chain_hit_targets"
    GlobalKey_ChainCount      GlobalDataKey = "chain_count"
    GlobalKey_ChainLastTarget GlobalDataKey = "chain_last_target"
    
    // 触发相关
    GlobalKey_HasTriggered  GlobalDataKey = "has_triggered"
    GlobalKey_TriggerCount  GlobalDataKey = "trigger_count"
)
```

### 3. 数据结构

```go
type SkillContext struct {
    // 全局数据使用 GlobalDataKey 作为 key
    globalInt64  map[GlobalDataKey]int64
    globalBool   map[GlobalDataKey]bool
    globalEntity map[GlobalDataKey][]score.IEntity
}

type EffectResult struct {
    // 扩展字段也使用 GlobalDataKey
    ExtraInt64  map[GlobalDataKey]int64
    ExtraBool   map[GlobalDataKey]bool
    ExtraEntity map[GlobalDataKey][]score.IEntity
}
```

## 使用方法

### 方法1：使用预定义的键（推荐）

```go
type MarkEffect struct{}

func (e *MarkEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // ✅ 使用预定义的键，有代码提示，不会拼写错误
    stacks := ctx.IncrementGlobalInt64(GlobalKey_MarkStacks, 1)
    
    if stacks > 5 {
        ctx.SetGlobalInt64(GlobalKey_MarkStacks, 5)
    }
    
    ctx.SetGlobalEntities(GlobalKey_MarkTargets, targets)
}

type DetonateEffect struct{}

func (e *DetonateEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // ✅ 读取时也使用预定义的键
    stacks, ok := ctx.GetGlobalInt64(GlobalKey_MarkStacks)
    if !ok || stacks == 0 {
        return
    }
    
    markedTargets, _ := ctx.GetGlobalEntities(GlobalKey_MarkTargets)
    
    // 计算伤害
    damage := int64(200) * (100 + stacks*50) / 100
    applyDamage(markedTargets, damage)
    
    // 清除标记
    ctx.SetGlobalInt64(GlobalKey_MarkStacks, 0)
}
```

**对比旧方式**：
```go
// ❌ 旧方式：容易拼写错误
ctx.IncrementGlobalInt64("mark_stacks", 1)    // 正确
ctx.GetGlobalInt64("mark_stack")              // 错误！少了 s
ctx.GetGlobalInt64("makr_stacks")             // 错误！拼写错误
ctx.GetGlobalInt64("Mark_Stacks")             // 错误！大小写错误
```

### 方法2：自定义键

如果预定义的键不够用，可以自定义：

```go
// 定义自己的键常量
const (
    MyKey_CustomData GlobalDataKey = "my_custom_data"
    MyKey_SpecialFlag GlobalDataKey = "special_flag"
)

func (e *MyEffect) Begin(ctx *SkillContext, ...) {
    // 使用自定义键
    ctx.SetGlobalInt64(MyKey_CustomData, 100)
    ctx.SetGlobalBool(MyKey_SpecialFlag, true)
}
```

或者临时使用：

```go
func (e *MyEffect) Begin(ctx *SkillContext, ...) {
    // 临时转换（不推荐，失去了类型安全的优势）
    ctx.SetGlobalInt64(GlobalDataKey("temp_key"), 100)
}
```

### 方法3：在 EffectResult 中使用

```go
func (e *MyEffect) Begin(ctx *SkillContext, ...) {
    result := ctx.GetCurrentResult()
    
    // ✅ 使用预定义的键
    result.ExtraInt64[GlobalKey_MarkStacks] = 5
    result.ExtraBool[GlobalKey_HasTriggered] = true
    result.ExtraEntity[GlobalKey_ChainHitTargets] = targets
    
    // 读取
    if stacks, ok := result.ExtraInt64[GlobalKey_MarkStacks]; ok {
        // 使用 stacks
    }
}
```

## 完整示例

### 示例1：连击系统

```go
type ComboEffect struct{}

func (e *ComboEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    result := ctx.GetCurrentResult()
    
    // 递增连击次数
    comboCount := ctx.IncrementGlobalInt64(GlobalKey_ComboCount, 1)
    
    // 基础伤害
    damage := int64(100 + comboCount*50)
    
    // 检查上一次是否暴击
    if hasCrit, ok := ctx.GetGlobalBool(GlobalKey_ComboCrit); ok && hasCrit {
        damage = damage * 150 / 100
    }
    
    // 暴击判定
    isCrit := calculateCrit(caster)
    if isCrit {
        damage *= 2
    }
    
    // 记录本次结果
    result.Damage = damage
    result.IsCrit = isCrit
    result.Targets = targets
    
    // 更新全局状态
    ctx.SetGlobalBool(GlobalKey_ComboCrit, isCrit)
    ctx.SetGlobalEntities(GlobalKey_ComboTargets, targets)
    ctx.TotalDamage += damage
    
    applyDamage(targets, damage)
}
```

### 示例2：链式闪电

```go
type ChainLightningEffect struct{}

func (e *ChainLightningEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    result := ctx.GetCurrentResult()
    
    // 读取已命中的目标
    hitTargets, _ := ctx.GetGlobalEntities(GlobalKey_ChainHitTargets)
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
        return
    }
    
    // 更新已命中列表
    hitTargets = append(hitTargets, newTargets...)
    ctx.SetGlobalEntities(GlobalKey_ChainHitTargets, hitTargets)
    
    // 递增弹跳次数
    chainCount := ctx.IncrementGlobalInt64(GlobalKey_ChainCount, 1)
    
    // 伤害衰减
    baseDamage := int64(300)
    damage := baseDamage * (100 - (chainCount-1)*20) / 100
    
    result.Damage = damage
    result.Targets = newTargets
    result.HitCount = int32(len(newTargets))
    
    ctx.TotalDamage += damage
    
    applyDamage(newTargets, damage)
    
    // 记录最后命中的目标
    if len(newTargets) > 0 {
        ctx.SetGlobalEntities(GlobalKey_ChainLastTarget, []score.IEntity{newTargets[0]})
    }
}
```

### 示例3：触发系统

```go
type PassiveTriggerEffect struct{}

func (e *PassiveTriggerEffect) Begin(ctx *SkillContext, caster score.IEntity, targets []score.IEntity) {
    // 检查是否已触发
    if hasTriggered, ok := ctx.GetGlobalBool(GlobalKey_HasTriggered); ok && hasTriggered {
        return // 已触发，不再执行
    }
    
    result := ctx.GetCurrentResult()
    
    // 执行触发效果
    damage := int64(500)
    result.Damage = damage
    result.Targets = targets
    
    ctx.TotalDamage += damage
    
    applyDamage(targets, damage)
    
    // 标记已触发
    ctx.SetGlobalBool(GlobalKey_HasTriggered, true)
    
    // 递增触发次数
    ctx.IncrementGlobalInt64(GlobalKey_TriggerCount, 1)
}
```

## 如何添加新的预定义键

当你需要添加新的常用键时，在 `skill_context.go` 中添加：

```go
const (
    // ... 现有的键 ...
    
    // 你的新键
    GlobalKey_YourNewKey GlobalDataKey = "your_new_key"
)
```

**命名规范**：
- 使用 `GlobalKey_` 前缀
- 使用 PascalCase（大驼峰）
- 描述性的名称

**示例**：
```go
const (
    GlobalKey_ShieldAmount     GlobalDataKey = "shield_amount"
    GlobalKey_BuffDuration     GlobalDataKey = "buff_duration"
    GlobalKey_CooldownReduction GlobalDataKey = "cooldown_reduction"
)
```

## 优势总结

### 1. 编译期检查
```go
// ✅ 编译通过
ctx.SetGlobalInt64(GlobalKey_MarkStacks, 5)

// ❌ 编译错误：undefined: GlobalKey_MarkStack
ctx.SetGlobalInt64(GlobalKey_MarkStack, 5)
```

### 2. 代码提示
IDE 会自动提示所有可用的 `GlobalKey_*` 常量，无需记忆。

### 3. 重构友好
如果需要修改键名，只需修改常量定义，所有使用的地方会自动更新。

### 4. 可读性更好
```go
// ✅ 清晰易读
ctx.GetGlobalInt64(GlobalKey_MarkStacks)

// ❌ 需要记忆字符串
ctx.GetGlobalInt64("mark_stacks")
```

### 5. 避免拼写错误
```go
// ❌ 旧方式：容易出错
ctx.SetGlobalInt64("mark_stacks", 5)
ctx.GetGlobalInt64("mark_stack")   // 错误！少了 s

// ✅ 新方式：编译器保证正确
ctx.SetGlobalInt64(GlobalKey_MarkStacks, 5)
ctx.GetGlobalInt64(GlobalKey_MarkStacks)  // 正确
```

## API 对比

### 旧 API（字符串 key）
```go
// ❌ 容易出错
ctx.SetGlobalInt64("mark_stacks", 5)
ctx.GetGlobalInt64("mark_stack")  // 拼写错误
ctx.GetGlobalBool("has_triggered")
ctx.SetGlobalEntities("chain_targets", targets)
```

### 新 API（GlobalDataKey）
```go
// ✅ 类型安全
ctx.SetGlobalInt64(GlobalKey_MarkStacks, 5)
ctx.GetGlobalInt64(GlobalKey_MarkStacks)  // 编译器检查
ctx.GetGlobalBool(GlobalKey_HasTriggered)
ctx.SetGlobalEntities(GlobalKey_ChainHitTargets, targets)
```

## 总结

使用 `GlobalDataKey` 类型替代字符串 key，带来以下好处：

1. ✅ **编译期检查**：拼写错误在编译时发现
2. ✅ **代码提示**：IDE 自动提示可用的键
3. ✅ **重构友好**：修改键名只需改一处
4. ✅ **可读性好**：常量名比字符串更清晰
5. ✅ **避免错误**：消除了字符串拼写错误的可能

这是一个简单但有效的改进，显著提升了代码质量和开发体验。
