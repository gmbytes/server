package robot

import (
	"context"
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"errors"
	"io"
	"math/rand/v2"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gmbytes/snow/core/encrypt/dh"
	"github.com/gmbytes/snow/routines/node"
	"google.golang.org/protobuf/proto"

	"server/internal/pb"
)

func init() {
	node.Register[Robot, *Robot]("Robot")
}

type Option struct {
	Name             string `snow:"Name"`
	GateHost         string `snow:"GateHost"`
	GatePort         int    `snow:"GatePort"`
	ServerId         int64  `snow:"ServerId"`
	Account          string `snow:"Account"`
	Username         string `snow:"Username"`
	ChannelId        string `snow:"ChannelId"`
	Platform         string `snow:"Platform"`
	Appid            string `snow:"Appid"`
	Version          string `snow:"Version"`
	ReconnectDelayMs int    `snow:"ReconnectDelayMs"`
	PingIntervalMs   int    `snow:"PingIntervalMs"`
	DialTimeoutMs    int    `snow:"DialTimeoutMs"`
}

type Robot struct {
	node.Service

	lock        sync.Mutex
	opt         Option
	ctx         context.Context
	cancel      context.CancelFunc
	waiter      sync.WaitGroup
	conn        net.Conn
	connId      uint64
	readCipher  *rc4.Cipher
	writeCipher *rc4.Cipher
	running     atomic.Bool
	serial      uint32
}

func (ss *Robot) Start(args any) {
	if !ss.running.CompareAndSwap(false, true) {
		return
	}
	ss.EnableRpc()

	opt := defaultOption()
	switch v := args.(type) {
	case *Option:
		if v != nil {
			opt = mergeOption(*v)
		}
	case Option:
		opt = mergeOption(v)
	}

	ctx, cancel := context.WithCancel(context.Background())

	ss.lock.Lock()
	ss.opt = opt
	ss.ctx = ctx
	ss.cancel = cancel
	ss.serial = 0
	ss.lock.Unlock()

	ss.waiter.Add(1)
	go ss.run()
}

func (ss *Robot) RpcApplyOption(ctx node.IRpcContext, opt Option) {
	ss.lock.Lock()
	ss.opt = mergeOption(opt)
	ss.lock.Unlock()
	ctx.Return(true)
}

func (ss *Robot) Stop(wg *sync.WaitGroup) {
	if !ss.running.CompareAndSwap(true, false) {
		return
	}

	ss.lock.Lock()
	cancel := ss.cancel
	ss.lock.Unlock()

	if cancel != nil {
		cancel()
	}
	ss.closeConn()

	if wg != nil {
		wg.Add(1)
		go func() {
			ss.waiter.Wait()
			wg.Done()
		}()
		return
	}

	ss.waiter.Wait()
}

func (ss *Robot) run() {
	defer ss.waiter.Done()

	for {
		if ss.done() {
			return
		}

		if err := ss.connectAndServe(); err != nil && ss.done() {
			return
		}

		if ss.done() {
			return
		}

		delay := time.Duration(ss.opt.ReconnectDelayMs) * time.Millisecond
		if delay <= 0 {
			delay = 2 * time.Second
		}

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ss.ctx.Done():
		}
		timer.Stop()
	}
}

func (ss *Robot) connectAndServe() error {
	address := net.JoinHostPort(ss.opt.GateHost, strconv.Itoa(ss.opt.GatePort))
	conn, err := net.DialTimeout("tcp4", address, time.Duration(ss.opt.DialTimeoutMs)*time.Millisecond)
	if err != nil {
		return err
	}

	if err = ss.handshake(conn); err != nil {
		_ = conn.Close()
		return err
	}

	ss.lock.Lock()
	ss.conn = conn
	ss.lock.Unlock()

	if err = ss.sendLogin(); err != nil {
		ss.closeConn()
		return err
	}

	readErrChan := make(chan error, 1)
	go func() {
		readErrChan <- ss.readLoop()
	}()

	pingTicker := time.NewTicker(time.Duration(ss.opt.PingIntervalMs) * time.Millisecond)
	defer pingTicker.Stop()

	for {
		select {
		case <-ss.ctx.Done():
			ss.closeConn()
			return nil
		case err = <-readErrChan:
			ss.closeConn()
			return err
		case <-pingTicker.C:
			if err = ss.sendPing(); err != nil {
				ss.closeConn()
				return err
			}
		}
	}
}

func (ss *Robot) handshake(conn net.Conn) error {
	var privateKey uint64
	for privateKey == 0 {
		privateKey = rand.Uint64()
	}
	publicKey := dh.PublicKeyOf(privateKey)

	buf8 := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf8, publicKey)
	if err := writeAll(conn, buf8); err != nil {
		return err
	}

	resp := make([]byte, 24)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return err
	}

	ss.connId = binary.LittleEndian.Uint64(resp[0:8])
	challenge := binary.LittleEndian.Uint64(resp[8:16])
	serverPublicKey := binary.LittleEndian.Uint64(resp[16:24])
	secret := dh.LocalKey(privateKey, serverPublicKey)

	md5Buffer := make([]byte, 16)
	binary.LittleEndian.PutUint64(md5Buffer[0:8], secret)
	binary.LittleEndian.PutUint64(md5Buffer[8:16], challenge)
	md5sum := md5.Sum(md5Buffer)
	if err := writeAll(conn, md5sum[:]); err != nil {
		return err
	}

	rc4Key := make([]byte, 8)
	binary.LittleEndian.PutUint64(rc4Key, secret)
	var err error
	ss.readCipher, err = rc4.NewCipher(rc4Key)
	if err != nil {
		return err
	}
	ss.writeCipher, err = rc4.NewCipher(rc4Key)
	if err != nil {
		return err
	}

	return nil
}

