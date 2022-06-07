package graphic_captcha

import (
	"errors"

	"github.com/yasin-wu/graphic_captcha/v2/slideblock"

	"github.com/yasin-wu/graphic_captcha/v2/clickword"

	"github.com/yasin-wu/graphic_captcha/v2/factory"

	"github.com/yasin-wu/graphic_captcha/v2/consts"
	"github.com/yasin-wu/graphic_captcha/v2/redis"
)

func New(captchaType consts.CaptchaType, redisOptions *redis.RedisOptions, options ...factory.Option) (factory.Captcha, error) {
	conf := &factory.Config{}
	for _, f := range options {
		f(conf)
	}
	factory.CheckConf(conf)
	redisCli := redis.New(redisOptions)
	if redisCli == nil {
		return nil, errors.New("redis client is nil")
	}
	switch captchaType {
	case consts.CaptchaTypeClickWord:
		return clickword.New(redisCli, *conf), nil
	case consts.CaptchaTypeBlockPuzzle:
		return slideblock.New(redisCli, *conf), nil
	}
	return nil, errors.New("验证类型错误")
}
