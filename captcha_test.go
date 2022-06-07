package graphic_captcha

import (
	"encoding/base64"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/yasin-wu/graphic_captcha/v2/util"

	"github.com/yasin-wu/graphic_captcha/v2/consts"
	"github.com/yasin-wu/graphic_captcha/v2/redis"

	"github.com/yasin-wu/graphic_captcha/v2/factory"
)

var (
	captchaType  = consts.CaptchaTypeClickWord
	redisOptions = &redis.RedisOptions{Addr: "10.34.5.69:6379", Password: ""}
)

func TestGet(t *testing.T) {
	c, err := New(captchaType, redisOptions, factory.WithExpireTime(30*time.Minute))
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
	c, err := New(captchaType, redisOptions)
	token := "XkNBUFQ6YmxvY2tfcHV6emxlO0NMSTp5YXNpbjtTVEFNUDoxNjU0NTY3NjU3Iw=="
	pointJson := "eyJYIjoyMjMsIlkiOjV9"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		log.Fatal(err)
	}
	util.Println(resp)
}
