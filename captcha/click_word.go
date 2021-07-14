package captcha

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"

	mredis "github.com/gomodule/redigo/redis"
	"github.com/yasin-wu/captcha/redis"

	"github.com/golang/freetype"
)

type ClickWord struct {
}

type FontPoint struct {
	X    int
	Y    int
	Text string
}

func (this *ClickWord) Get(token string) (*CaptchaVO, error) {
	imgFile, err := randomFile(captchaConf.ClickImagePath)
	if err != nil {
		return nil, errors.New("random file error:" + err.Error())
	}
	defer imgFile.Close()
	fileType := strings.Replace(path.Ext(path.Base(imgFile.Name())), ".", "", -1)
	var staticImg image.Image
	switch fileType {
	case "png":
		staticImg, err = png.Decode(imgFile)
	default:
		staticImg, err = jpeg.Decode(imgFile)
	}
	if err != nil {
		return nil, errors.New("image decode error:" + err.Error())
	}
	img := image2RGBA(staticImg)
	if img == nil {
		return nil, errors.New("image to rgba failed")
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	err = drawText(img)
	if err != nil {
		return nil, errors.New("draw watermark failed:" + err.Error())
	}

	fontBytes, err := ioutil.ReadFile(captchaConf.FontFile)
	if err != nil {
		return nil, errors.New("read font file error:" + err.Error())
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, errors.New("parse font error:" + err.Error())
	}
	str, err := this.randomHanZi()
	if err != nil {
		return nil, errors.New("randomHanZi error:" + err.Error())
	}
	//需要把这个存入Redis作为校验
	var allDots []FontPoint
	words := this.randomNoCheck(str)
	fontSize := captchaConf.FontSize
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len(str); i++ {
		_w := (width - 24) / len(str)
		x := i*_w + rand.Intn(_w-fontSize)
		y := rand.Intn(height - fontSize - fontSize/2)
		//随机生成字体颜色
		fontColor := image.NewUniform(color.RGBA{R: uint8(rand.Intn(255)), G: uint8(rand.Intn(255) + 50),
			B: uint8(rand.Intn(255)), A: uint8(255)})
		text := fmt.Sprintf("%c", str[i])
		//随机旋转角度
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		angle := float64(r.Intn(40) - 20)
		drawTextOnBackground(img, image.Pt(x, y), font, text, fontColor, fontSize, angle)
		if StringsContains(words, text) {
			allDots = append(allDots, FontPoint{x, y, text})
		}
	}

	base64_, err := imgToBase64(img, fileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}
	//校验数据base64后存入Redis
	allDotsBuff, err := json.Marshal(allDots)
	if err != nil {
		return nil, errors.New("json marshal error:" + err.Error())
	}
	data64 := base64.StdEncoding.EncodeToString(allDotsBuff)
	err = SetRedis(token, data64)
	if err != nil {
		return nil, err
	}
	return &CaptchaVO{
		OriginalImageBase64: base64_,
		Words:               words,
		CaptchaType:         string(CaptchaTypeClickWord),
		Token:               token,
	}, nil
}

func (this *ClickWord) Check(token, pointJson string) (*RespMsg, error) {
	var cachedWord []FontPoint
	var checkedWord []FontPoint
	ttl, err := redis.ExecRedisCommand("TTL", token)
	if err != nil {
		return nil, err
	}
	if ttl.(int64) <= 0 {
		_, err = redis.ExecRedisCommand("DEL", token)
		return nil, errors.New("验证码已过期,请刷新重试")
	}
	//Redis里面存在的数据
	cachedBuff, err := mredis.Bytes(redis.ExecRedisCommand("GET", token))
	if err != nil {
		return nil, errors.New("get captcha error:" + err.Error())
	}
	base64Buff, err := base64.StdEncoding.DecodeString(string(cachedBuff))
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	err = json.Unmarshal(base64Buff, &cachedWord)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	//待校验数据
	base64Buff, err = base64.StdEncoding.DecodeString(pointJson)
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	err = json.Unmarshal(base64Buff, &checkedWord)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	if len(cachedWord) != len(checkedWord) {
		return nil, errors.New("验证码有误")
	}
	success := true
	msg := "验证通过"
	fontSize := captchaConf.FontSize
	for index, word := range cachedWord {
		if !(((checkedWord)[index].X >= word.X && (checkedWord)[index].X <= word.X+fontSize) &&
			((checkedWord)[index].Y >= word.Y && (checkedWord)[index].Y <= word.Y+fontSize) &&
			((checkedWord)[index].Text == word.Text)) {
			msg = "验证失败"
			success = false
		}
	}
	//验证后将缓存删除，同一个验证码只能用于验证一次
	_, err = redis.ExecRedisCommand("DEL", token)
	if err != nil {
		log.Printf("验证码缓存删除失败:%s", token)
	}
	return &RespMsg{Success: success, Msg: msg}, nil
}

func (this *ClickWord) randomNoCheck(words []rune) []string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(captchaConf.ClickWordCount)
	var result []string
	for i, v := range words {
		if i == index {
			continue
		}
		result = append(result, string(v))
	}
	return result
}

func (this *ClickWord) randomHanZi() ([]rune, error) {
	words, err := this.initWords()
	if err != nil {
		return nil, err
	}
	wordRunes := []rune(words)
	var result []rune
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < captchaConf.ClickWordCount; i++ {
		index := rand.Intn(len(wordRunes))
		result = append(result, wordRunes[index])
	}
	return result, nil
}

func (this *ClickWord) initWords() (string, error) {
	license, err := os.Open(captchaConf.ClickWordFile)
	if err != nil {
		return "", err
	}
	licenseByte, err := ioutil.ReadAll(license)
	if err != nil {
		return "", err
	}
	licenseStr := string(licenseByte)
	licenseStr = strings.ReplaceAll(licenseStr, "\n", "")
	licenseStr = strings.ReplaceAll(licenseStr, "\r", "")
	licenseStr = strings.ReplaceAll(licenseStr, " ", "")
	return licenseStr, nil
}
