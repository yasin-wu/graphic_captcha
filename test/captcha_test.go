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
)

func TestCaptchaGet(t *testing.T) {
	c, err := captcha.New(captchaType, "47.108.155.25:6379",
		captcha.WithRedisOptions(redis.WithPassWord("yasinwu")))
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
	c, err := captcha.New(captchaType, "47.108.155.25:6379",
		captcha.WithRedisOptions(redis.WithPassWord("yasinwu")))
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTYzNjk1NjkyOCM="
	pointJson := "W3siWCI6MzMsIlkiOjEwNSwiVGV4dCI6IueVmSJ9LHsiWCI6NzEsIlkiOjYsIlRleHQiOiLlroMifSx7IlgiOjI0MCwiWSI6MzksIlRleHQiOiLlpLEifV0="
	resp, err := c.Check(token, pointJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump(resp)
}
