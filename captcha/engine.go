package captcha

import (
	"errors"

	"github.com/yasin-wu/graphic_captcha/redis"
)

/**
 * @author: yasin
 * @date: 2022/1/13 14:16
 * @description: 验证器
 */
type Engine interface {
	/**
	 * @author: yasin
	 * @date: 2022/1/13 14:17
	 * @params: token string
	 * @return: *common.Captcha, error
	 * @description: 获取待验证信息
	 */
	Get(token string) (*Captcha, error)
	/**
	 * @author: yasin
	 * @date: 2022/1/13 14:17
	 * @params: token, pointJson string
	 * @return: *common.RespMsg, error
	 * @description: 校验用户操作结果
	 */
	Check(token, pointJson string) (*RespMsg, error)
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:18
 * @params: captchaType common.CaptchaType, redisHost string, options ...Option
 * @return: Engine, error
 * @description: 新建验证器
 */
func New(captchaType CaptchaType, redisHost string, options ...Option) (Engine, error) {
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
	case CaptchaTypeClickWord:
		return &ClickWord{conf: conf}, nil
	case CaptchaTypeBlockPuzzle:
		return &SlideBlock{conf: conf}, nil
	}
	return nil, errors.New("验证类型错误")
}
