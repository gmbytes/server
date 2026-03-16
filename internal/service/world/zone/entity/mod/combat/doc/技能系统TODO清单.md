# 技能系统完整性改进 TODO 清单

> 基于魔兽世界技能需求的系统完整性分析
> 当前完整度：约 35%
> 目标完整度：80-90%

---

## 📊 优先级说明

- 🔴 **P0 - 高优先级**：核心功能，必须实现才能支持基本的魔兽技能
- 🟡 **P1 - 中优先级**：重要功能，显著提升系统能力
- 🟢 **P2 - 低优先级**：锦上添花，增强用户体验

---

## 🔴 P0 - 高优先级任务（1-2周）

### 1. 完整的 Buff/Debuff 系统

**当前状态**：有 ApplyAura Effect 框架，但缺少核心功能（完整度：30%）

**需要实现**：

- [ ] **BuffManager 管理器**
  ```go
  type BuffManager struct {
      buffs map[int64]*BuffInstance  // BuffID -> 实例
      buffsByType map[BuffType][]*BuffInstance  // 按类型索引
  }
  ```

- [ ] **BuffInstance 实例**
  ```go
  type BuffInstance struct {
      ID           int64
      Stacks       int32              // 当前层数
      MaxStacks    int32              // 最大层数
      Duration     int64              // 持续时间
      RemainingMs  int64              // 剩余时间
      DispelType   DispelType         // 驱散类型（Magic/Curse/Poison/Disease）
      CanDispel    bool               // 是否可驱散
      CanSteal     bool               // 是否可偷取
      
      // 属性修改
      AttributeMods []AttributeModifier
      
      // 周期效果（DoT/HoT）
      PeriodicEffects []PeriodicEffect
      TickIntervalMs  int64
      LastTickMs      int64
  }
  ```

- [ ] **叠加规则**
  - [ ] 不可叠加（同名存在时失败或刷新）
  - [ ] 刷新时间（层数不变，仅刷新持续时间）
  - [ ] 叠层增加（层数增加，持续时间刷新或不刷新）
  - [ ] 独立计时（同名可多实例并存）
  - [ ] 取更强/取更长（新 Buff 更强或更长则替换）

- [ ] **驱散系统**
  ```go
  type DispelType int32
  const (
      DispelType_Magic    // 魔法
      DispelType_Curse    // 诅咒
      DispelType_Poison   // 中毒
      DispelType_Disease  // 疾病
      DispelType_Enrage   // 激怒
  )
  
  func (m *BuffManager) Dispel(dispelType DispelType, count int32) []int64
  ```

- [ ] **属性修改系统**
  ```go
  type AttributeModifier struct {
      AttrType  AttributeType  // 属性类型
      ModType   ModifierType   // 修改类型（加法/乘法）
      Value     int64          // 修改值
  }
  ```

- [ ] **DoT/HoT 周期效果**
  ```go
  type PeriodicEffect struct {
      Type       EffectType  // Damage/Heal
      Value      int64       // 每跳数值
      School     DamageSchool // 伤害学校
      CanCrit    bool        // 是否可暴击
      UseSnapshot bool       // 是否使用快照
      SnapshotData *SnapshotData // 快照数据
  }
  ```

**预计工作量**：3-4 天

---

### 2. 属性系统

**当前状态**：仅有 HP/MaxHP（完整度：10%）

**需要实现**：

- [ ] **Attributes 结构**
  ```go
  type Attributes struct {
      // 主属性
      Strength  int64  // 力量
      Agility   int64  // 敏捷
      Intellect int64  // 智力
      Stamina   int64  // 耐力
      Spirit    int64  // 精神
      
      // 战斗属性
      AttackPower   int64  // 攻击强度
      SpellPower    int64  // 法术强度
      Armor         int64  // 护甲值
      
      // 抗性（按伤害学校）
      FireResist    int64  // 火焰抗性
      FrostResist   int64  // 冰霜抗性
      NatureResist  int64  // 自然抗性
      ShadowResist  int64  // 暗影抗性
      ArcaneResist  int64  // 奥术抗性
      HolyResist    int64  // 神圣抗性
      
      // 速度属性
      AttackSpeed   float32  // 攻击速度（倍率）
      CastSpeed     float32  // 施法速度（倍率）
      MoveSpeed     float32  // 移动速度
      
      // 概率属性
      CritChance    float32  // 暴击率（%）
      HitChance     float32  // 命中率（%）
      HasteRating   int64    // 急速等级
      DodgeChance   float32  // 躲闪率（%）
      ParryChance   float32  // 招架率（%）
      BlockChance   float32  // 格挡率（%）
      BlockValue    int64    // 格挡值
      
      // 伤害修正
      DamageBonusPct    float32  // 伤害加成（%）
      DamageTakenPct    float32  // 易伤（%）
      HealingBonusPct   float32  // 治疗加成（%）
  }
  ```

