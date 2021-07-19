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
		ClickImagePath:     "/Users/yasin/GolandProjects/yasin-wu/captcha/pic_click",
		ClickWordFile:      "/Users/yasin/GolandProjects/yasin-wu/captcha/fonts/license.txt",
		FontFile:           "/Users/yasin/GolandProjects/yasin-wu/captcha/fonts/captcha.ttf",
		JigsawOriginalPath: "/Users/yasin/GolandProjects/yasin-wu/captcha/jigsaw/original",
		JigsawBlockPath:    "/Users/yasin/GolandProjects/yasin-wu/captcha/jigsaw/sliding_block",
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
	spew.Dump(cv.Token)
}

func TestCaptchaCheck(t *testing.T) {
	c, err := captcha.New(captchaType, captchaConf, redisConf)
	/*if err != nil {
		fmt.Println(err.Error())
		return
	}
	p1 := captcha.FontPoint{
		X:    104,
		Y:    83,
		Text: "前",
	}
	p2 := captcha.FontPoint{
		X:    152,
		Y:    72,
		Text: "你",
	}
	p3 := captcha.FontPoint{
		X:    252,
		Y:    94,
		Text: "只",
	}
	var points []captcha.FontPoint
	points = append(points, p1, p2, p3)
	pointBuff, _ := json.Marshal(points)
	pointJson := base64.StdEncoding.EncodeToString(pointBuff)*/
	pointJson := "eyJYIjoxOTcsIlkiOjV9"
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTYyNjY3NTU1MCM="
	resp, err := c.Check(token, pointJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump(resp)
}
