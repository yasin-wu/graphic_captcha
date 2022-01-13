package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/yasin-wu/graphic_captcha/captcha"

	"github.com/yasin-wu/graphic_captcha/common"

	"github.com/yasin-wu/utils/redis"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

var (
	captchaType = common.CaptchaTypeBlockPuzzle
	host        = "47.108.155.25:6379"
	password    = "yasinwu"
)

func TestGet(t *testing.T) {

	c, err := captcha.New(captchaType, host,
		captcha.WithRedisOptions(redis.WithPassWord(password)))
	if err != nil {
		log.Fatal(err)
	}
	token := fmt.Sprintf(common.TokenFormat, string(captchaType), "yasin", time.Now().Unix())
	token = base64.StdEncoding.EncodeToString([]byte(token))
	cv, err := c.Get(token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Token:" + cv.Token)
}

func TestCheck(t *testing.T) {
	c, err := captcha.New(captchaType, host,
		captcha.WithRedisOptions(redis.WithPassWord(password)))
	token := "XkNBUFQ6YmxvY2tfcHV6emxlO0NMSTp5YXNpbjtTVEFNUDoxNjQxNDUzMzExIw=="
	pointJson := "eyJYIjoxMTEsIlkiOjV9"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		log.Fatal(err)
	}
	resps, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(resps))
}