- [ ] **AttributeManager 管理器**
  ```go
  type AttributeManager struct {
      base      Attributes  // 基础属性
      equipment Attributes  // 装备加成
      buffs     Attributes  // Buff 加成
      talents   Attributes  // 天赋加成
  }
  
  // 获取最终属性值
  func (m *AttributeManager) GetFinalValue(attrType AttributeType) int64
  
  // 添加/移除 Buff 属性修改
  func (m *AttributeManager) AddBuffModifier(buffID int64, mod AttributeModifier)
  func (m *AttributeManager) RemoveBuffModifier(buffID int64)
  ```

- [ ] **属性计算公式**
  - [ ] 主属性 → 战斗属性转换（力量→攻击力，智力→法强等）
  - [ ] 等级缩放
  - [ ] 加法/乘法修正器叠加规则

**预计工作量**：2-3 天

---

### 3. 完整的伤害计算系统

**当前状态**：仅有 `damage = baseDamage`（完整度：20%）

**需要实现**：

- [ ] **DamageParams 参数**
  ```go
  type DamageParams struct {
      Attacker    score.IEntity
      Target      score.IEntity
      BaseDamage  int64
      DamageType  DamageType   // Physical/Magical/True
      School      DamageSchool // Fire/Frost/Nature/Shadow/Arcane/Holy
      CanCrit     bool
      CanDodge    bool
      CanBlock    bool
      CanParry    bool
      IgnoreArmor float32      // 忽视护甲百分比
  }
  ```

- [ ] **DamageResult 结果**
  ```go
  type DamageResult struct {
      FinalDamage   int64
      IsCrit        bool
      IsDodged      bool
      IsBlocked     bool
      IsParried     bool
      IsImmune      bool
      AbsorbedByShield int64
      
      // 详细计算过程（用于调试）
      BaseDamage    int64
      AfterArmor    int64
      AfterCrit     int64
      AfterBuffs    int64
      AfterShield   int64
  }
  ```

- [ ] **伤害计算流程**
  ```go
  func CalculateDamage(params DamageParams) DamageResult {
      // 1. 基础伤害
      damage := params.BaseDamage
      
      // 2. 攻击力/法强加成
      damage += GetAttackPowerBonus(params.Attacker, params.DamageType)
      
      // 3. 命中判定
      if params.CanDodge && RollDodge(params.Target) {
          return DamageResult{IsDodged: true}
      }
      
      // 4. 招架判定
      if params.CanParry && RollParry(params.Target) {
          return DamageResult{IsParried: true}
      }
      
      // 5. 护甲/抗性减免
      if params.DamageType == Physical {
          damage = ApplyArmorReduction(damage, params.Target.Armor, params.IgnoreArmor)
      } else if params.DamageType == Magical {
          damage = ApplyResistanceReduction(damage, params.Target, params.School)
      }
      
      // 6. 暴击判定
      isCrit := false
      if params.CanCrit && RollCrit(params.Attacker) {
          damage *= 2  // 暴击倍率
          isCrit = true
      }
      
      // 7. 格挡判定
      isBlocked := false
      if params.CanBlock && RollBlock(params.Target) {
          damage -= params.Target.BlockValue
          isBlocked = true
      }
      
      // 8. Buff 伤害修正
      damage = ApplyDamageBuffs(damage, params.Attacker, params.Target)
      
      // 9. 护盾吸收
      absorbed := ApplyShieldAbsorb(damage, params.Target)
      damage -= absorbed
      
      return DamageResult{
          FinalDamage: damage,
          IsCrit: isCrit,
          IsBlocked: isBlocked,
          AbsorbedByShield: absorbed,
      }
  }
  ```

