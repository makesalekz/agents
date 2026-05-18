//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"github.com/makesalekz/agents/internal/biz"
	"github.com/makesalekz/agents/internal/conf"
	"github.com/makesalekz/agents/internal/data"
	"github.com/makesalekz/agents/internal/server"
	"github.com/makesalekz/agents/internal/service"
)

func wireApp(*conf.Bootstrap, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
