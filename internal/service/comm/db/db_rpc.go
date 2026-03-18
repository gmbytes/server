package db

import (
	"database/sql"
	"errors"
	"fmt"
	"server/internal/data"
	"server/pkg/utils"

	"github.com/gmbytes/snow/pkg/task"
	"github.com/gmbytes/snow/routines/node"
	"github.com/lib/pq"
)

func (ss *DB) RpcGetRoles(ctx node.IRpcContext, userid string, sid int64) {
	nowt := ss.GetSecond()

	res, err := ss.db.Exec(`insert into users values($1, $2) on conflict (userid) do nothing`, userid, nowt)
	if err != nil {
		ss.Fatalf("%v", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		ss.Fatalf("%v", err)
	}

	ret := &data.UserRoles{
		NewUser: n != 0,
	}
	rows, err := ss.db.Query(
		`select data->'Show', coalesce((data->'Base'->>'LogoutTs')::bigint, 0) from role`+
			` where userid = $1 and del = 0 and position($2 in name) = 1;`,
		userid, utils.Itoa(sid),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Return(ret)
			return
		}
		ss.Fatalf("%v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var d []byte
		var logoutTs int64
		if err := rows.Scan(&d, &logoutTs); err != nil {
			ss.Fatalf("%v", err)
		}
		ret.Roles = append(ret.Roles, &data.RoleShow{
			Data:     d,
			LogoutTs: logoutTs,
		})
	}
	if err := rows.Err(); err != nil {
		ss.Fatalf("%v", err)
	}
	ctx.Return(ret)
}

// 必须有 userid，用于校验 id 合法性
func (ss *DB) RpcGetRoleData(ctx node.IRpcContext, userid string, id int64) {
	if info, ok := ss.roleCache[id]; ok {
		if info.UserId != userid {
			ctx.Return(false, "{}")
			return
		}
		ctx.Return(true, info.Role)
		return
	}
	var dt string
	row := ss.db.QueryRow(`select data from role where userid = $1 and id = $2;`, userid, id)
	if err := row.Scan(&dt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Return(false, "{}")
			return
		}
		ss.Fatalf("sql error: %+v", err)
	}
	ss.addRoleCache(id, &RoleCacheInfo{
		Role:   dt,
		UserId: userid,
		Time:   ss.GetSecond(),
	})

	ctx.Return(true, dt)
}

func (ss *DB) RpcInsertRoleData(ctx node.IRpcContext, userid string, id int64, name, pname string, data string) {
	_, err := ss.db.Exec(
		`insert into role (id, name, pname, userid, del, data, version) values($1, $2, $3, $4, $5, $6, $7);`,
		id, name, pname, userid, 0, data, 0,
	)
	if err != nil {
		ss.Debugf("InsertRoleData sql error: %+v", err)
		ctx.Return(false)
	} else {
		ss.addRoleCache(id, &RoleCacheInfo{Role: data, UserId: userid, Time: ss.GetSecond()})
		ctx.Return(true)
	}
}

func (ss *DB) RpcUpdateRoleName(ctx node.IRpcContext, id int64, name, pname string) {
	ss.waiter.Add(1)
	task.Execute(func() {
		_, err := ss.db.Exec(
			`update role set name = $2, pname = $3  where id = $1 ;`,
			id, name, pname,
		)
		if err != nil {
			ss.Debugf("UpdateRoleName sql error: %+v", err)
			ctx.Return(false)
		} else {
			ctx.Return(true)
		}
		ss.waiter.Done()
	})

}

func (ss *DB) RpcUpdateRoleData(ctx node.IRpcContext, version int64, userid string, id int64, data string, detail string) {
	ss.addRoleCache(id, &RoleCacheInfo{Role: data, UserId: userid, Time: ss.GetSecond()})
	ss.waiter.Add(1)
	task.Execute(func() {
		_, err := ss.db.Exec(
			`update role set version = $4, data = $2, detail = $3 where id = $1 and version < $4;`,
			id, data, detail, version,
		)
		if err != nil {
			// 因为在协程里面，不能崩
			ss.Errorf("sql error: %+v", err)
		}
		ss.waiter.Done()
		ctx.Return()
	})
}

func (ss *DB) RpcGetAppData(ctx node.IRpcContext, id string) {
	var dt string
	row := ss.db.QueryRow(`select data from app where appid = $1;`, id)
	if err := row.Scan(&dt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Return("{}")
			return
		}
		ss.Fatalf("sql error: %+v", err)
	}
	ctx.Return(dt)
}

