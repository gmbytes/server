package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gmbytes/snow/pkg/routines/node"
)

func init() {
	node.Register[Auth, *Auth]("Auth")
}

type Auth struct {
	node.Service

	secret []byte
}

func (ss *Auth) Start(_ any) {
	ss.secret = make([]byte, 32)
	_, _ = rand.Read(ss.secret)
	ss.EnableRpc()
	ss.Infof("Auth service started")
}

func (ss *Auth) Stop(_ *sync.WaitGroup) {}

func (ss *Auth) RpcStatus(ctx node.IRpcContext) {
	ctx.Return(map[string]any{"status": "Auth.OK"})
}

func (ss *Auth) RpcHashPassword(_ node.IRpcContext, raw string) {
	// placeholder: production should use bcrypt/argon2
}

func (ss *Auth) RpcVerifyPassword(ctx node.IRpcContext, raw string, hashed string) {
	h := sha256.Sum256([]byte(raw))
	ctx.Return(hex.EncodeToString(h[:]) == hashed)
}

func (ss *Auth) RpcGenAccessToken(ctx node.IRpcContext, account string, device string) {
	payload := fmt.Sprintf("at|%s|%s|%d", account, device, time.Now().UnixMilli())
	sig := ss.sign(payload)
	token := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
	ctx.Return(token)
}

func (ss *Auth) RpcVerifyAccessToken(ctx node.IRpcContext, token string) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		ctx.Return(false, "", "")
		return
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		ctx.Return(false, "", "")
		return
	}
	payload := string(payloadBytes)
	if ss.sign(payload) != parts[1] {
		ctx.Return(false, "", "")
		return
	}
	fields := strings.Split(payload, "|")
	if len(fields) < 4 || fields[0] != "at" {
		ctx.Return(false, "", "")
		return
	}
	ctx.Return(true, fields[1], fields[2])
}

func (ss *Auth) RpcGenGateTicket(ctx node.IRpcContext, account string, realmId int64, roleId int64) {
	payload := fmt.Sprintf("gt|%s|%d|%d|%d", account, realmId, roleId, time.Now().UnixMilli())
	sig := ss.sign(payload)
	ticket := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
	ctx.Return(ticket)
}

// RpcGenReconnectTicket 签发断线重连票据（HTTP /auth/reconnect 使用，与 gate ticket 分离）。
func (ss *Auth) RpcGenReconnectTicket(ctx node.IRpcContext, account string, realmId int64, roleId int64) {
	payload := fmt.Sprintf("rt|%s|%d|%d|%d", account, realmId, roleId, time.Now().UnixMilli())
	sig := ss.sign(payload)
	ticket := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
	ctx.Return(ticket)
}

// RpcVerifyReconnectTicket 校验重连票据，返回账号与区服、角色 ID。
func (ss *Auth) RpcVerifyReconnectTicket(ctx node.IRpcContext, ticket string) {
	parts := strings.SplitN(ticket, ".", 2)
	if len(parts) != 2 {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	payload := string(payloadBytes)
	if ss.sign(payload) != parts[1] {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	fields := strings.Split(payload, "|")
	if len(fields) < 5 || fields[0] != "rt" {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	var realmId, roleId int64
	_, _ = fmt.Sscanf(fields[2], "%d", &realmId)
	_, _ = fmt.Sscanf(fields[3], "%d", &roleId)
	ctx.Return(true, fields[1], realmId, roleId)
}

func (ss *Auth) RpcVerifyGateTicket(ctx node.IRpcContext, ticket string) {
	parts := strings.SplitN(ticket, ".", 2)
	if len(parts) != 2 {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	payload := string(payloadBytes)
	if ss.sign(payload) != parts[1] {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	fields := strings.Split(payload, "|")
	if len(fields) < 5 || fields[0] != "gt" {
		ctx.Return(false, "", int64(0), int64(0))
		return
	}
	var realmId, roleId int64
	_, _ = fmt.Sscanf(fields[2], "%d", &realmId)
	_, _ = fmt.Sscanf(fields[3], "%d", &roleId)
	ctx.Return(true, fields[1], realmId, roleId)
}

func (ss *Auth) sign(payload string) string {
	mac := hmac.New(sha256.New, ss.secret)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
