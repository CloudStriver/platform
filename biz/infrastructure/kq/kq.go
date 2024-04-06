package kq

import (
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	"github.com/zeromicro/go-queue/kq"
)

type DeleteCommentRelationKq struct {
	*kq.Pusher
}

func NewDeleteCommentRelationKq(c *config.Config) *DeleteCommentRelationKq {
	pusher := kq.NewPusher(c.DeleteCommentRelationKq.Brokers, c.DeleteCommentRelationKq.Topic)
	return &DeleteCommentRelationKq{
		Pusher: pusher,
	}
}
