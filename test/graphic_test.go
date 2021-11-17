package test

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"

	"github.com/yasin-wu/graphic_captcha/captcha"

	"github.com/yasin-wu/graphic_captcha/common"

	"github.com/davecgh/go-spew/spew"

	"github.com/yasin-wu/utils/redis"
)

var (
	captchaType = common.CaptchaTypeClickWord
)

func TestGet(t *testing.T) {
	client, _ := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return apolloConf, nil
	})
	fmt.Println("初始化Apollo配置成功")
	cache := client.GetConfigCache(apolloConf.NamespaceName)
	host, _ := cache.Get("redis.host")
	password, _ := cache.Get("redis.password")
	c, err := captcha.New(captchaType, host.(string),
		captcha.WithRedisOptions(redis.WithPassWord(password.(string))))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	token := fmt.Sprintf(common.TokenFormat, string(captchaType), "yasin", time.Now().Unix())
	token = base64.StdEncoding.EncodeToString([]byte(token))
	cv, err := c.Get(token)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump("Token:" + cv.Token)
}

func TestCheck(t *testing.T) {
	client, _ := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return apolloConf, nil
	})
	fmt.Println("初始化Apollo配置成功")
	cache := client.GetConfigCache(apolloConf.NamespaceName)
	host, _ := cache.Get("redis.host")
	password, _ := cache.Get("redis.password")
	c, err := captcha.New(captchaType, host.(string),
		captcha.WithRedisOptions(redis.WithPassWord(password.(string))))
	token := "XkNBUFQ6Y2xpY2tfd29yZDtDTEk6eWFzaW47U1RBTVA6MTYzNzExODA5MyM="
	pointJson := "W3siWCI6OSwiWSI6OCwiVCI6IuadpSJ9LHsiWCI6MTg0LCJZIjo5MiwiVCI6IuWMliJ9LHsiWCI6MjU0LCJZIjo2LCJUIjoi566hIn1d"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	spew.Dump(resp)
}