func (ss *Robot) sendLogin() error {
	req := &pb.ReqLogin{
		Fast:      false,
		Timestamp: time.Now().UnixMilli(),
		SId:       ss.opt.ServerId,
		Account:   ss.opt.Account,
		Username:  ss.opt.Username,
		ChannelId: ss.opt.ChannelId,
		Platform:  ss.opt.Platform,
		Appid:     ss.opt.Appid,
		Version:   ss.opt.Version,
	}
	return ss.sendReq(pb.EKey_ReqLogin, req)
}

func (ss *Robot) sendPing() error {
	return ss.sendReq(pb.EKey_ReqPing, &pb.ReqPing{})
}

func (ss *Robot) sendReq(key pb.EKey_T, msg proto.Message) error {
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	serial := atomic.AddUint32(&ss.serial, 1)
	payload := make([]byte, 10+len(body))
	binary.LittleEndian.PutUint16(payload[0:2], uint16(key))
	binary.LittleEndian.PutUint32(payload[2:6], uint32(4+len(body)))
	binary.LittleEndian.PutUint32(payload[6:10], serial)
	copy(payload[10:], body)

	ss.writeCipher.XORKeyStream(payload, payload)
	return writeAll(ss.conn, payload)
}

func (ss *Robot) readLoop() error {
	for {
		header := make([]byte, 4)
		if _, err := io.ReadFull(ss.conn, header); err != nil {
			return err
		}
		ss.readCipher.XORKeyStream(header, header)

		key := pb.EKey_T(binary.LittleEndian.Uint16(header[0:2]))
		errCode := pb.EErrorCode_T(binary.LittleEndian.Uint16(header[2:4]))
		if errCode != pb.EErrorCode_Ok {
			if key == pb.EKey_DspKickRole {
				return errors.New("kicked")
			}
			continue
		}

		lengthBuf := make([]byte, 4)
		if _, err := io.ReadFull(ss.conn, lengthBuf); err != nil {
			return err
		}
		ss.readCipher.XORKeyStream(lengthBuf, lengthBuf)

		bodyLen := binary.LittleEndian.Uint32(lengthBuf)
		if bodyLen == 0 {
			continue
		}

		body := make([]byte, bodyLen)
		if _, err := io.ReadFull(ss.conn, body); err != nil {
			return err
		}
		ss.readCipher.XORKeyStream(body, body)

		_ = pb.Unmarshal(key, body)
	}
}

func (ss *Robot) closeConn() {
	ss.lock.Lock()
	conn := ss.conn
	ss.conn = nil
	ss.lock.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
}

func (ss *Robot) done() bool {
	ss.lock.Lock()
	ctx := ss.ctx
	ss.lock.Unlock()
	if ctx == nil {
		return true
	}
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func defaultOption() Option {
	return Option{
		Name:             "robot",
		GateHost:         "127.0.0.1",
		GatePort:         61101,
		ServerId:         1,
		Account:          "robot_0",
		Username:         "robot_0",
		ChannelId:        "robot",
		Platform:         "robot",
		Appid:            "robot",
		Version:          "0.0.1",
		ReconnectDelayMs: 2000,
		PingIntervalMs:   5000,
		DialTimeoutMs:    3000,
	}
}

func mergeOption(opt Option) Option {
	ret := defaultOption()

	if opt.Name != "" {
		ret.Name = opt.Name
	}
	if opt.GateHost != "" {
		ret.GateHost = opt.GateHost
	}
	if opt.GatePort > 0 {
		ret.GatePort = opt.GatePort
	}
	if opt.ServerId > 0 {
		ret.ServerId = opt.ServerId
	}
	if opt.Account != "" {
		ret.Account = opt.Account
	}
	if opt.Username != "" {
		ret.Username = opt.Username
	}
	if opt.ChannelId != "" {
		ret.ChannelId = opt.ChannelId
	}
	if opt.Platform != "" {
		ret.Platform = opt.Platform
	}
	if opt.Appid != "" {
		ret.Appid = opt.Appid
	}
	if opt.Version != "" {
		ret.Version = opt.Version
	}
	if opt.ReconnectDelayMs > 0 {
		ret.ReconnectDelayMs = opt.ReconnectDelayMs
	}
	if opt.PingIntervalMs > 0 {
		ret.PingIntervalMs = opt.PingIntervalMs
	}
	if opt.DialTimeoutMs > 0 {
		ret.DialTimeoutMs = opt.DialTimeoutMs
	}

	return ret
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
