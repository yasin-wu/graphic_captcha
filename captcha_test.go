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
	spew.Dump(cv)
}

func TestCaptchaCheck(t *testing.T) {
	c, err := captcha.New(captchaType, captchaConf, redisConf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	/*p1 := captcha.FontPoint{
		X:    15,
		Y:    111,
		Text: "更",
	}
	p2 := captcha.FontPoint{
		X:    167,
		Y:    105,
		Text: "恋",
	}
	p3 := captcha.FontPoint{
		X:    226,
		Y:    62,
		Text: "岁",
	}
	var points []captcha.FontPoint
	points = append(points, p1, p2, p3)
	pointBuff, _ := json.Marshal(points)*/
	token := "XkNBUFQ6YmxvY2tfcHV6emxlO0NMSTp5YXNpbjtTVEFNUDoxNjI2MTU3ODM4Iw=="
	pointJson := "eyJYIjoxMDAsIlkiOjZ9"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump(resp)
}
