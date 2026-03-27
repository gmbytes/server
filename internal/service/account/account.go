package account

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"server/internal/data"
	"server/internal/pb"
	"server/pkg/uid"

	"github.com/gmbytes/snow/routines/node"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	node.Register[Account, *Account]("Account")
}

type Account struct {
	node.Service

	sDB node.IProxy
}

func (ss *Account) Start(_ any) {
	ss.sDB = ss.CreateProxy("DB")
	ss.EnableRpc()
}

func (ss *Account) Stop(_ *sync.WaitGroup) {}

func (ss *Account) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{
		"status":  "Account.OK",
		"dbAvail": ss.sDB != nil && ss.sDB.Avail(),
	})
}

func (ss *Account) RpcCreateAccount(ctx node.IRpcContext, account string, hashedPwd string, platform string) {
	account = strings.TrimSpace(account)
	if account == "" {
		ctx.Error(fmt.Errorf("account is empty"))
		return
	}
	ctx.Return(true)
}

func (ss *Account) RpcGetAccount(ctx node.IRpcContext, account string) {
	account = strings.TrimSpace(account)
	if account == "" {
		ctx.Error(fmt.Errorf("account is empty"))
		return
	}
	ctx.Return(account)
}

func (ss *Account) RpcGetRoles(ctx node.IRpcContext, account string) {
	account = strings.TrimSpace(account)
	if account == "" {
		ctx.Error(fmt.Errorf("account is empty"))
		return
	}
	if ss.sDB == nil || !ss.sDB.Avail() {
		ctx.Error(fmt.Errorf("db proxy unavailable"))
		return
	}

	ss.sDB.Call("GetRoles", account, int64(0)).
		Then(func(ret *data.UserRoles) {
			roles, err := decodeRoleSummaries(ret)
			if err != nil {
				ctx.Error(err)
				return
			}
			ctx.Return(roles)
		}).
		Catch(func(err error) {
			ctx.Error(err)
		}).
		Done()
}

func (ss *Account) RpcGetRole(ctx node.IRpcContext, account string, roleID int64) {
	account = strings.TrimSpace(account)
	if account == "" {
		ctx.Error(fmt.Errorf("account is empty"))
		return
	}
	if roleID == 0 {
		ctx.Error(fmt.Errorf("role id is empty"))
		return
	}
	if ss.sDB == nil || !ss.sDB.Avail() {
		ctx.Error(fmt.Errorf("db proxy unavailable"))
		return
	}

	ss.sDB.Call("GetRoleData", account, roleID).
		Then(func(ok bool, raw string) {
			if !ok {
				ctx.Error(fmt.Errorf("role %d not found", roleID))
				return
			}
			role := &pb.RoleSummaryData{}
			if err := protojson.Unmarshal([]byte(raw), role); err != nil {
				ctx.Error(fmt.Errorf("decode role %d failed: %w", roleID, err))
				return
			}
			ctx.Return(role)
		}).
		Catch(func(err error) {
			ctx.Error(err)
		}).
		Done()
}

func (ss *Account) RpcCreateRole(ctx node.IRpcContext, account string, cid int64, name string) {
	account = strings.TrimSpace(account)
	name = strings.TrimSpace(name)
	if account == "" {
		ctx.Error(fmt.Errorf("account is empty"))
		return
	}
	if cid == 0 {
		ctx.Error(fmt.Errorf("cid is empty"))
		return
	}
	if name == "" {
		ctx.Error(fmt.Errorf("name is empty"))
		return
	}
	if ss.sDB == nil || !ss.sDB.Avail() {
		ctx.Error(fmt.Errorf("db proxy unavailable"))
		return
	}

	nowMs := time.Now().UnixMilli()
	role := &pb.RoleSummaryData{
		Id:       int64(uid.Gen()),
		Cid:      cid,
		Lv:       1,
		Name:     name,
		CreateTs: nowMs,
	}
	raw, err := protojson.Marshal(role)
	if err != nil {
		ctx.Error(fmt.Errorf("encode role failed: %w", err))
		return
	}

	ss.sDB.Call("InsertRoleData", account, role.Id, name, name, string(raw)).
		Then(func(ok bool) {
			if !ok {
				ctx.Error(fmt.Errorf("create role failed"))
				return
			}
			ctx.Return(role)
		}).
		Catch(func(err error) {
			ctx.Error(err)
		}).
		Done()
}

func decodeRoleSummaries(ret *data.UserRoles) ([]*pb.RoleSummaryData, error) {
	if ret == nil || len(ret.Roles) == 0 {
		return []*pb.RoleSummaryData{}, nil
	}
	roles := make([]*pb.RoleSummaryData, 0, len(ret.Roles))
	for _, item := range ret.Roles {
		if item == nil || len(item.Data) == 0 {
			continue
		}
		role := &pb.RoleSummaryData{}
		if err := protojson.Unmarshal(item.Data, role); err != nil {
			return nil, fmt.Errorf("decode role summary failed: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}
