package test

import (
	"encoding/base64"
	"fmt"
	"log"
	"testing"
	"time"

	graphiccaptcha "github.com/yasin-wu/graphic_captcha/v2"

	"github.com/yasin-wu/graphic_captcha/v2/util"

	"github.com/yasin-wu/graphic_captcha/v2/consts"
	"github.com/yasin-wu/graphic_captcha/v2/redis"

	"github.com/yasin-wu/graphic_captcha/v2/factory"
)

var (
	captchaType  = consts.CaptchaTypeBlockPuzzle
	redisOptions = &redis.Options{Addr: "localhost:6379", Password: "yasinwu"}
)

func TestGet(t *testing.T) {
	c, err := graphiccaptcha.New(captchaType, redisOptions, factory.WithExpireTime(30*time.Minute))
	if err != nil {
		log.Fatal(err)
	}
	token := fmt.Sprintf(consts.TokenFormat, string(captchaType), "yasin", time.Now().Unix())
	token = base64.StdEncoding.EncodeToString([]byte(token))
	_, err = c.Get(token)
	if err != nil {
		log.Fatal(err)
	}
}

func TestCheck(t *testing.T) {
	c, err := graphiccaptcha.New(captchaType, redisOptions)
	if err != nil {
		log.Fatal(err)
	}
	token := "XkNBUFQ6YmxvY2tfcHV6emxlO0NMSTp5YXNpbjtTVEFNUDoxNjU4NzI5MDIzIw==" //nolint:gosec
	pointJSON := "eyJYIjoxMTgsIlkiOjV9"
	resp, err := c.Check(token, pointJSON)
	if err != nil {
		log.Fatal(err)
	}
	util.Println(resp)
}
