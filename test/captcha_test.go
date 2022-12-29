package test

import (
	"encoding/base64"
	"fmt"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/config"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/consts"
	"log"
	"testing"
	"time"

	graphiccaptcha "github.com/yasin-wu/graphic_captcha/v2/captcha"
	"github.com/yasin-wu/graphic_captcha/v2/internal/util"
)

var (
	captchaType  = consts.CaptchaTypeBlockPuzzle
	redisOptions = &config.RedisOptions{Addr: "localhost:6379", Password: "yasinwu"}
)

func TestGet(t *testing.T) {
	c, err := graphiccaptcha.New(captchaType, redisOptions, config.WithExpireTime(30*time.Minute))
	if err != nil {
		log.Fatal(err)
	}
	token := fmt.Sprintf(consts.TokenFormat, string(captchaType), "yasin", time.Now().Unix())
	token = base64.StdEncoding.EncodeToString([]byte(token))
	resp, err := c.Get(token)
	if err != nil {
		log.Fatal(err)
	}
	util.Println(resp)
}

func TestCheck(t *testing.T) {
	c, err := graphiccaptcha.New(captchaType, redisOptions)
	if err != nil {
		log.Fatal(err)
	}
	token := "XkNBUFQ6YmxvY2tfcHV6emxlO0NMSTp5YXNpbjtTVEFNUDoxNjYwMTEwODA0Iw==" //nolint:gosec
	pointJSON := "eyJYIjoyMDIsIlkiOjV9"
	resp, err := c.Check(token, pointJSON)
	if err != nil {
		log.Fatal(err)
	}
	util.Println(resp)
}
