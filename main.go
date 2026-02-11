package main

import (
	"server/service/gate"

	"github.com/gmbytes/snow/core/host"
	"github.com/gmbytes/snow/core/host/builder"
	"github.com/gmbytes/snow/routines/ignore_input"
	"github.com/gmbytes/snow/routines/node"
)

func main() {
	b := builder.NewDefaultBuilder()
	host.AddHostedRoutine[*ignore_input.IgnoreInput](b)

	// 节点配置：当前启动节点名与要运行的服务
	host.AddOption[*node.Option](b, "Node")
	host.AddOptionFactory[*node.Option](b, func() *node.Option {
		return &node.Option{
			BootName: "GameNode",
			LocalIP:  "127.0.0.1",
			Nodes: map[string]*node.ElementOption{
				"GameNode": {
					Services: []string{"Gate"},
				},
			},
		}
	})

	// 注册 Node Routine 与 Gate 服务
	node.AddNode(b, func() *node.RegisterOption {
		return &node.RegisterOption{
			ServiceRegisterInfos: []*node.ServiceRegisterInfo{
				node.CheckedServiceRegisterInfoName[gate.Gate, *gate.Gate](1, "Gate"),
			},
		}
	})

	host.Run(b.Build())
}
