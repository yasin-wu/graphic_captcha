package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/yasin-wu/graphic_captcha/v2/captcha"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

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
	cv, err := c.Get(token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Token:" + cv.Token)
}

func TestCheck(t *testing.T) {
	c, err := captcha.New(captchaType, redisOptions)
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTY0NDQ1OTUwMCM="
	pointJson := "W3siWCI6MzksIlkiOjUxLCJUIjoi5ZOAIn0seyJYIjoxNzEsIlkiOjc1LCJUIjoi6L+eIn0seyJYIjoyMzgsIlkiOjUyLCJUIjoi5oCVIn1d"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		log.Fatal(err)
	}
	resps, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(resps))
}
