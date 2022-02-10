package captcha

import (
	"errors"
)

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:16
 * @description: 验证器
 */
type Engine interface {
	/**
	 * @author: yasinWu
	 * @date: 2022/1/13 14:17
	 * @params: token string
	 * @return: *common.Captcha, error
	 * @description: 获取待验证信息
	 */
	Get(token string) (*Captcha, error)
	/**
	 * @author: yasinWu
	 * @date: 2022/1/13 14:17
	 * @params: token, pointJson string
	 * @return: *common.RespMsg, error
	 * @description: 校验用户操作结果
	 */
	Check(token, pointJson string) (*RespMsg, error)
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:18
 * @params: captchaType CaptchaType, redisOptions *RedisOptions, options ...Option
 * @return: Engine, error
 * @description: 新建验证器
 */
func New(captchaType CaptchaType, redisOptions *RedisOptions, options ...Option) (Engine, error) {
	conf := &config{}
	for _, f := range options {
		f(conf)
	}
	checkConf(conf)
	redisCli := newRedis(redisOptions)
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
