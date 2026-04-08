package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gmbytes/snow/pkg/host"
	"github.com/gmbytes/snow/pkg/option"
	"github.com/gmbytes/snow/pkg/routines/node"
	"github.com/lib/pq"
)

func init() {
	node.Register[DB, *DB]("DB", func(b host.IBuilder) {
		host.AddOption[*Option](b, "DB")
	})
}

type DB struct {
	node.Service

	opt       *Option
	db        *sql.DB
	roleCache map[int64]*RoleCacheInfo
	waiter    *sync.WaitGroup
}

func (ss *DB) ConstructDB(opt *option.Option[*Option]) {
	ss.opt = opt.Get()
	opt.OnChanged(func() {
		newOpt := opt.Get()
		ss.opt = newOpt
	})
}

func (ss *DB) Start(arg interface{}) {
	ss.roleCache = make(map[int64]*RoleCacheInfo)

	if opt, ok := arg.(*Option); ok && opt != nil {
		ss.opt = opt
	}

	if ss.opt == nil {
		ss.Fatalf("db start failed: option is nil, please check DB config and host options binding")
		return
	}

	dbURL := ss.getDatabaseURL(
		ss.opt.Host,
		ss.opt.Port,
		ss.opt.Name,
		ss.opt.User,
		ss.opt.Password,
	)
	resetDB := ss.opt.Reset

	for resetDB {
		ss.Infof("try to retset db '%s'...", ss.opt.Name)
		if ss.dropDB() {
			ss.Infof("reset db '%s' success", ss.opt.Name)
			break
		}

		time.Sleep(2 * time.Second)
		ss.Infof("try again...")
	}

	var err error
	ss.db, err = sql.Open("postgres", dbURL)
	if err != nil {
		ss.Fatalf("open db '%v' failed: %v", ss.opt.Name, err)
	}

	for {
		if err := ss.db.Ping(); err != nil {
			if e, ok := err.(*pq.Error); ok && e.Code == "3D000" {
				ss.Infof("db '%v' does not exist, try to create one...", ss.opt.Name)
				if ss.createDB() {
					ss.Infof("create db '%s' success", ss.opt.Name)
					continue
				}
			} else {
				ss.Infof("connect to db '%v' failed: %v", ss.opt.Name, err)
			}

			time.Sleep(2 * time.Second)
			ss.Infof("try again...")
			continue
		}
		break
	}

	ss.db.SetMaxOpenConns(20)
	ss.initDB()

	ss.EnableRpc()

	ss.Tick(30*10, 0, ss.updateCache)
}

func (ss *DB) Stop(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	if err := ss.db.Close(); err != nil {
		ss.Warnf("close db failed: %v", err)
	}

}

func (ss *DB) getDatabaseURL(host string, port int, name, user, passwd string) string {
	if len(passwd) == 0 {
		return fmt.Sprintf(
			`host=%s port=%d dbname=%s user=%s sslmode=disable`,
			host, port, name, user,
		)
	}
	return fmt.Sprintf(
		`host=%s port=%d dbname=%s user=%s password=%s sslmode=disable`,
		host, port, name, user, passwd,
	)
}

func (ss *DB) getTableName(sql string) string {
	return strings.FieldsFunc(strings.TrimSpace(sql), func(r rune) bool {
		return unicode.IsSpace(r) || r == '('
	})[5]
}

func (ss *DB) initTables(sqls []string) {
	for _, sql := range sqls {
		_, err := ss.db.Exec(sql)
		if err != nil {
			ss.Fatalf("db sql(%s) error: %+v", sql, err)
		}
		name := ss.getTableName(sql)
		ss.Infof("init table '%v' ok", name)
	}
}

func (ss *DB) createDB() bool {
	dbURL := ss.getDatabaseURL(
		ss.opt.Host,
		ss.opt.Port,
		"postgres",
		"postgres",
		ss.opt.AdminPassword,
	)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		ss.Warnf("open postgres failed: %v", err)
		return false
	}

	createSql := fmt.Sprintf("create database %s owner %s", ss.opt.Name, ss.opt.User)
	if _, err := db.Exec(createSql); err != nil {
		ss.Warnf("create db '%v' for '%v' failed: %v", ss.opt.Name, ss.opt.User, err)
		return false
	}

	db.Close()
	return true
}

func (ss *DB) dropDB() bool {
	dbURL := ss.getDatabaseURL(
		ss.opt.Host,
		ss.opt.Port,
		"postgres",
		"postgres",
		ss.opt.AdminPassword,
	)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		ss.Warnf("open postgres failed: %v", err)
		return false
	}

	createDB := fmt.Sprintf("drop database if exists %s", ss.opt.Name)
	if _, err := db.Exec(createDB); err != nil {
		ss.Warnf("drop db '%v' failed: %v", ss.opt.Name, err)
		return false
	}

	db.Close()
	return true
}

func (ss *DB) initDB() {
	ss.waiter = &sync.WaitGroup{}

	{
		sqls := []string{
			`create table if not exists app (
				appid			varchar(128) primary key,
				data			jsonb
			);`,
		}
		ss.Infof("-- init common tables...")
		ss.initTables(sqls)
	}

	// 游戏服
	if node.Config.CurNodeMap["Game"] {
		sqls := []string{
			`create table if not exists users (
				userid			varchar(128) primary key,
				time			bigint not null
			);`,

			`create table if not exists role (
				id				bigint primary key,
				name			varchar(128) unique not null,
				pname 			varchar(128) not null,
				userid			varchar(128) not null,
				version			bigint not null,
				del				bigint not null,
				data			jsonb,
				detail			jsonb
			);
			create index if not exists index_role_userid on role (userid);
			create index if not exists index_role_pname on role (pname);`,
		}

		ss.Infof("-- init game tables...")
		ss.initTables(sqls)
	}

}

func (ss *DB) updateCache() {
	cur := ss.GetSecond()
	for id, v := range ss.roleCache {
		if cur-v.Time > DURATION {
			delete(ss.roleCache, id)
		}
	}
}

func (ss *DB) addRoleCache(id int64, cache *RoleCacheInfo) {
	ss.roleCache[id] = cache
}
