## 介绍

Golang版本的文字点选验证和滑块验证

## 安装

````
go get -u github.com/yasin-wu/graphic_captcha
````

推荐使用go.mod

````
require github.com/yasin-wu/graphic_captcha/v2 v2.0.1
````

## 使用

````go
func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

var (
	captchaType = captcha.CaptchaTypeBlockPuzzle
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
    c, err := captcha.New(captchaType, redisOptions, captcha.WithExpireTime(30*time.Minute))
    if err != nil {
        log.Fatal(err)
    }
    token := "XkNBUFQ6YmxvY2tfcHV6emxlO0NMSTp5YXNpbjtTVEFNUDoxNjQyMDU1MDk0Iw=="
	pointJson := "eyJYIjoxNDEsIlkiOjV9"
	resp, err := c.Check(token, pointJson)
	if err != nil {
		log.Fatal(err)
	}
	resps, _ := json.MarshalIndent(resp, "", "\t")
	fmt.Println(string(resps))
}

````