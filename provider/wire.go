//go:build wireinject
// +build wireinject

package provider

import (
	"github.com/CloudStriver/platform-comment/biz/adaptor"
	"github.com/google/wire"
)

func NewCommentServerImpl() (*adaptor.CommentServerImpl, error) {
	wire.Build(
		wire.Struct(new(adaptor.CommentServerImpl), "*"),
		AllProvider,
	)
	return nil, nil
}
