//go:build wireinject
// +build wireinject

package provider

import (
	"github.com/CloudStriver/platform-comment/biz/adaptor"
	"github.com/google/wire"
)

func NewPlatformServerImpl() (*adaptor.PlatformServerImpl, error) {
	wire.Build(
		wire.Struct(new(adaptor.PlatformServerImpl), "*"),
		AllProvider,
	)
	return nil, nil
}
