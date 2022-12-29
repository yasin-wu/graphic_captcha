package slideblock

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/config"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/consts"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/entity"
	"github.com/yasin-wu/graphic_captcha/v2/pkg/factory"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	image2 "github.com/yasin-wu/graphic_captcha/v2/internal/image"
	"github.com/yasin-wu/graphic_captcha/v2/internal/redis"
	"github.com/yasin-wu/graphic_captcha/v2/internal/util"
)

type slideBlock struct {
	originalPath         string
	blockPath            string
	threshold            float64
	blur                 float64
	brightness           float64
	fontFile             string
	watermarkText        string
	watermarkSize        int
	dpi                  float64
	transparentThreshold uint32
	expireTime           time.Duration
	redisCli             *redis.Client
}

var _ factory.Captchaer = (*slideBlock)(nil)

func New(redisCli *redis.Client, config config.Config) factory.Captchaer {
	return &slideBlock{
		originalPath:         config.OriginalPath,
		blockPath:            config.BlockPath,
		threshold:            config.Threshold,
		blur:                 config.Blur,
		brightness:           config.Brightness,
		fontFile:             config.FontFile,
		watermarkText:        config.WatermarkText,
		watermarkSize:        config.WatermarkSize,
		dpi:                  config.DPI,
		transparentThreshold: 150 << 8,
		expireTime:           config.ExpireTime,
		redisCli:             redisCli,
	}
}

//nolint:funlen
func (c *slideBlock) Get(token string) (*entity.Response, error) {
	oriImg, err := image2.New(c.originalPath)
	if err != nil {
		return nil, err
	}
	blockImg, err := image2.New(c.blockPath)
	if err != nil {
		return nil, err
	}
	oriRGBA := util.Image2RGBA(oriImg.Image)
	blockWidth := blockImg.Image.Bounds().Dx()
	blockHeight := blockImg.Image.Bounds().Dy()
	point := c.generateJigsawPoint(oriImg.Image.Bounds().Dx(), oriImg.Image.Bounds().Dy(), blockWidth, blockHeight)
	if err = util.DrawText(oriRGBA, c.watermarkText, c.fontFile, c.watermarkSize, c.dpi); err != nil {
		return nil, err
	}
	c.interfereBlock(oriRGBA, point, blockImg.FileName)
	jigsaw := c.cropJigsaw(blockImg.Image, oriImg.Image, point)
	blur := imaging.Blur(jigsaw, c.blur)
	blur = imaging.AdjustBrightness(blur, c.brightness)
	blurRGB := util.Image2RGBA(blur)
	newImage := image.NewRGBA(blockImg.Image.Bounds())
	for x := 0; x < blockWidth; x++ {
		for y := 0; y < blockHeight; y++ {
			_, _, _, a := blockImg.Image.At(x, y).RGBA()
			if a > c.transparentThreshold {
				r, g, b, _ := oriRGBA.At(point.X+x, point.Y+y).RGBA()
				newImage.Set(x, y, c.colorTransparent(r, g, b, a))
				oriRGBA.Set(x+point.X, y+point.Y,
					c.colorMix((blurRGB.At(x, y)).(color.RGBA),
						(oriRGBA.At(x+point.X, y+point.Y)).(color.RGBA)))
			}
			if x == (blockWidth-1) || y == (blockHeight-1) {
				continue
			}
			_, _, _, ra := blockImg.Image.At(x+1, y).RGBA()
			_, _, _, da := blockImg.Image.At(x, y+1).RGBA()
			if c.isThreshold(a, ra, da) {
				mix := c.colorMix(color.RGBA{R: 255, G: 255, B: 255, A: 220}, (oriRGBA.At(point.X+x, point.Y+y)).(color.RGBA))
				newImage.Set(x, y, color.White)
				oriRGBA.Set(point.X+x, point.Y+y, mix)
			}
		}
	}
	oriBase64, err := util.Image2Base64(oriRGBA, oriImg.FileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}
	blockBase64, err := util.Image2Base64(newImage, blockImg.FileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}

	//util.SaveImage("/Users/yasin/tmp.png", "png", oriRGBA)
	//util.SaveImage("/Users/yasin/block.png", "png", newImage)

	if err = c.redisCli.Set(token, point, c.expireTime); err != nil {
		return nil, err
	}
	resp := &entity.Response{
		Status:  200,
		Message: "OK",
		Data: entity.CaptchaDO{
			Token:       token,
			Type:        string(consts.CaptchaTypeBlockPuzzle),
			OriImage:    oriBase64,
			JigsawImage: blockBase64,
		},
	}
	return resp, nil
}

