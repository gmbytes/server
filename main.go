package main

import (
	"server/lib/pkg"

	"github.com/gmbytes/snow/core/configuration/sources"
	"github.com/gmbytes/snow/core/host"
	"github.com/gmbytes/snow/core/host/builder"
	"github.com/gmbytes/snow/routines/ignore_input"
	"github.com/gmbytes/snow/routines/node"
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
		opt.ServerHandlePreprocessor = pkg.ServerPkgPreprocessor
		opt.ClientHandlePreprocessor = pkg.ClientPkgPreprocessor
		//MetricCollector:          host.GetRoutine[*metrics.Meter](b.GetRoutineProvider()),
		opt.PostInitializer = func() {
			//if len(app.GameConfigPath) > 0 {
			//	_ = conf.GetConfig()
			//}
		}
	})
	host.Run(b.Build())
}
