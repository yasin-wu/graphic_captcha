package test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/yasin-wu/captcha/captcha"

	"github.com/yasin-wu/captcha/common"

	"github.com/davecgh/go-spew/spew"

	"github.com/yasin-wu/utils/redis"
)

var (
	captchaType = common.CaptchaTypeClickWord
	captchaConf = &captcha.Config{
		ExpireTime: 30 * time.Minute,
	}
)

func TestCaptchaGet(t *testing.T) {
	c, err := captcha.New(captchaType, captchaConf, "47.108.155.25:6379",
		redis.WithPassWord("yasinwu"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	token := fmt.Sprintf(common.TokenFormat, string(captchaType), "yasin", time.Now().Unix())
	token = base64.StdEncoding.EncodeToString([]byte(token))
	cv, err := c.Get(token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump("Token:" + cv.Token)
}

func TestCaptchaCheck(t *testing.T) {
	c, err := captcha.New(captchaType, captchaConf, "47.108.155.25:6379",
		redis.WithPassWord("yasinwu"))
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTYzNjk1MDcxNiM="
	pointJson := "W3siWCI6MTEzLCJZIjoxMDgsIlRleHQiOiLlh6AifSx7IlgiOjE0OSwiWSI6MTExLCJUZXh0Ijoi5ZyoIn0seyJYIjoyMjIsIlkiOjYxLCJUZXh0Ijoi5L2gIn1d"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump(resp)
}