func (ss *DB) RpcSetAppData(ctx node.IRpcContext, id, dt string) {
	ss.waiter.Add(1)
	task.Execute(func() {
		_, err := ss.db.Exec(
			`insert into app (appid, data) values($1, $2)
			on conflict (appid) do update set data = $2;`,
			id, dt,
		)
		if err != nil {
			ss.Errorf("sql error: %+v", err)
		}
		ss.waiter.Done()
		ctx.Return()
	})
}

func (ss *DB) RpcGetMultiAppData(ctx node.IRpcContext, ids []string) {
	var dts []string
	for _, id := range ids {
		var dt string
		row := ss.db.QueryRow(`select data from app where appid = $1;`, id)
		if err := row.Scan(&dt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				dts = append(dts, "{}")
				continue
			}
			ss.Fatalf("sql error: %+v", err)
		}
		dts = append(dts, dt)
	}
	ctx.Return(dts)
}

func (ss *DB) RpcSetMultiAppData(ctx node.IRpcContext, dts map[string]string) {
	ss.waiter.Add(1)
	task.Execute(func() {
		for appid, dt := range dts {
			_, err := ss.db.Exec(`insert into app (appid, data) values($1, $2) on conflict (appid) do update set data = $2;`, appid, dt)
			if err != nil {
				ss.Errorf("sql error: %+v", err)
			}
		}
		ss.waiter.Done()
		ctx.Return()
	})
}

func (ss *DB) RpcDelRoleData(ctx node.IRpcContext, id int64) {
	_, err := ss.db.Exec(`update role set del = $2 where id = $1;`, id, ss.GetSecond())
	if err != nil {
		ss.Fatalf("sql error: %+v", err)
		ctx.Return(false)
		return
	}
	ctx.Return(true)
}

func (ss *DB) RpcGetTableData(ctx node.IRpcContext, tb string) {
	rets := make(map[int64]string)
	tb = pq.QuoteIdentifier(tb)
	rows, err := ss.db.Query(fmt.Sprintf("select * from %s ", tb))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Return(rets)
			return
		}
		ss.Fatalf("%v", err)
	}

	defer rows.Close()
	for rows.Next() {
		var dt string
		var id int64
		if err := rows.Scan(&id, &dt); err != nil {
			ss.Fatalf("%v", err)
		}
		rets[id] = dt
	}
	if err := rows.Err(); err != nil {
		ss.Fatalf("%v", err)
	}
	ctx.Return(rets)
}

func (ss *DB) RpcSetTableData(ctx node.IRpcContext, tb string, tbDts map[int64][]byte) {
	ss.waiter.Add(1)
	task.Execute(func() {
		tb = pq.QuoteIdentifier(tb)
		for id, dt := range tbDts {
			_, err := ss.db.Exec(fmt.Sprintf("insert into %s (id,data) values($1,$2) on conflict (id) do update set data = $2;", tb),
				id, string(dt))
			if err != nil {
				ss.Errorf("set %s (id:%d ) data, sql error: %+v", tb, id, err)
			}
		}
		ss.waiter.Done()
		ctx.Return()
	})
}

func (ss *DB) RpcSetMultiTableData(ctx node.IRpcContext, tbDts map[string]map[int64][]byte) {
	ss.waiter.Add(1)
	task.Execute(func() {
		for tb, dts := range tbDts {
			tb = pq.QuoteIdentifier(tb)
			for id, dt := range dts {
				_, err := ss.db.Exec(fmt.Sprintf("insert into %s (id,data) values($1,$2) on conflict (id) do update set data = $2;", tb),
					id, string(dt))
				if err != nil {
					ss.Errorf("set %s (id:%d ) data, sql error: %+v", tb, id, err)
				}
			}
		}
		ss.waiter.Done()
		ctx.Return()
	})
}

func (ss *DB) RpcDelTableDataById(ctx node.IRpcContext, tb string, id int64) {
	tb = pq.QuoteIdentifier(tb)
	_, err := ss.db.Exec(fmt.Sprintf("delete from %s where id = $1;", tb), id)
	if err != nil {
		ss.Fatalf("sql error: %+v", err)
		ctx.Return(false)
		return
	}
	ctx.Return(true)
}

func (ss *DB) RpcShutdown(ctx node.IRpcContext) {
	ss.waiter.Wait()
	ctx.Return()
}
