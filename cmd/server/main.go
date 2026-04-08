package main

import (
	"context"

	"github.com/gmbytes/snow/pkg/configuration/sources"
	"github.com/gmbytes/snow/pkg/host"
	"github.com/gmbytes/snow/pkg/host/builder"
	"github.com/gmbytes/snow/pkg/routines/ignore_input"
	"github.com/gmbytes/snow/pkg/routines/node"

	"server/pkg/net_pkg"
)

func main() {
	b := builder.NewDefaultBuilder()

	// 加载 YAML 配置文件
	b.GetConfigurationManager().AddSource(&sources.YamlConfigurationSource{
		Path:           "conf/app.yml",
		Optional:       true, // 如果文件不存在也不报错
		ReloadOnChange: true, // 支持热更新
	})

	host.AddHostedRoutine[*ignore_input.IgnoreInput](b)

	// 节点配置：从 app.yml 读取，如果没有配置则使用默认值
	host.AddOption[*node.Option](b, "Node")

	node.RegisterService(b, func(opt *node.RegisterOption) {
		opt.ServerHandlePreprocessor = net_pkg.ServerPkgPreprocessor
		opt.ClientHandlePreprocessor = net_pkg.ClientPkgPreprocessor
		// MetricCollector:          host.GetRoutine[*metrics.Meter](b.GetRoutineProvider()),
		opt.PostInitializer = func() {
		}
	})
	host.Run(b.Build(), context.Background())
}
