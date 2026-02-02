package combat

import (
	"server/lib/container"
	"server/lib/uid"
	"server/service/scene/entity/mod/combat/skill"
)

// EffectManager 效果管理器
// 负责管理所有持续性效果的运行时数据
type EffectManager struct {
	owner *CombatManager

	// 运行中的效果列表
	runningEffects *container.LMap[uid.Uid, *skill.EffectRuntime]

	nowMs int64
}

func newEffectManager(combatMgr *CombatManager) *EffectManager {
	return &EffectManager{
		owner:          combatMgr,
		runningEffects: container.NewLMap[uid.Uid, *skill.EffectRuntime](),
	}
}

// AddEffect 添加持续性效果
func (m *EffectManager) AddEffect(runtime *skill.EffectRuntime) {
	if runtime == nil {
		return
	}

	// 激活效果
	runtime.Activate(m.nowMs)

	m.runningEffects.Set(runtime.Id, runtime)
}

// RemoveEffect 移除效果
func (m *EffectManager) RemoveEffect(effectId uid.Uid) {
	runtime, ok := m.runningEffects.Get(effectId)
	if !ok {
		return
	}

	runtime.Finish()
	m.runningEffects.Delete(effectId)
}

// CancelEffect 取消效果（带回滚）
func (m *EffectManager) CancelEffect(effectId uid.Uid) {
	runtime, ok := m.runningEffects.Get(effectId)
	if !ok {
		return
	}

	runtime.Cancel()
	m.runningEffects.Delete(effectId)
}

// Update 更新所有效果
func (m *EffectManager) Update(deltaMs int64) {
	m.nowMs += deltaMs

	// 收集过期的效果
	expiredIds := make([]uid.Uid, 0)

	for _, entry := range m.runningEffects.Entries() {
		runtime := entry.Value
		// 检查是否过期
		if runtime.IsExpired(m.nowMs) {
			runtime.Finish()
			expiredIds = append(expiredIds, entry.Key)
			continue
		}

		// 执行Tick
		runtime.DoTick(m.nowMs)
	}

	// 清理过期效果
	for _, id := range expiredIds {
		m.runningEffects.Delete(id)
	}
}

// GetEffect 获取效果运行时数据
func (m *EffectManager) GetEffect(effectId uid.Uid) *skill.EffectRuntime {
	runtime, _ := m.runningEffects.Get(effectId)
	return runtime
}

// GetEffectsByTarget 获取目标身上的所有效果
func (m *EffectManager) GetEffectsByTarget(targetId uid.Uid) []*skill.EffectRuntime {
	result := make([]*skill.EffectRuntime, 0)

	m.runningEffects.ForEach(func(runtime *skill.EffectRuntime) {
		for _, target := range runtime.Targets {
			if target.GetId() == targetId {
				result = append(result, runtime)
				break
			}
		}
	})

	return result
}

// PauseEffect 暂停效果
func (m *EffectManager) PauseEffect(effectId uid.Uid) {
	runtime, ok := m.runningEffects.Get(effectId)
	if !ok {
		return
	}

	runtime.Pause()
}

// ResumeEffect 恢复效果
func (m *EffectManager) ResumeEffect(effectId uid.Uid) {
	runtime, ok := m.runningEffects.Get(effectId)
	if !ok {
		return
	}

	runtime.Resume(m.nowMs)
}

// PauseAllEffects 暂停目标身上的所有效果
func (m *EffectManager) PauseAllEffects(targetId uid.Uid) {
	m.runningEffects.ForEach(func(runtime *skill.EffectRuntime) {
		for _, target := range runtime.Targets {
			if target.GetId() == targetId {
				runtime.Pause()
				break
			}
		}
	})
}

// ResumeAllEffects 恢复目标身上的所有效果
func (m *EffectManager) ResumeAllEffects(targetId uid.Uid) {
	m.runningEffects.ForEach(func(runtime *skill.EffectRuntime) {
		for _, target := range runtime.Targets {
			if target.GetId() == targetId {
				runtime.Resume(m.nowMs)
				break
			}
		}
	})
}

// GetActiveEffectCount 获取激活中的效果数量
func (m *EffectManager) GetActiveEffectCount() int {
	count := 0
	m.runningEffects.ForEach(func(runtime *skill.EffectRuntime) {
		if runtime.IsRunning() {
			count++
		}
	})
	return count
}

// GetEffectState 获取效果状态
func (m *EffectManager) GetEffectState(effectId uid.Uid) skill.EffectState {
	runtime, ok := m.runningEffects.Get(effectId)
	if !ok {
		return skill.EffectState_Finished
	}
	return runtime.State
}

// GetEffectProgress 获取效果进度
func (m *EffectManager) GetEffectProgress(effectId uid.Uid) float32 {
	runtime, ok := m.runningEffects.Get(effectId)
	if !ok {
		return 1.0
	}
	return runtime.GetProgress(m.nowMs)
}

// GetEffectRemainingMs 获取效果剩余时间
func (m *EffectManager) GetEffectRemainingMs(effectId uid.Uid) int64 {
	runtime, ok := m.runningEffects.Get(effectId)
	if !ok {
		return 0
	}
	return runtime.GetRemainingMs(m.nowMs)
}

// Clear 清空所有效果
func (m *EffectManager) Clear() {
	m.runningEffects.ForEach(func(runtime *skill.EffectRuntime) {
		runtime.Finish()
	})
	m.runningEffects.Clear()
}

// ClearByTarget 清空目标身上的所有效果
func (m *EffectManager) ClearByTarget(targetId uid.Uid) {
	toRemove := make([]uid.Uid, 0)

	for _, entry := range m.runningEffects.Entries() {
		runtime := entry.Value
		for _, target := range runtime.Targets {
			if target.GetId() == targetId {
				runtime.Finish()
				toRemove = append(toRemove, entry.Key)
				break
			}
		}
	}

	for _, id := range toRemove {
		m.runningEffects.Delete(id)
	}
}
