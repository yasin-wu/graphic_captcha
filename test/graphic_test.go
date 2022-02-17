package test

import (
	"encoding/base64"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/yasin-wu/graphic_captcha/v2/captcha"
)

var (
	captchaType  = captcha.CaptchaTypeClickWord
	redisOptions = &captcha.RedisOptions{Addr: "47.108.155.25:6379", Password: "yasinwu"}
)

func TestGet(t *testing.T) {
	c, err := captcha.New(captchaType, redisOptions, captcha.WithExpireTime(30*time.Minute))
	if err != nil {
		log.Fatal(err)
	}
	token := fmt.Sprintf(captcha.TokenFormat, string(captchaType), "yasin", time.Now().Unix())
	token = base64.StdEncoding.EncodeToString([]byte(token))
	_, err = c.Get(token)
	if err != nil {
		log.Fatal(err)
	}
}

func TestCheck(t *testing.T) {
	c, err := captcha.New(captchaType, redisOptions)
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTY0NTA2NDg4NyM="
	pointJson := "W3siWCI6MTIsIlkiOjM4LCJUIjoi5aeLIn0seyJYIjo4MywiWSI6OCwiVCI6IumTuiJ9LHsiWCI6MTY3LCJZIjo0MSwiVCI6IumHjCJ9XQ=="
	resp, err := c.Check(token, pointJson)
	if err != nil {
		log.Fatal(err)
	}
	captcha.Println(resp)
}
