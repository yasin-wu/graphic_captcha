package captcha

import (
	"errors"

	"github.com/yasin-wu/captcha/redis"

	"github.com/yasin-wu/captcha/common"
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
	switch captchaType {
	case common.CaptchaTypeClickWord:
		return &ClickWord{
			imagePath:     conf.clickImagePath,
			wordFile:      conf.clickWordFile,
			wordCount:     conf.clickWordCount,
			fontFile:      conf.fontFile,
			fontSize:      conf.fontSize,
			watermarkText: conf.watermarkText,
			watermarkSize: conf.watermarkSize,
			dpi:           conf.dpi,
			expireTime:    conf.expireTime,
			redisCli:      redisCli,
		}, nil
	case common.CaptchaTypeBlockPuzzle:
		return &SlideBlock{
			originalPath:  conf.originalPath,
			blockPath:     conf.blockPath,
			threshold:     conf.threshold,
			blur:          conf.blur,
			brightness:    conf.brightness,
			fontFile:      conf.fontFile,
			fontSize:      conf.fontSize,
			watermarkText: conf.watermarkText,
			watermarkSize: conf.watermarkSize,
			dpi:           conf.dpi,
			expireTime:    conf.expireTime,
			redisCli:      redisCli,
		}, nil
	}
	return nil, errors.New("验证类型错误")
}
