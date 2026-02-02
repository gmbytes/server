package skill

import (
	"server/lib/uid"
	"server/pb"
	"server/service/scene/score"
	"time"
)

type SkillContext struct {
	id            uid.Uid // 给ctx 一个唯一id 方便清理ctx
	destructionMs int64   // 对象销毁时间
	isReset       bool

	Scene score.IScene  // 当前场景
	Owner score.IEntity // 技能拥有者

	Req        *pb.ReqCastSkill // 技能请求
	SkillLevel int64            // 技能等级
	IsFinished bool             // 技能已结束
}

func NewSkillContext(owner score.IEntity, req *pb.ReqCastSkill, skillLevel int64) *SkillContext {
	ctx := &SkillContext{
		id:         uid.Gen(),
		Owner:      owner,
		Req:        req,
		SkillLevel: skillLevel,
		IsFinished: false,
	}
	if owner != nil {
		ctx.Scene = owner.GetScene()
	}
	return ctx
}

func (c *SkillContext) Finish() {
	c.IsFinished = true
}

func (c *SkillContext) GetId() uid.Uid {
	return c.id
}

type skillEffect interface {
	Begin(ctx *SkillContext, causer score.IEntity, targets []score.IEntity)
	Update(ctx *SkillContext, delta time.Duration)
	End(ctx *SkillContext)
	Revert(ctx *SkillContext)
}
