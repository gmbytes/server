package main

import (
	"context"

	"github.com/gmbytes/snow/pkg/configuration/sources"
	"github.com/gmbytes/snow/pkg/host"
	"github.com/gmbytes/snow/pkg/host/builder"
	"github.com/gmbytes/snow/routines/ignore_input"
	"github.com/gmbytes/snow/routines/node"
)

func main() {
	b := builder.NewDefaultBuilder()

	b.GetConfigurationManager().AddSource(&sources.YamlConfigurationSource{
		Path:           "conf/app.yml",
		Optional:       true,
		ReloadOnChange: true,
	})

	host.AddHostedRoutine[*ignore_input.IgnoreInput](b)
	host.AddOption[*node.Option](b, "Node")

	node.RegisterService(b, func(opt *node.RegisterOption) {
		opt.PostInitializer = func() {}
	})

	host.Run(b.Build(), context.Background())
}
