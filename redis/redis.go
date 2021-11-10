package redis

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/yasin-wu/utils/redis"
)

type Client struct {
	Client *redis.Client
}

func New(conf *redis.Config) *Client {
	client, err := redis.New(conf)
	if err != nil {
		panic(err)
	}
	return &Client{Client: client}
}

//校验数据存入Redis,存入时进行base64
func (this *Client) Set(token string, data interface{}, expireTime time.Duration) error {
	dataBuff, err := json.Marshal(data)
	if err != nil {
		return errors.New("json marshal error:" + err.Error())
	}
	data64 := base64.StdEncoding.EncodeToString(dataBuff)
	spew.Dump("数据:" + data64)
	err = this.Client.Set(token, data64, expireTime)
	if err != nil {
		return errors.New("存储至redis失败")
	}
	return nil
}

//从Redis获取待校验数据,并解base64
func (this *Client) Get(token string) ([]byte, error) {
	ttl, err := this.Client.TTL(token)
	if err != nil {
		return nil, err
	}
	if ttl <= 0 {
		err = this.Client.Del(token)
		return nil, errors.New("验证码已过期，请刷新重试")
	}
	cachedBuff, err := this.Client.Get(token)
	if err != nil {
		return nil, errors.New("get captcha error:" + err.Error())
	}
	var cachedStr string
	err = json.Unmarshal(cachedBuff, &cachedStr)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	base64Buff, err := base64.StdEncoding.DecodeString(cachedStr)
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	return base64Buff, nil
}
