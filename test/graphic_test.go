package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env/config"

	"github.com/yasin-wu/graphic_captcha/captcha"

	"github.com/yasin-wu/graphic_captcha/common"

	"github.com/yasin-wu/utils/redis"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

var (
	captchaType = common.CaptchaTypeBlockPuzzle
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
	client, _ := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return apolloConf, nil
	})
	fmt.Println("初始化Apollo配置成功")
	cache := client.GetConfigCache(apolloConf.NamespaceName)
	host, _ := cache.Get("redis.host")
	password, _ := cache.Get("redis.password")
	c, err := captcha.New(captchaType, host.(string),
		captcha.WithRedisOptions(redis.WithPassWord(password.(string))))
	token := "XkNBUFQ6YmxvY2tfcHV6emxlO0NMSTp5YXNpbjtTVEFNUDoxNjM5MDE0MDk4Iw=="
	pointJson := "eyJYIjoyMDksIlkiOjV9"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		log.Fatal(err)
	}
	resps, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(resps))
}