- [ ] **护甲减免公式**
  ```go
  func ApplyArmorReduction(damage int64, armor int64, ignorePercent float32) int64 {
      effectiveArmor := armor * (1 - ignorePercent)
      reduction := effectiveArmor / (effectiveArmor + 400 + 85*attackerLevel)
      return damage * (1 - reduction)
  }
  ```

- [ ] **护盾吸收系统**
  ```go
  type ShieldManager struct {
      shields []*Shield  // 按优先级排序
  }
  
  type Shield struct {
      BuffID        int64
      RemainingAmount int64
      Priority      int32
  }
  
  func (m *ShieldManager) Absorb(damage int64) int64
  ```

**预计工作量**：3-4 天

---

### 4. 资源系统

**当前状态**：完全缺失（完整度：0%）

**需要实现**：

- [ ] **ResourceType 资源类型**
  ```go
  type ResourceType int32
  const (
      Resource_Health      // 生命值
      Resource_Mana        // 法力值
      Resource_Energy      // 能量
      Resource_Rage        // 怒气
      Resource_Focus       // 集中值
      Resource_RunicPower  // 符文能量
      Resource_ComboPoints // 连击点数
      Resource_HolyPower   // 神圣能量
      Resource_SoulShards  // 灵魂碎片
      Resource_Chi         // 真气
  )
  ```

- [ ] **ResourceManager 管理器**
  ```go
  type ResourceManager struct {
      resources map[ResourceType]*Resource
  }
  
  type Resource struct {
      Type        ResourceType
      Current     int64
      Max         int64
      RegenRate   int64  // 每秒恢复速率
      LastRegenMs int64
  }
  
  // 消耗资源
  func (m *ResourceManager) Consume(resType ResourceType, amount int64) bool
  
  // 恢复资源
  func (m *ResourceManager) Restore(resType ResourceType, amount int64)
  
  // 检查是否足够
  func (m *ResourceManager) HasEnough(resType ResourceType, amount int64) bool
  
  // 自动恢复（每帧调用）
  func (m *ResourceManager) Update(deltaMs int64)
  ```

- [ ] **技能资源消耗检查**
  ```go
  // 在 Skill.CanCast 中添加资源检查
  func (s *Skill) CanCast(now int64) bool {
      // ... 现有检查 ...
      
      // 检查资源消耗
      if !CheckResourceCost(s.Cfg.ResourceCost) {
          return false
      }
      
      return true
  }
  ```

**预计工作量**：2 天

---

## 🟡 P1 - 中优先级任务（2-4周）

### 5. 战斗事件系统

**当前状态**：完全缺失（完整度：0%）

**需要实现**：

- [ ] **CombatEvent 事件类型**
  ```go
  type CombatEvent int32
  const (
      Event_OnDamageDealt     // 造成伤害时
      Event_OnDamageTaken     // 受到伤害时
      Event_OnHeal            // 治疗时
      Event_OnCrit            // 暴击时
      Event_OnKill            // 击杀时
      Event_OnCastStart       // 开始施法时
      Event_OnCastSuccess     // 施法成功时
      Event_OnCastInterrupted // 施法被打断时
      Event_OnBuffApplied     // Buff 施加时
      Event_OnBuffRemoved     // Buff 移除时
      Event_OnBuffRefresh     // Buff 刷新时
      Event_OnDodge           // 躲闪时
      Event_OnParry           // 招架时
      Event_OnBlock           // 格挡时
      Event_OnResourceChange  // 资源变化时
  )
  ```

- [ ] **EventData 事件数据**
  ```go
  type EventData struct {
      Event      CombatEvent
      Source     score.IEntity
      Target     score.IEntity
      Damage     int64
      Heal       int64
      BuffID     int64
      ResourceType ResourceType
      ResourceAmount int64
      IsCrit     bool
      // ... 其他字段
  }
  ```

