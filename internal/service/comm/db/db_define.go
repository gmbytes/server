package db

import (
	"time"
)

type Option struct {
	Reset         bool   // 是否重置数据库
	Host          string // 数据库地址
	Port          int    // 数据库端口
	Name          string // 数据库名字
	User          string // 数据库用户名
	Password      string // 数据库密码
	AdminPassword string // 游戏数据库管理员密码
}

type AccountInfo struct {
	UpdateTime time.Time
}

const DURATION = int64(60)

type RoleCacheInfo struct {
	UserId string
	Role   string
	Time   int64
}
