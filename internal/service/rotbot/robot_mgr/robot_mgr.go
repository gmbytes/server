package robot_mgr

import (
	"fmt"
	"sync"

	"server/internal/service/rotbot/robot"

	"github.com/gmbytes/snow/pkg/host"
	"github.com/gmbytes/snow/routines/node"
)

func init() {
	node.Register[RobotManager, *RobotManager]("RobotManager", func(b host.IBuilder) {
		host.AddOption[*Option](b, "RobotManager")
	})
}

type Option struct {
	RobotCount       int    `snow:"RobotCount"`
	StartIndex       int    `snow:"StartIndex"`
	GateHost         string `snow:"GateHost"`
	GatePort         int    `snow:"GatePort"`
	ServerId         int64  `snow:"ServerId"`
	AccountPrefix    string `snow:"AccountPrefix"`
	UsernamePrefix   string `snow:"UsernamePrefix"`
	ChannelId        string `snow:"ChannelId"`
	Platform         string `snow:"Platform"`
	Appid            string `snow:"Appid"`
	Version          string `snow:"Version"`
	ReconnectDelayMs int    `snow:"ReconnectDelayMs"`
	PingIntervalMs   int    `snow:"PingIntervalMs"`
	DialTimeoutMs    int    `snow:"DialTimeoutMs"`
}

type RobotManager struct {
	node.Service

	lock        sync.Mutex
	opt         Option
	robotAddrs  []int32
	robotProxys []node.IProxy
}

func (ss *RobotManager) Start(args any) {
	opt := defaultOption()
	switch v := args.(type) {
	case *Option:
		if v != nil {
			opt = mergeOption(*v)
		}
	case Option:
		opt = mergeOption(v)
	}

	ss.EnableRpc()

	ss.lock.Lock()
	ss.opt = opt
	ss.robotAddrs = make([]int32, 0, opt.RobotCount)
	ss.robotProxys = make([]node.IProxy, 0, opt.RobotCount)
	ss.lock.Unlock()

	ss.Fork("RobotManager.Start.CreateRobots", func() {
		for i := 0; i < opt.RobotCount; i++ {
			id := opt.StartIndex + i
			robotOpt := robot.Option{
				Name:             fmt.Sprintf("robot_%d", id),
				GateHost:         opt.GateHost,
				GatePort:         opt.GatePort,
				ServerId:         opt.ServerId,
				Account:          fmt.Sprintf("%s_%d", opt.AccountPrefix, id),
				Username:         fmt.Sprintf("%s_%d", opt.UsernamePrefix, id),
				ChannelId:        opt.ChannelId,
				Platform:         opt.Platform,
				Appid:            opt.Appid,
				Version:          opt.Version,
				ReconnectDelayMs: opt.ReconnectDelayMs,
				PingIntervalMs:   opt.PingIntervalMs,
				DialTimeoutMs:    opt.DialTimeoutMs,
			}

			sAddr, proxy, err := ss.CreateService("Robot", &robotOpt)
			if err != nil {
				ss.Errorf("create robot service failed: %v", err)
				continue
			}

			ss.lock.Lock()
			ss.robotAddrs = append(ss.robotAddrs, sAddr)
			ss.robotProxys = append(ss.robotProxys, proxy)
			ss.lock.Unlock()

			proxy.Call("ApplyOption", robotOpt).Done()
		}
	})
}

func (ss *RobotManager) Stop(wg *sync.WaitGroup) {
	ss.lock.Lock()
	addrs := ss.robotAddrs
	ss.robotAddrs = nil
	ss.robotProxys = nil
	ss.lock.Unlock()

	if wg != nil {
		wg.Add(1)
	}
	go func() {
		defer func() {
			if wg != nil {
				wg.Done()
			}
		}()
		for _, sAddr := range addrs {
			_ = node.StopService(sAddr)
		}
	}()
}

func (ss *RobotManager) CreateService(name string, arg any) (int32, node.IProxy, error) {
	sAddr, err := node.NewService(name)
	if err != nil {
		return 0, nil, err
	}
	if ok := node.StartService(sAddr, arg); !ok {
		return 0, nil, fmt.Errorf("start service %s failed", name)
	}
	proxy := ss.CreateProxyByNodeAddr(node.AddrLocal, sAddr)
	if proxy == nil || !proxy.Avail() {
		_ = node.StopService(sAddr)
		return 0, nil, fmt.Errorf("create proxy for %s failed", name)
	}
	return sAddr, proxy, nil
}

func defaultOption() Option {
	return Option{
		RobotCount:       1,
		StartIndex:       0,
		GateHost:         "127.0.0.1",
		GatePort:         61101,
		ServerId:         1,
		AccountPrefix:    "robot",
		UsernamePrefix:   "robot",
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

	if opt.RobotCount > 0 {
		ret.RobotCount = opt.RobotCount
	}
	if opt.StartIndex >= 0 {
		ret.StartIndex = opt.StartIndex
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
	if opt.AccountPrefix != "" {
		ret.AccountPrefix = opt.AccountPrefix
	}
	if opt.UsernamePrefix != "" {
		ret.UsernamePrefix = opt.UsernamePrefix
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
