package captcha

import (
	"errors"

	"github.com/yasin-wu/graphic_captcha/v2/internal/clickword"
	"github.com/yasin-wu/graphic_captcha/v2/internal/redis"
	"github.com/yasin-wu/graphic_captcha/v2/internal/slideblock"
	"github.com/yasin-wu/graphic_captcha/v2/pkg"
)

// nolint:lll
func New(captchaType pkg.CaptchaType, redisOptions *pkg.RedisOptions, options ...pkg.Option) (pkg.Captchaer, error) {
	conf := &pkg.Config{}
	for _, f := range options {
		f(conf)
	}
	pkg.CheckConf(conf)
	redisCli := redis.New(redisOptions)
	if redisCli == nil {
		return nil, errors.New("redis client is nil")
	}
	switch captchaType {
	case pkg.CaptchaTypeClickWord:
		return clickword.New(redisCli, *conf), nil
	case pkg.CaptchaTypeBlockPuzzle:
		return slideblock.New(redisCli, *conf), nil
	}
	return nil, errors.New("验证类型错误")
}
