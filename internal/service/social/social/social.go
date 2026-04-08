package social

import (
	"sync"

	"github.com/gmbytes/snow/pkg/routines/node"
)

func init() {
	node.Register[Social, *Social]("Social")
}

type Social struct {
	node.Service

	sDB node.IProxy
}

func (ss *Social) Start(_ any) {
	ss.sDB = ss.CreateProxy("DB")
	ss.EnableRpc()
	ss.Infof("Social service started")
}

func (ss *Social) Stop(_ *sync.WaitGroup) {}

func (ss *Social) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{"status": "Social.OK"})
}

func (ss *Social) RpcAddFriend(ctx node.IRpcContext, roleId int64, targetRoleId int64) {
	ctx.Return(true)
}

func (ss *Social) RpcRemoveFriend(ctx node.IRpcContext, roleId int64, targetRoleId int64) {
	ctx.Return(true)
}

func (ss *Social) RpcGetFriends(ctx node.IRpcContext, roleId int64) {
	ctx.Return([]int64{})
}

func (ss *Social) RpcBroadcastChat(ctx node.IRpcContext, channel int, roleId int64, content string) {
	ctx.Return(true)
}

func (ss *Social) RpcSendMail(ctx node.IRpcContext, fromRoleId int64, toRoleId int64, title string, content string) {
	ctx.Return(true)
}

func (ss *Social) RpcGetMails(ctx node.IRpcContext, roleId int64) {
	ctx.Return([]any{})
}
