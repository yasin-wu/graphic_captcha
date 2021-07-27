package redis

import (
	"github.com/yasin-wu/utils/redis"
)

var (
	RedisClient *redis.Client
)

func InitRedis(conf *redis.Config) {
	var err error
	RedisClient, err = redis.New(conf)
	if err != nil {
		panic(err)
	}
}
