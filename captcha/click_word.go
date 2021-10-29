package captcha

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/yasin-wu/captcha/redis"

	"github.com/golang/freetype"
)

type ClickWord struct {
	imagePath     string  //点选校验图片目录
	wordFile      string  //点选文字文件
	wordCount     int     //点选文字个数
	fontFile      string  //字体文件
	fontSize      int     //字体大小
	watermarkText string  //水印信息
	watermarkSize int     //水印大小
	dpi           float64 //分辨率
	expireTime    int     //校验过期时间
}

type FontPoint struct {
	X    int
	Y    int
	Text string
}

var _ Captcha = (*ClickWord)(nil)

func (this *ClickWord) Get(token string) (*CaptchaVO, error) {
	oriImage, err := NewImage(this.imagePath)
	if err != nil {
		return nil, errors.New("new image error:" + err.Error())
	}
	staticImg := oriImage.Image
	fileType := oriImage.FileType
	img := image2RGBA(staticImg)
	if img == nil {
		return nil, errors.New("image to rgba failed")
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	err = drawText(img, this.watermarkText, this.fontFile, this.watermarkSize, this.dpi)
	if err != nil {
		return nil, errors.New("draw watermark failed:" + err.Error())
	}

	fontBytes, err := ioutil.ReadFile(this.fontFile)
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
	fontSize := this.fontSize
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

	//saveImage("/Users/yasin/tmp.png", "png", img)

	//校验数据存入Redis,存入时进行base64
	err = setRedis(token, allDots, this.expireTime)
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
	//Redis里面存在的数据
	cachedBuff, err := getRedis(token)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cachedBuff, &cachedWord)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	//待校验数据
	base64Buff, err := base64.StdEncoding.DecodeString(pointJson)
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
	fontSize := this.fontSize
	for index, word := range cachedWord {
		if !(((checkedWord)[index].X >= word.X && (checkedWord)[index].X <= word.X+fontSize) &&
			((checkedWord)[index].Y >= word.Y && (checkedWord)[index].Y <= word.Y+fontSize) &&
			((checkedWord)[index].Text == word.Text)) {
			msg = "验证失败"
			success = false
		}
	}
	//验证后将缓存删除，同一个验证码只能用于验证一次
	_, err = redis.RedisClient.Exec("DEL", token)
	if err != nil {
		log.Printf("验证码缓存删除失败:%s", token)
	}
	return &RespMsg{Success: success, Message: msg}, nil
}

func (this *ClickWord) randomNoCheck(words []rune) []string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(this.wordCount)
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
	for i := 0; i < this.wordCount; i++ {
		index := rand.Intn(len(wordRunes))
		result = append(result, wordRunes[index])
	}
	return result, nil
}

func (this *ClickWord) initWords() (string, error) {
	license, err := os.Open(this.wordFile)
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
