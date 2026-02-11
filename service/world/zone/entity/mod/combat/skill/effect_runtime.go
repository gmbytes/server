package skill

import (
	"server/lib/uid"
	"server/service/world/zone/izone"
	"time"
)

// EffectState 效果生命周期状态
type EffectState int32

const (
	EffectState_Pending   EffectState = 0 // 待激活
	EffectState_Active    EffectState = 1 // 激活中
	EffectState_Paused    EffectState = 2 // 暂停
	EffectState_Finished  EffectState = 3 // 正常结束
	EffectState_Cancelled EffectState = 4 // 被取消/驱散
)

// EffectRuntime 效果运行时数据
// 用于管理需要持续执行的效果（DoT/HoT/Buff等）
type EffectRuntime struct {
	Id uid.Uid // 效果实例唯一ID

	Effect skillEffect   // 效果实例
	Ctx    *SkillContext // 效果上下文

	Caster  izone.IEntity   // 施法者
	Targets []izone.IEntity // 目标列表

	StartMs    int64 // 开始时间（毫秒时间戳）
	EndMs      int64 // 结束时间（毫秒时间戳）
	LastTickMs int64 // 上次Tick时间

	TickIntervalMs int64 // Tick间隔（毫秒）
	TickCount      int32 // 已执行Tick次数
	MaxTicks       int32 // 最大Tick次数（0表示无限制）

	State EffectState // 生命周期状态

	// 生命周期回调（可选）
	OnActivate func(*EffectRuntime) // 激活时回调
	OnPause    func(*EffectRuntime) // 暂停时回调
	OnResume   func(*EffectRuntime) // 恢复时回调
	OnFinish   func(*EffectRuntime) // 正常结束时回调
	OnCancel   func(*EffectRuntime) // 取消时回调
}

// NewEffectRuntime 创建效果运行时实例
func NewEffectRuntime(effect skillEffect, ctx *SkillContext, caster izone.IEntity, targets []izone.IEntity) *EffectRuntime {
	return &EffectRuntime{
		Id:      uid.Gen(),
		Effect:  effect,
		Ctx:     ctx,
		Caster:  caster,
		Targets: targets,
		State:   EffectState_Pending,
	}
}

// Activate 激活效果
func (r *EffectRuntime) Activate(nowMs int64) {
	if r.State != EffectState_Pending {
		return
	}

	r.State = EffectState_Active
	r.StartMs = nowMs
	r.LastTickMs = nowMs

	if r.OnActivate != nil {
		r.OnActivate(r)
	}
}

// Pause 暂停效果
func (r *EffectRuntime) Pause() {
	if r.State != EffectState_Active {
		return
	}

	r.State = EffectState_Paused

	if r.OnPause != nil {
		r.OnPause(r)
	}
}

// Resume 恢复效果
func (r *EffectRuntime) Resume(nowMs int64) {
	if r.State != EffectState_Paused {
		return
	}

	r.State = EffectState_Active
	r.LastTickMs = nowMs // 重置Tick时间

	if r.OnResume != nil {
		r.OnResume(r)
	}
}

// IsRunning 判断效果是否正在运行
func (r *EffectRuntime) IsRunning() bool {
	return r.State == EffectState_Active
}

// ShouldTick 判断是否应该执行Tick
func (r *EffectRuntime) ShouldTick(nowMs int64) bool {
	if r.State != EffectState_Active {
		return false
	}

	// 检查是否到达Tick时间
	if r.TickIntervalMs > 0 && nowMs-r.LastTickMs < r.TickIntervalMs {
		return false
	}

	// 检查是否达到最大Tick次数
	if r.MaxTicks > 0 && r.TickCount >= r.MaxTicks {
		return false
	}

	return true
}

// IsExpired 判断效果是否过期
func (r *EffectRuntime) IsExpired(nowMs int64) bool {
	if r.State == EffectState_Finished || r.State == EffectState_Cancelled {
		return true
	}

	if r.State != EffectState_Active {
		return false
	}

	if r.EndMs > 0 && nowMs >= r.EndMs {
		return true
	}

	if r.MaxTicks > 0 && r.TickCount >= r.MaxTicks {
		return true
	}

	return false
}

// DoTick 执行一次Tick
func (r *EffectRuntime) DoTick(nowMs int64) {
	if !r.ShouldTick(nowMs) {
		return
	}

	delta := time.Duration(nowMs-r.LastTickMs) * time.Millisecond
	r.Effect.Update(r.Ctx, delta)

	r.LastTickMs = nowMs
	r.TickCount++
}

// Finish 正常结束效果
func (r *EffectRuntime) Finish() {
	if r.State == EffectState_Finished || r.State == EffectState_Cancelled {
		return
	}

	r.Effect.End(r.Ctx)
	r.State = EffectState_Finished

	if r.OnFinish != nil {
		r.OnFinish(r)
	}
}

// Cancel 取消效果（需要回滚）
func (r *EffectRuntime) Cancel() {
	if r.State == EffectState_Finished || r.State == EffectState_Cancelled {
		return
	}

	r.Effect.Revert(r.Ctx)
	r.State = EffectState_Cancelled

	if r.OnCancel != nil {
		r.OnCancel(r)
	}
}

// GetRemainingMs 获取剩余时间（毫秒）
func (r *EffectRuntime) GetRemainingMs(nowMs int64) int64 {
	if r.EndMs <= 0 {
		return -1 // 无限持续
	}

	remaining := r.EndMs - nowMs
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetProgress 获取进度（0.0 - 1.0）
func (r *EffectRuntime) GetProgress(nowMs int64) float32 {
	if r.EndMs <= 0 || r.StartMs <= 0 {
		return 0.0
	}

	total := r.EndMs - r.StartMs
	if total <= 0 {
		return 1.0
	}

	elapsed := nowMs - r.StartMs
	if elapsed >= total {
		return 1.0
	}

	return float32(elapsed) / float32(total)
}