func (c *slideBlock) Check(token, pointJSON string) (*entity.Response, error) {
	var cachedPoint image.Point
	var checkedPoint image.Point

	cachedBuff, err := c.redisCli.Get(token)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(cachedBuff, &cachedPoint); err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	base64Buff, err := base64.StdEncoding.DecodeString(pointJSON)
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	if err = json.Unmarshal(base64Buff, &checkedPoint); err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	status := 201
	msg := "验证失败"
	if (math.Abs(float64(cachedPoint.X-checkedPoint.X)) <= c.threshold) &&
		(math.Abs(float64(cachedPoint.Y-checkedPoint.Y)) <= c.threshold) {
		status = 200
		msg = "验证通过"
	}

	if err = c.redisCli.Del(token); err != nil {
		log.Printf("验证码缓存删除失败:%v", token)
	}
	return &entity.Response{Status: status, Message: msg}, nil
}

func (c *slideBlock) interfereBlock(img *image.RGBA, point image.Point, srcBlockName string) {
	blockPath := c.blockPath
	var blockName1 string
	for {
		blockName1, _ = util.RandomFileName(blockPath)
		if blockName1 != srcBlockName {
			break
		}
	}
	blockFileName1 := blockPath + "/" + blockName1
	c.doInterfere(blockFileName1, img, point, 1)

	var blockName2 string
	for {
		blockName2, _ = util.RandomFileName(blockPath)
		if blockName2 != srcBlockName && blockName2 != blockName1 {
			break
		}
	}
	blockFileName2 := blockPath + "/" + blockName2
	c.doInterfere(blockFileName2, img, point, 2)
}

func (c *slideBlock) doInterfere(blockFileName string, img *image.RGBA, point image.Point, _type int) {
	blockFile, err := os.Open(blockFileName)
	if err != nil {
		log.Printf("open file error: %v", err)
		return
	}
	defer func(blockFile *os.File) {
		_ = blockFile.Close()
	}(blockFile)
	blockImg, _ := png.Decode(blockFile)
	originalWidth := img.Bounds().Dx()
	jigsawWidth := blockImg.Bounds().Dx()
	x := point.X
	position := 0
	switch _type {
	case 1:
		if originalWidth-x-5 > jigsawWidth*2 {
			position = util.RandInt(x+jigsawWidth+5, originalWidth-jigsawWidth)
		} else {
			position = util.RandInt(100, x-jigsawWidth-5)
		}
	case 2:
		position = util.RandInt(jigsawWidth, 100-jigsawWidth)
	}
	point = blockImg.Bounds().Min.Sub(image.Pt(-position, 0))
	jigsaw := c.cropJigsaw(blockImg, img, point)
	blur := imaging.Blur(jigsaw, c.blur)
	blur = imaging.AdjustBrightness(blur, c.brightness)
	blurRGB := util.Image2RGBA(blur)

	for x := 0; x < blockImg.Bounds().Dx(); x++ {
		for y := 0; y < blockImg.Bounds().Dy(); y++ {
			_, _, _, a := blockImg.At(x, y).RGBA()
			if a > c.transparentThreshold {
				img.Set(x+point.X, y+point.Y,
					c.colorMix((blurRGB.At(x, y)).(color.RGBA),
						(img.At(x+point.X, y+point.Y)).(color.RGBA)))
			}
			if x == (blockImg.Bounds().Dx()-1) || y == (blockImg.Bounds().Dy()-1) {
				continue
			}
			_, _, _, ra := blockImg.At(x+1, y).RGBA()
			_, _, _, da := blockImg.At(x, y+1).RGBA()
			if c.isThreshold(a, ra, da) {
				mix := c.colorMix(color.RGBA{R: 255, G: 255, B: 255, A: 220}, (img.At(point.X+x, point.Y+y)).(color.RGBA))
				img.Set(point.X+x, point.Y+y, mix)
			}
		}
	}
}

