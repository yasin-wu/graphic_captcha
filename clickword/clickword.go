package clickword

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/yasin-wu/graphic_captcha/v2/consts"

	"github.com/yasin-wu/graphic_captcha/v2/redis"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype/truetype"

	"github.com/yasin-wu/graphic_captcha/v2/util"

	"github.com/yasin-wu/graphic_captcha/v2/entity"

	"github.com/golang/freetype"

	"github.com/yasin-wu/graphic_captcha/v2/factory"
	image2 "github.com/yasin-wu/graphic_captcha/v2/image"
)

type Point struct {
	X int
	Y int
	T string
}

type clickWord struct {
	clickImagePath string
	clickWordFile  string
	clickWordCount int
	fontFile       string
	fontSize       int
	watermarkText  string
	watermarkSize  int
	dpi            float64
	expireTime     time.Duration
	redisCli       *redis.Client
}

var _ factory.Captcha = (*clickWord)(nil)

func New(redisCli *redis.Client, config factory.Config) *clickWord {
	return &clickWord{
		clickImagePath: config.ClickImagePath,
		clickWordFile:  config.ClickWordFile,
		clickWordCount: config.ClickWordCount,
		fontFile:       config.FontFile,
		fontSize:       config.FontSize,
		watermarkText:  config.WatermarkText,
		watermarkSize:  config.WatermarkSize,
		dpi:            config.DPI,
		expireTime:     config.ExpireTime,
		redisCli:       redisCli,
	}
}

func (c *clickWord) Get(token string) (*entity.Response, error) {
	oriImage, err := image2.New(c.clickImagePath)
	if err != nil {
		return nil, errors.New("new image error:" + err.Error())
	}
	staticImg := oriImage.Image
	fileType := oriImage.FileType
	img := util.Image2RGBA(staticImg)
	if img == nil {
		return nil, errors.New("image to rgba failed")
	}
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	err = util.DrawText(img, c.watermarkText, c.fontFile, c.watermarkSize, c.dpi)
	if err != nil {
		return nil, errors.New("draw watermark failed:" + err.Error())
	}

	str, err := c.randomHanZi()
	if err != nil {
		return nil, errors.New("randomHanZi error:" + err.Error())
	}
	var allDots []Point
	clickWords := c.randomNoCheck(str)
	fontSize := c.fontSize
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len(str); i++ {
		_w := (width - 24) / len(str)
		x := i*_w + rand.Intn(_w-fontSize)             //nolint:gosec
		y := rand.Intn(height - fontSize - fontSize/2) //nolint:gosec
		text := fmt.Sprintf("%c", str[i])
		err := c.drawTextOnBackground(img, image.Pt(x, y), text)
		if err != nil {
			continue
		}
		if c.stringsContains(clickWords, text) {
			allDots = append(allDots, Point{x, y, text})
		}
	}
	base64, err := util.Image2Base64(img, fileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}

	// util.SaveImage("/Users/yasin/tmp.png", "png", img)

	err = c.redisCli.Set(token, allDots, c.expireTime)
	if err != nil {
		return nil, err
	}
	resp := &entity.Response{
		Status:  200,
		Message: "OK",
		Data: entity.Captcha{
			Token:      token,
			Type:       string(consts.CaptchaTypeClickWord),
			OriImage:   base64,
			ClickWords: clickWords,
		},
	}
	return resp, nil
}

func (c *clickWord) Check(token, pointJSON string) (*entity.Response, error) {
	var cachedWord []Point
	var checkedWord []Point
	cachedBuff, err := c.redisCli.Get(token)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cachedBuff, &cachedWord)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	base64Buff, err := base64.StdEncoding.DecodeString(pointJSON)
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	err = json.Unmarshal(base64Buff, &checkedWord)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	if len(cachedWord) != len(checkedWord) {
		return nil, errors.New("???????????????")
	}
	status := 200
	msg := "????????????"
	fontSize := c.fontSize
	for index, word := range cachedWord {
		if !(((checkedWord)[index].X >= word.X && (checkedWord)[index].X <= word.X+fontSize) &&
			((checkedWord)[index].Y >= word.Y && (checkedWord)[index].Y <= word.Y+fontSize) &&
			((checkedWord)[index].T == word.T)) {
			msg = "????????????"
			status = 201
		}
	}
	err = c.redisCli.Del(token)
	if err != nil {
		log.Printf("???????????????????????????:%s", token)
	}
	return &entity.Response{Status: status, Message: msg}, nil
}

func (c *clickWord) randomNoCheck(words []rune) []string {
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(c.clickWordCount) //nolint:gosec
	var result []string                  //nolint:prealloc
	for i, v := range words {
		if i == index {
			continue
		}
		result = append(result, string(v))
	}
	return result
}

func (c *clickWord) randomHanZi() ([]rune, error) {
	words, err := c.initWords()
	if err != nil {
		return nil, err
	}
	wordRunes := []rune(words)
	var result []rune
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < c.clickWordCount; i++ {
		index := rand.Intn(len(wordRunes)) //nolint:gosec
		result = append(result, wordRunes[index])
	}
	return result, nil
}

func (c *clickWord) initWords() (string, error) {
	license, err := os.Open(c.clickWordFile)
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

func (c *clickWord) stringsContains(stringArray []string, substr string) bool {
	for _, v := range stringArray {
		if v == substr {
			return true
		}
	}
	return false
}

func (c *clickWord) drawTextOnBackground(bg draw.Image, pt image.Point, text string) error {
	fontBytes, err := ioutil.ReadFile(c.fontFile)
	if err != nil {
		return errors.New("read font file error:" + err.Error())
	}
	fontStyle, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return errors.New("parse font error:" + err.Error())
	}
	angle := float64(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(40) - 20) //nolint:gosec
	fontPng := c.drawString2Png(fontStyle, text)
	rotate := imaging.Rotate(fontPng, angle, color.Transparent)
	resize := imaging.Resize(rotate, c.fontSize, c.fontSize, imaging.Lanczos)
	resizePng := util.Image2RGBA(resize)
	draw.Draw(bg, image.Rect(pt.X, pt.Y, pt.X+c.fontSize, pt.Y+c.fontSize), resizePng, image.Point{}, draw.Over)
	return nil
}

func (c *clickWord) drawString2Png(font *truetype.Font, str string) *image.RGBA {
	fontColor := image.NewUniform(color.RGBA{
		R: uint8(rand.Intn(255)),      //nolint:gosec
		G: uint8(rand.Intn(255) + 50), //nolint:gosec
		B: uint8(rand.Intn(255)),      //nolint:gosec
		A: uint8(255)})
	img := image.NewRGBA(image.Rect(0, 0, int(c.fontSize), int(c.fontSize))) //nolint:gosec
	ctx := freetype.NewContext()
	ctx.SetDst(img)
	ctx.SetClip(img.Bounds())
	ctx.SetSrc(image.NewUniform(fontColor))
	ctx.SetFontSize(float64(c.fontSize))
	ctx.SetFont(font)
	pt := freetype.Pt(0, int(-c.fontSize/6)+ctx.PointToFixed(float64(c.fontSize)).Ceil()) //nolint:gosec
	ctx.DrawString(str, pt)
	return img
}
