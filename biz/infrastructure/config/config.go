package config

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"os"
)

type ElasticsearchConf struct {
	Addresses []string
	Username  string
	Password  string
}

type KqConfig struct {
	Brokers []string
	Topic   string
}

type EtcdConf struct {
	Hosts []string
}

type Config struct {
	service.ServiceConf
	ListenOn string
	Mongo    struct {
		URL string
		DB  string
	}
	CacheConf     cache.CacheConf
	Elasticsearch ElasticsearchConf
	Redis         *redis.RedisConf
	Neo4jConf     struct {
		Url      string
		Username string
		Password string
		DataBase string
		Enable   bool
	}
	DeleteCommentRelationKq KqConfig
}

func NewConfig() (*Config, error) {
	c := new(Config)
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "etc/config.yaml"
	}
	err := conf.Load(path, c)
	if err != nil {
		return nil, err
	}
	err = c.SetUp()
	if err != nil {
		return nil, err
	}
	return c, nil
}
