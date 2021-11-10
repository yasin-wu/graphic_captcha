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
		ClickImagePath:     "../conf/pic_click",
		ClickWordFile:      "../conf/fonts/license.txt",
		FontFile:           "../conf/fonts/captcha.ttf",
		JigsawOriginalPath: "../conf/jigsaw/original",
		JigsawBlockPath:    "../conf/jigsaw/sliding_block",
		ExpireTime:         30 * time.Minute,
	}
	redisConf = &redis.Config{
		Host:     "47.108.155.25:6379",
		PassWord: "yasinwu",
		DB:       0,
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
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTYzNjUwOTM0NyM="
	pointJson := "W3siWCI6NiwiWSI6NjgsIlRleHQiOiLmnIkifSx7IlgiOjE4NiwiWSI6MjgsIlRleHQiOiLmiJEifSx7IlgiOjIyMCwiWSI6MTEzLCJUZXh0Ijoi5LykIn1d"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump(resp)
}
