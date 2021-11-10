package test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/yasin-wu/captcha/captcha"
	"github.com/yasin-wu/utils/redis"
)

var (
	captchaType = captcha.CaptchaTypeClickWord
	captchaConf = &captcha.Config{
		ExpireTime: 30 * time.Minute,
	}
	redisConf = &redis.Config{
		Host:     "47.108.155.25:6379",
		PassWord: "yasinwu",
	}
)

func TestCaptchaGet(t *testing.T) {
	c, err := captcha.New(captchaType, captchaConf, redisConf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	token := fmt.Sprintf(captcha.TokenFormat, string(captchaType), "yasin", time.Now().Unix())
	token = base64.StdEncoding.EncodeToString([]byte(token))
	cv, err := c.Get(token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump("Token:" + cv.Token)
}

func TestCaptchaCheck(t *testing.T) {
	c, err := captcha.New(captchaType, captchaConf, redisConf)
	//先转为byte,然后base64
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTYzNjUxMDU0MyM="
	pointJson := "W3siWCI6ODIsIlkiOjYwLCJUZXh0Ijoi6YeMIn0seyJYIjoxODUsIlkiOjMzLCJUZXh0Ijoi5YiwIn0seyJYIjoyNDYsIlkiOjI0LCJUZXh0Ijoi6YCDIn1d"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump(resp)
}