func (c *slideBlock) cropJigsaw(blockImg, oriImg image.Image, point image.Point) *image.RGBA {
	newImage := image.NewRGBA(blockImg.Bounds())
	for x := 0; x < blockImg.Bounds().Dx(); x++ {
		for y := 0; y < blockImg.Bounds().Dy(); y++ {
			_, _, _, a := blockImg.At(x, y).RGBA()
			if a > c.transparentThreshold {
				r, g, b, _ := oriImg.At(point.X+x, point.Y+y).RGBA()
				newImage.Set(x, y, c.colorTransparent(r, g, b, a))
			}
			if x == (blockImg.Bounds().Dx()-1) || y == (blockImg.Bounds().Dy()-1) {
				continue
			}
		}
	}
	return newImage
}

func (c *slideBlock) colorTransparent(r, g, b, a uint32) color.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, 1, 1))
	convert := rgba.ColorModel().Convert(color.NRGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: uint16(a)})
	rr, gg, bb, aa := convert.RGBA()
	return color.RGBA{R: uint8(rr), G: uint8(gg), B: uint8(bb), A: uint8(aa)}
}

func (c *slideBlock) colorMix(fg, bg color.RGBA) color.RGBA {
	rgba := color.RGBA{}
	fa := c.floatRound(float64(fg.A)/255, 2)
	ba := c.floatRound(float64(bg.A)/255, 2)
	alpha := 1 - (1-fa)*(1-ba)
	rgba.R = uint8((float64(fg.R)*fa + float64(bg.R)*ba*(1-fa)) / alpha)
	rgba.G = uint8((float64(fg.G)*fa + float64(bg.G)*ba*(1-fa)) / alpha)
	rgba.B = uint8((float64(fg.B)*fa + float64(bg.B)*ba*(1-fa)) / alpha)
	rgba.A = uint8(alpha * 255)
	return rgba
}

func (c *slideBlock) floatRound(f float64, n int) float64 {
	format := "%." + strconv.Itoa(n) + "f"
	res, _ := strconv.ParseFloat(fmt.Sprintf(format, f), 64)
	return res
}

func (c *slideBlock) generateJigsawPoint(originalWidth, originalHeight, jigsawWidth, jigsawHeight int) image.Point {
	rand.Seed(time.Now().UnixNano())
	widthDifference := originalWidth - jigsawWidth
	heightDifference := originalHeight - jigsawHeight
	x, y := 0, 0
	if widthDifference <= 0 {
		x = 5
	} else {
		x = rand.Intn(originalWidth-jigsawWidth-100) + 100 //nolint:gosec
	}
	if heightDifference <= 0 {
		y = 5
	} else {
		y = rand.Intn(originalWidth-jigsawWidth) + 5 //nolint:gosec
	}
	return image.Point{X: x, Y: y}
}

func (c *slideBlock) isThreshold(a, ra, da uint32) bool {
	return (a > c.transparentThreshold && ra <= c.transparentThreshold) ||
		(a <= c.transparentThreshold && ra > c.transparentThreshold) ||
		(a > c.transparentThreshold && da <= c.transparentThreshold) ||
		(a <= c.transparentThreshold && da > c.transparentThreshold)
}
