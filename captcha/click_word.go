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

	"github.com/yasin-wu/graphic_captcha/common"

	"github.com/golang/freetype"
)

type ClickWord struct {
	conf *config
}

type Point struct {
	X int
	Y int
	T string
}

var _ Engine = (*ClickWord)(nil)

func (this *ClickWord) Get(token string) (*common.Captcha, error) {
	oriImage, err := common.NewImage(this.conf.clickImagePath)
	if err != nil {
		return nil, errors.New("new image error:" + err.Error())
	}
	staticImg := oriImage.Image
	fileType := oriImage.FileType
	img := common.Image2RGBA(staticImg)
	if img == nil {
		return nil, errors.New("image to rgba failed")
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	err = common.DrawText(img, this.conf.watermarkText, this.conf.fontFile, this.conf.watermarkSize, this.conf.dpi)
	if err != nil {
		return nil, errors.New("draw watermark failed:" + err.Error())
	}

	fontBytes, err := ioutil.ReadFile(this.conf.fontFile)
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
	var allDots []Point
	clickWords := this.randomNoCheck(str)
	fontSize := this.conf.fontSize
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len(str); i++ {
		_w := (width - 24) / len(str)
		x := i*_w + rand.Intn(_w-fontSize)
		y := rand.Intn(height - fontSize - fontSize/2)
		fontColor := image.NewUniform(color.RGBA{R: uint8(rand.Intn(255)), G: uint8(rand.Intn(255) + 50),
			B: uint8(rand.Intn(255)), A: uint8(255)})
		text := fmt.Sprintf("%c", str[i])
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		angle := float64(r.Intn(40) - 20)
		common.DrawTextOnBackground(img, image.Pt(x, y), font, text, fontColor, fontSize, angle)
		if common.StringsContains(clickWords, text) {
			allDots = append(allDots, Point{x, y, text})
		}
	}
	base64_, err := common.ImgToBase64(img, fileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}

	//saveImage("/Users/yasin/tmp.png", "png", img)

	err = this.conf.redisCli.Set(token, allDots, this.conf.expireTime)
	if err != nil {
		return nil, err
	}
	return &common.Captcha{
		OriImage:   base64_,
		ClickWords: clickWords,
		Type:       string(common.CaptchaTypeClickWord),
		Token:      token,
	}, nil
}

func (this *ClickWord) Check(token, pointJson string) (*common.RespMsg, error) {
	var cachedWord []Point
	var checkedWord []Point
	cachedBuff, err := this.conf.redisCli.Get(token)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cachedBuff, &cachedWord)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
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
	status := 200
	msg := "验证通过"
	fontSize := this.conf.fontSize
	for index, word := range cachedWord {
		if !(((checkedWord)[index].X >= word.X && (checkedWord)[index].X <= word.X+fontSize) &&
			((checkedWord)[index].Y >= word.Y && (checkedWord)[index].Y <= word.Y+fontSize) &&
			((checkedWord)[index].T == word.T)) {
			msg = "验证失败"
			status = 201
		}
	}
	err = this.conf.redisCli.Client.Del(token)
	if err != nil {
		log.Printf("验证码缓存删除失败:%s", token)
	}
	return &common.RespMsg{Status: status, Message: msg}, nil
}

func (this *ClickWord) randomNoCheck(words []rune) []string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(this.conf.clickWordCount)
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
	for i := 0; i < this.conf.clickWordCount; i++ {
		index := rand.Intn(len(wordRunes))
		result = append(result, wordRunes[index])
	}
	return result, nil
}

func (this *ClickWord) initWords() (string, error) {
	license, err := os.Open(this.conf.clickWordFile)
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
