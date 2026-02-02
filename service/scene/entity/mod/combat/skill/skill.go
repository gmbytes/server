package skill

import (
	"server/data/conf"
	"sort"
)

// Stage 表示技能运行时的阶段（Effect 执行时机）。
// 配置层通过 SkillEffects 将 EffectCfg 分配到不同阶段；运行时在相应时机触发。
type Stage int32

const (
	// Stage_Invalid 无效阶段。
	Stage_Invalid Stage = 0
	// Stage_CastStart 开始施法（校验通过后进入施法/释放流程）。
	Stage_CastStart Stage = 1
	// Stage_CastFinish 吟唱结束/释放成功。
	Stage_CastFinish Stage = 2
	// Stage_Channel 引导阶段 Tick。
	Stage_Channel Stage = 3
	// Stage_Hit 命中阶段（弹道到达/范围生效）。
	Stage_Hit Stage = 4
	// Stage_Cancel 取消/被打断。
	Stage_Cancel Stage = 5
)

// ScheduledEffect 为延迟执行的 Effect。
// 用于支持多段结算（Times/IntervalMs）、以及未来的弹道/延迟命中等。
type ScheduledEffect struct {
	At     int64
	Stage  Stage
	Effect conf.EffectCfg
}

// RuntimeState 为技能运行时状态（是否正在吟唱/引导）。
type RuntimeState int32

const (
	// RuntimeState_Idle 空闲。
	RuntimeState_Idle RuntimeState = 0
	// RuntimeState_Casting 吟唱中。
	RuntimeState_Casting RuntimeState = 1
	// RuntimeState_Channeling 引导中。
	RuntimeState_Channeling RuntimeState = 2
)

// Skill 为技能运行时实例（每个单位、每个技能一份）。
// 注意：Skill 不负责具体伤害/治疗/加 Buff 的逻辑，只负责阶段推进与 Effect 调度；
// 具体结算通过 Update 的 exec 回调交给上层（SkillManager/战斗系统）。
type Skill struct {
	Cfg *conf.CSkill

	// CdEndAt/GcdEndAt 用于简单的 CD/GCD 判定（单位：毫秒时间戳）。
	CdEndAt  int64
	GcdEndAt int64

	// State/CastEndAt/ChannelEndAt 用于推进吟唱/引导流程。
	State        RuntimeState
	CastEndAt    int64
	ChannelEndAt int64

	// Ctx 为当前技能上下文；Pending 为待执行的 Effect 队列。
	Ctx     *SkillContext
	Pending []ScheduledEffect
}

// NewSkill 创建技能运行时实例。
func NewSkill(cfg *conf.CSkill) *Skill {
	return &Skill{Cfg: cfg}
}

// CanCast 判定当前是否允许施放（Idle + CD/GCD 到期）。
func (s *Skill) CanCast(now int64) bool {
	if s == nil || s.Cfg == nil {
		return false
	}
	if s.State != RuntimeState_Idle {
		return false
	}
	if s.GcdEndAt > now {
		return false
	}
	if s.CdEndAt > now {
		return false
	}
	return true
}

// StartCast 尝试开始施法。
// 成功后会触发 OnCastStart，并进入 Casting（若有吟唱）或直接 finishCast（瞬发）。
func (s *Skill) StartCast(now int64, ctx *SkillContext) bool {
	if !s.CanCast(now) {
		return false
	}

	s.Ctx = ctx

	gcdStartAt := s.Cfg.GcdStartAt
	if gcdStartAt == conf.TimingPoint_Invalid {
		gcdStartAt = conf.TimingPoint_CastStart
	}
	cdStartAt := s.Cfg.CooldownStartAt
	if cdStartAt == conf.TimingPoint_Invalid {
		cdStartAt = conf.TimingPoint_CastStart
	}

	if gcdStartAt == conf.TimingPoint_CastStart && s.Cfg.GcdMs > 0 {
		s.GcdEndAt = now + int64(s.Cfg.GcdMs)
	}
	if cdStartAt == conf.TimingPoint_CastStart && s.Cfg.CooldownMs > 0 {
		s.CdEndAt = now + int64(s.Cfg.CooldownMs)
	}

	s.scheduleList(Stage_CastStart, now, 0, s.Cfg.Effects.OnCastStart)

	if s.Cfg.CastTimeMs > 0 {
		s.State = RuntimeState_Casting
		s.CastEndAt = now + int64(s.Cfg.CastTimeMs)
		return true
	}

	s.finishCast(now)
	return true
}

// Cancel 取消/打断施法或引导。
func (s *Skill) Cancel(now int64) {
	if s == nil || s.Cfg == nil {
		return
	}
	if s.State == RuntimeState_Idle {
		return
	}

	s.State = RuntimeState_Idle
	s.CastEndAt = 0
	s.ChannelEndAt = 0
	s.Pending = nil

	s.scheduleList(Stage_Cancel, now, 0, s.Cfg.Effects.OnCancel)
}