- [ ] **EventListener 监听器**
  ```go
  type EventListener interface {
      OnEvent(data EventData) bool  // 返回 true 表示消费事件
      GetPriority() int32            // 优先级
  }
  
  type EventManager struct {
      listeners map[CombatEvent][]EventListener
  }
  
  func (m *EventManager) RegisterListener(event CombatEvent, listener EventListener)
  func (m *EventManager) UnregisterListener(event CombatEvent, listener EventListener)
  func (m *EventManager) TriggerEvent(data EventData)
  ```

- [ ] **触发型 Buff（Proc）**
  ```go
  type ProcBuff struct {
      *BuffInstance
      TriggerEvent  CombatEvent
      TriggerChance float32  // 触发概率
      Cooldown      int64    // 内置CD
      LastProcMs    int64
      
      OnProc func(data EventData)  // 触发时执行
  }
  ```

**预计工作量**：3-4 天

---

### 6. 控制状态管理系统

**当前状态**：完全缺失（完整度：0%）

**需要实现**：

- [ ] **CCType 控制类型**
  ```go
  type CCType int32
  const (
      CC_Stun      // 眩晕
      CC_Silence   // 沉默
      CC_Root      // 定身
      CC_Slow      // 减速
      CC_Fear      // 恐惧
      CC_Charm     // 魅惑
      CC_Sleep     // 催眠
      CC_Polymorph // 变形
      CC_Disarm    // 缴械
      CC_Blind     // 致盲
      CC_Freeze    // 冰冻
      CC_Banish    // 放逐
  )
  ```

- [ ] **CCManager 管理器**
  ```go
  type CCManager struct {
      activeCC map[CCType]*CCInstance
      
      // 免疫系统
      immunityMask CCType  // 免疫的控制类型（位掩码）
      
      // 递减抗性（DR）
      drHistory map[CCType]*DRData
  }
  
  type CCInstance struct {
      Type       CCType
      BuffID     int64
      EndMs      int64
      CanBreak   bool  // 是否可被打破（如受伤打破睡眠）
  }
  
  // 检查是否被控制
  func (m *CCManager) IsControlled(ccType CCType) bool
  
  // 检查是否免疫
  func (m *CCManager) IsImmune(ccType CCType) bool
  
  // 应用递减抗性
  func (m *CCManager) ApplyDR(ccType CCType, duration int64) int64
  ```

- [ ] **递减抗性（DR）系统**
  ```go
  type DRData struct {
      LastApplyMs int64
      ApplyCount  int32
  }
  
  // DR 公式：duration * (1 / 2^count)
  // 第1次：100%，第2次：50%，第3次：25%，第4次：免疫
  func CalculateDR(baseDuration int64, count int32) int64 {
      if count >= 3 {
          return 0  // 免疫
      }
      return baseDuration / (1 << count)
  }
  ```

**预计工作量**：2-3 天

---

### 7. 扩展的目标选择系统

**当前状态**：仅支持圆形范围（完整度：40%）

**需要实现**：

- [ ] **扩展 ShapeType**
  ```go
  type ShapeType int32
  const (
      Shape_Circle     // 圆形（已实现）
      Shape_Cone       // 扇形/锥形
      Shape_Rectangle  // 矩形
      Shape_Ring       // 环形（内外半径）
      Shape_Line       // 直线
  )
  ```

- [ ] **扩展 TargetCfg**
  ```go
  type TargetCfg struct {
      Mode     TargetMode
      Shape    ShapeType
      Radius   float32
      Angle    float32   // 扇形角度
      Width    float32   // 矩形宽度
      Length   float32   // 矩形长度
      InnerRadius float32 // 环形内半径
      
      // 目标过滤
      Relation    TargetRelation  // Self/Ally/Enemy/All
      MinHP       int64           // 最低血量
      MaxHP       int64           // 最高血量
      RequireBuff int64           // 需要的 Buff ID
      ExcludeBuff int64           // 排除的 Buff ID
      
      // 目标排序
      Sort     TargetSort  // Nearest/Farthest/LowestHP/HighestHP
      MaxCount int32       // 最大目标数
  }
  ```

- [ ] **实现各种形状的目标选择**
  ```go
  func SelectTargetsInCone(center Vector3D, direction Vector3D, radius float32, angle float32) []IEntity
  func SelectTargetsInRectangle(start Vector3D, direction Vector3D, width float32, length float32) []IEntity
  func SelectTargetsInRing(center Vector3D, innerRadius float32, outerRadius float32) []IEntity
  func SelectTargetsInLine(start Vector3D, end Vector3D, width float32) []IEntity
  ```

