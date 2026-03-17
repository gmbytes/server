package skill

import "server/internal/pb"

// CastSkillReq 是技能释放的输入参数（由外层协议适配后填充）。
// 该模块不直接依赖具体的网络协议消息结构，避免 pb 变更导致核心战斗逻辑无法编译。
type CastSkillReq struct {
	LockTarget int64
	Pos        *pb.Vector
}

