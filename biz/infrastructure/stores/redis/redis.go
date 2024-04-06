package redis

import (
	"github.com/CloudStriver/platform-comment/biz/infrastructure/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func NewRedis(config *config.Config) *redis.Redis {
	return redis.MustNewRedis(*config.Redis)
}