- [ ] **目标过滤和排序**
  ```go
  func FilterTargets(targets []IEntity, filter TargetFilter) []IEntity
  func SortTargets(targets []IEntity, sort TargetSort, maxCount int32) []IEntity
  ```

**预计工作量**：2-3 天

---

### 8. 光环系统

**当前状态**：完全缺失（完整度：0%）

**需要实现**：

- [ ] **AuraType 光环类型**
  ```go
  type AuraType int32
  const (
      Aura_Personal  // 个人光环（只影响自己）
      Aura_Range     // 范围光环（影响范围内单位）
      Aura_Party     // 小队光环
      Aura_Raid      // 团队光环
  )
  ```

- [ ] **Aura 光环实例**
  ```go
  type Aura struct {
      ID         int64
      Type       AuraType
      Radius     float32
      BuffID     int64  // 应用的 Buff ID
      Relation   TargetRelation  // 影响的目标关系
      
      // 当前影响的目标
      affectedTargets map[int64]bool
  }
  ```

- [ ] **AuraManager 管理器**
  ```go
  type AuraManager struct {
      activeAuras []*Aura
  }
  
  // 定期检测（每秒或每帧）
  func (m *AuraManager) Update(deltaMs int64) {
      for _, aura := range m.activeAuras {
          // 1. 扫描范围内的目标
          targets := FindTargetsInRange(aura.Radius, aura.Relation)
          
          // 2. 对新进入的目标施加 Buff
          for _, target := range targets {
              if !aura.affectedTargets[target.GetId()] {
                  ApplyBuff(target, aura.BuffID)
                  aura.affectedTargets[target.GetId()] = true
              }
          }
          
          // 3. 对离开的目标移除 Buff
          for targetID := range aura.affectedTargets {
              if !contains(targets, targetID) {
                  RemoveBuff(targetID, aura.BuffID)
                  delete(aura.affectedTargets, targetID)
              }
          }
      }
  }
  ```

**预计工作量**：2 天

---

## 🟢 P2 - 低优先级任务（1-2月）

### 9. 真实弹道系统

**当前状态**：仅有延迟命中（完整度：30%）

**需要实现**：

- [ ] **Projectile 弹道实体**
  ```go
  type Projectile struct {
      ID         uid.Uid
      StartPos   Vector3D
      TargetPos  Vector3D
      CurrentPos Vector3D
      Speed      float32
      
      SkillID    int64
      Caster     score.IEntity
      Target     score.IEntity
      
      OnHit      func(target score.IEntity)
      OnExpire   func()
      
      CanDodge   bool
      CanReflect bool
      CanIntercept bool
      
      CreateMs   int64
      ExpireMs   int64
  }
  ```

- [ ] **ProjectileManager 管理器**
  ```go
  type ProjectileManager struct {
      projectiles map[uid.Uid]*Projectile
  }
  
  func (m *ProjectileManager) Update(deltaMs int64) {
      for _, proj := range m.projectiles {
          // 更新位置
          proj.CurrentPos = CalculatePosition(proj, deltaMs)
          
          // 检测碰撞
          if CheckCollision(proj) {
              proj.OnHit(proj.Target)
              m.Remove(proj.ID)
          }
          
          // 检测超时
          if now >= proj.ExpireMs {
              proj.OnExpire()
              m.Remove(proj.ID)
          }
      }
  }
  ```

**预计工作量**：2-3 天

---

### 10. 快照机制

**当前状态**：完全缺失（完整度：0%）

**需要实现**：

- [ ] **SnapshotData 快照数据**
  ```go
  type SnapshotData struct {
      // 施法者属性快照
      SpellPower    int64
      AttackPower   int64
      CritChance    float32
      HasteRating   int64
      DamageBonus   float32
      
      // 快照时间
      SnapshotMs    int64
  }
  ```

