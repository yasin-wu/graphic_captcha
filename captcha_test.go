package captcha

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/yasin-wu/captcha/captcha"
	"github.com/yasin-wu/captcha/redis"
)

var (
	captchaType = captcha.CaptchaTypeClickWord
	captchaConf = &captcha.CaptchaConfig{
		ClickImagePath:     "./pic_click",
		ClickWordFile:      "./fonts/license.txt",
		FontFile:           "./fonts/captcha.ttf",
		JigsawOriginalPath: "./jigsaw/original",
		JigsawBlockPath:    "./jigsaw/sliding_block",
		ExpireTime:         30 * 60,
	}
	redisConf = &redis.RedisConfig{
		NetWork:  "tcp",
		Host:     "192.168.131.135:6379",
		PassWord: "1qazxsw21201",
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
	pointJson := "W3siWCI6MSwiWSI6OTAsIlRleHQiOiLor6UifSx7IlgiOjc2LCJZIjozOSwiVGV4dCI6IuecvCJ9LHsiWCI6MTY4LCJZIjo1NSwiVGV4dCI6IuWwmCJ9XQ=="
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTYyNjY3OTAzNCM="
	resp, err := c.Check(token, pointJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump(resp)
}