// TriggerHit 外部命中事件入口（例如弹道系统回调）。
func (s *Skill) TriggerHit(now int64, ctx *SkillContext) {
	if s == nil || s.Cfg == nil {
		return
	}
	s.Ctx = ctx
	s.scheduleList(Stage_Hit, now, 0, s.Cfg.Effects.OnHit)
}

// Update 推进技能运行时，并在时间到达时执行 Pending 队列。
// exec 回调由上层实现，用来处理"实际结算"（伤害/治疗/施加 Buff 等）。
func (s *Skill) Update(now int64, exec func(Stage, conf.EffectCfg, *SkillContext)) {
	if s == nil || s.Cfg == nil {
		return
	}

	if s.State == RuntimeState_Casting && s.CastEndAt > 0 && now >= s.CastEndAt {
		s.finishCast(now)
	}
	if s.State == RuntimeState_Channeling && s.ChannelEndAt > 0 && now >= s.ChannelEndAt {
		s.State = RuntimeState_Idle
		s.ChannelEndAt = 0
	}

	if len(s.Pending) == 0 {
		return
	}

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
			exec(se.Stage, se.Effect, s.Ctx)
		}
		idx++
	}

	if idx > 0 {
		copy(s.Pending, s.Pending[idx:])
		s.Pending = s.Pending[:len(s.Pending)-idx]
	}
}

// finishCast 表示释放成功：触发 OnCastFinish，并在配置了引导时进入 Channeling。
func (s *Skill) finishCast(now int64) {
	s.State = RuntimeState_Idle
	s.CastEndAt = 0

	gcdStartAt := s.Cfg.GcdStartAt
	if gcdStartAt == conf.TimingPoint_Invalid {
		gcdStartAt = conf.TimingPoint_CastStart
	}
	cdStartAt := s.Cfg.CooldownStartAt
	if cdStartAt == conf.TimingPoint_Invalid {
		cdStartAt = conf.TimingPoint_CastStart
	}

	if gcdStartAt == conf.TimingPoint_CastFinish && s.Cfg.GcdMs > 0 {
		s.GcdEndAt = now + int64(s.Cfg.GcdMs)
	}
	if cdStartAt == conf.TimingPoint_CastFinish && s.Cfg.CooldownMs > 0 {
		s.CdEndAt = now + int64(s.Cfg.CooldownMs)
	}

	s.scheduleList(Stage_CastFinish, now, 0, s.Cfg.Effects.OnCastFinish)
	if s.Cfg.HitOnCastFinish {
		hitAt := now
		if s.Cfg.HitDelayMs > 0 {
			hitAt = now + int64(s.Cfg.HitDelayMs)
		}
		s.scheduleList(Stage_Hit, hitAt, 0, s.Cfg.Effects.OnHit)
	}

	if s.Cfg.ChannelTimeMs > 0 {
		s.State = RuntimeState_Channeling
		s.ChannelEndAt = now + int64(s.Cfg.ChannelTimeMs)
		startAt := now
		if s.Cfg.ChannelTickDelayMs > 0 {
			startAt = now + int64(s.Cfg.ChannelTickDelayMs)
		}
		s.scheduleList(Stage_Channel, startAt, s.ChannelEndAt, s.Cfg.Effects.OnChannelTick)
	}
}

// scheduleList 将某阶段的 EffectCfg 列表加入调度队列。
func (s *Skill) scheduleList(stage Stage, startAt int64, endAt int64, list []conf.EffectCfg) {
	for _, eff := range list {
		s.scheduleEffect(stage, startAt, endAt, eff)
	}
}

// scheduleEffect 将单个 EffectCfg 调度为 1 次或多次执行。
// 若 eff.Times > 1，则按 eff.IntervalMs 间隔追加多条 ScheduledEffect。
func (s *Skill) scheduleEffect(stage Stage, startAt int64, endAt int64, eff conf.EffectCfg) {
	times := eff.Times
	if times <= 1 {
		times = 1
		if stage == Stage_Channel && s.Cfg != nil && s.Cfg.ChannelTickMs > 0 && s.Cfg.ChannelTimeMs > 0 {
			tick := int64(s.Cfg.ChannelTickMs)
			total := int64(s.Cfg.ChannelTimeMs)
			times = int32((total + tick - 1) / tick)
			if times < 1 {
				times = 1
			}
		}
	}
	interval := eff.IntervalMs
	if interval < 0 {
		interval = 0
	}
	if interval == 0 && stage == Stage_Channel && s.Cfg != nil && s.Cfg.ChannelTickMs > 0 {
		interval = s.Cfg.ChannelTickMs
	}

	for i := int32(0); i < times; i++ {
		at := startAt + int64(i)*int64(interval)
		if endAt > 0 && at > endAt {
			break
		}
		s.Pending = append(s.Pending, ScheduledEffect{At: at, Stage: stage, Effect: eff})
	}
}
