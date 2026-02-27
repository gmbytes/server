package data

type UserRoles struct {
	NewUser bool
	Roles   []*RoleShow
}

type RoleShow struct {
	Data     []byte // 角色数据
	LogoutTs int64  // 下线时间
}