- [ ] **在 DoT/HoT 中使用快照**
  ```go
  type PeriodicEffect struct {
      // ... 现有字段 ...
      
      UseSnapshot  bool
      SnapshotData *SnapshotData
  }
  
  func CalculatePeriodicDamage(effect *PeriodicEffect) int64 {
      if effect.UseSnapshot {
          // 使用快照的属性
          return effect.BaseValue + effect.SnapshotData.SpellPower * coefficient
      } else {
          // 使用当前属性（动态计算）
          return effect.BaseValue + GetCurrentSpellPower() * coefficient
      }
  }
  ```

**预计工作量**：1-2 天

---

### 11. 姿态/形态系统

**当前状态**：完全缺失（完整度：0%）

**需要实现**：

- [ ] **StanceType 姿态类型**
  ```go
  type StanceType int32
  const (
      Stance_None
      Stance_Battle      // 战斗姿态
      Stance_Defensive   // 防御姿态
      Stance_Berserker   // 狂暴姿态
      Stance_Cat         // 猫形态
      Stance_Bear        // 熊形态
      // ... 其他姿态
  )
  ```

- [ ] **StanceManager 管理器**
  ```go
  type StanceManager struct {
      currentStance StanceType
      stanceBuffs   map[StanceType]int64  // 姿态 -> Buff ID
  }
  
  func (m *StanceManager) SwitchStance(newStance StanceType) {
      // 1. 移除当前姿态的 Buff
      if oldBuffID, ok := m.stanceBuffs[m.currentStance]; ok {
          RemoveBuff(oldBuffID)
      }
      
      // 2. 应用新姿态的 Buff
      if newBuffID, ok := m.stanceBuffs[newStance]; ok {
          ApplyBuff(newBuffID)
      }
      
      // 3. 更新当前姿态
      m.currentStance = newStance
      
      // 4. 触发姿态切换事件
      TriggerEvent(Event_OnStanceChange)
  }
  ```

**预计工作量**：1-2 天

---

## 📋 实施计划

### 第 1 周（P0-1）
- [ ] Day 1-2: Buff 系统框架（BuffManager, BuffInstance）
- [ ] Day 3-4: Buff 叠加规则和驱散系统
- [ ] Day 5: 属性系统基础结构

### 第 2 周（P0-2）
- [ ] Day 1-2: 属性系统完整实现
- [ ] Day 3-4: 伤害计算系统
- [ ] Day 5: 资源系统

### 第 3 周（P1-1）
- [ ] Day 1-2: 战斗事件系统
- [ ] Day 3-4: 控制状态管理
- [ ] Day 5: 目标选择扩展

### 第 4 周（P1-2）
- [ ] Day 1-2: 光环系统
- [ ] Day 3-5: 集成测试和 Bug 修复

### 第 5-6 周（P2）
- [ ] 弹道系统
- [ ] 快照机制
- [ ] 姿态系统

---

## 🎯 里程碑

### 里程碑 1：核心战斗系统（2周后）
- ✅ Buff/Debuff 系统完整
- ✅ 属性系统完整
- ✅ 伤害计算完整
- ✅ 资源系统完整
- **系统完整度：60%**

### 里程碑 2：高级战斗机制（4周后）
- ✅ 战斗事件系统
- ✅ 控制状态管理
- ✅ 扩展目标选择
- ✅ 光环系统
- **系统完整度：80%**

### 里程碑 3：完整系统（6周后）
- ✅ 弹道系统
- ✅ 快照机制
- ✅ 姿态系统
- **系统完整度：90%**

---

## 📝 注意事项

1. **向后兼容**：新功能应该不影响现有的技能配置
2. **性能优化**：Buff 系统和光环系统需要注意性能（大量单位时）
3. **测试覆盖**：每个新功能都应该有对应的单元测试
4. **文档更新**：及时更新使用文档和示例代码
5. **配置灵活性**：所有参数应该可配置，避免硬编码

---

## 🔗 相关文档

- [魔兽技能设计 README](魔兽技能设计README.md)
- [SkillContext 新设计说明](SkillContext新设计说明.md)
- [GlobalDataKey 使用指南](GlobalDataKey使用指南.md)
- [技能执行流程详解](./技能执行流程详解.md)

---

**最后更新**：2026-02-03
**当前完整度**：35%
**目标完整度**：90%
