package captcha

import (
	"errors"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/config"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/consts"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/factory"

	"github.com/yasin-wu/graphic_captcha/v2/internal/clickword"
	"github.com/yasin-wu/graphic_captcha/v2/internal/redis"
	"github.com/yasin-wu/graphic_captcha/v2/internal/slideblock"
)

// nolint:lll
func New(captchaType consts.CaptchaType, redisOptions *config.RedisOptions, options ...config.Option) (factory.Captchaer, error) {
	conf := &config.Config{}
	for _, f := range options {
		f(conf)
	}
	config.CheckConf(conf)
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
