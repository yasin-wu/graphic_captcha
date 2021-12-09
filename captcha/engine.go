package captcha

import (
	"errors"

	"github.com/yasin-wu/graphic_captcha/redis"

	"github.com/yasin-wu/graphic_captcha/common"
)

type Engine interface {
	Get(token string) (*common.Captcha, error)
	Check(token, pointJson string) (*common.RespMsg, error)
}

func New(captchaType common.CaptchaType, redisHost string, options ...Option) (Engine, error) {
	if redisHost == "" {
		return nil, errors.New("redis host is nil")
	}
	conf := &config{}
	for _, f := range options {
		f(conf)
	}
	checkConf(conf)
	redisCli := redis.New(redisHost, conf.redisOptions...)
	if redisCli == nil {
		return nil, errors.New("redis client is nil")
	}
	conf.redisCli = redisCli
	switch captchaType {
	case common.CaptchaTypeClickWord:
		return &ClickWord{conf: conf}, nil
	case common.CaptchaTypeBlockPuzzle:
		return &SlideBlock{conf: conf}, nil
	}
	return nil, errors.New("验证类型错误")
}
