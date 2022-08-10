[![OSCS Status](https://www.oscs1024.com/platform/badge/yasin-wu/graphic_captcha.svg?size=small)](https://www.murphysec.com/dr/eyEhxXb8cpZUwJ4dfO)
## 介绍

Golang版本的文字点选验证和滑块验证

## 安装

```
go get -u github.com/yasin-wu/graphic_captcha
```

推荐使用go.mod

```
require github.com/yasin-wu/graphic_captcha/v2 v2.2.0
```

## 使用

```go
var (
captchaType = consts.CaptchaTypeClickWord
redisOptions = &redis.RedisOptions{Addr: "127.0.0.1:6379", Password: ""}
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
```
