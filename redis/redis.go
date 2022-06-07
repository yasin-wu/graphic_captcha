package redis

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisOptions redis.Options

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

var defaultRedisOptions = &RedisOptions{Addr: "localhost:6379", Password: "", DB: 0}

func New(options *RedisOptions) *RedisClient {
	if options == nil {
		options = defaultRedisOptions
	}
	return &RedisClient{client: redis.NewClient((*redis.Options)(options)), ctx: context.Background()}
}

func (r *RedisClient) Set(token string, data interface{}, expireTime time.Duration) error {
	dataBuff, err := json.Marshal(data)
	if err != nil {
		return errors.New("json marshal error:" + err.Error())
	}
	data64 := base64.StdEncoding.EncodeToString(dataBuff)
	fmt.Println("Token: ", token)
	fmt.Println("PointJson: ", data64)
	err = r.client.Set(r.ctx, token, data64, expireTime).Err()
	if err != nil {
		return errors.New("存储至redis失败")
	}
	return nil
}

func (r *RedisClient) Get(token string) ([]byte, error) {
	ttl, err := r.client.TTL(r.ctx, token).Result()
	if err != nil {
		return nil, err
	}
	if ttl <= 0 {
		err = r.client.Del(r.ctx, token).Err()
		if err != nil {
			log.Println("删除key失败")
		}
		return nil, errors.New("验证码已过期，请刷新重试")
	}
	cachedBuff, err := r.client.Get(r.ctx, token).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("%s does not exist", token)
	}
	if err != nil {
		return nil, errors.New("get captcha error:" + err.Error())
	}
	base64Buff, err := base64.StdEncoding.DecodeString(cachedBuff)
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	return base64Buff, nil
}

func (r *RedisClient) Del(token string) error {
	return r.client.Del(r.ctx, token).Err()
}
