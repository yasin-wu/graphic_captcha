package captcha

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/yasin-wu/graphic_captcha/common"

	"github.com/disintegration/imaging"
)

type SlideBlock struct {
	conf *config
}

var _ Engine = (*SlideBlock)(nil)

func (this *SlideBlock) Get(token string) (*common.Captcha, error) {
	oriImg, err := common.NewImage(this.conf.originalPath)
	if err != nil {
		return nil, err
	}
	blockImg, err := common.NewImage(this.conf.blockPath)
	if err != nil {
		return nil, err
	}
	oriRGBA := common.Image2RGBA(oriImg.Image)
	blockWidth := blockImg.Image.Bounds().Dx()
	blockHeight := blockImg.Image.Bounds().Dy()
	point := common.GenerateJigsawPoint(oriImg.Image.Bounds().Dx(), oriImg.Image.Bounds().Dy(), blockWidth, blockHeight)
	err = common.DrawText(oriRGBA, this.conf.watermarkText, this.conf.fontFile, this.conf.watermarkSize, this.conf.dpi)
	if err != nil {
		return nil, err
	}
	this.interfereBlock(oriRGBA, point, blockImg.FileName)
	jigsaw := this.cropJigsaw(blockImg.Image, oriImg.Image, point)
	blur := imaging.Blur(jigsaw, this.conf.blur)
	blur = imaging.AdjustBrightness(blur, this.conf.brightness)
	blurRGB := common.Image2RGBA(blur)

	newImage := image.NewRGBA(blockImg.Image.Bounds())
	for x := 0; x < blockWidth; x++ {
		for y := 0; y < blockHeight; y++ {
			_, _, _, a := blockImg.Image.At(x, y).RGBA()
			if a > common.TransparentThreshold {
				r, g, b, _ := oriRGBA.At(point.X+x, point.Y+y).RGBA()
				newImage.Set(x, y, common.ColorTransparent(r, g, b, a))
				oriRGBA.Set(x+point.X, y+point.Y, common.ColorMix((blurRGB.At(x, y)).(color.RGBA),
					(oriRGBA.At(x+point.X, y+point.Y)).(color.RGBA)))
			}
			if x == (blockWidth-1) || y == (blockHeight-1) {
				continue
			}
			_, _, _, ra := blockImg.Image.At(x+1, y).RGBA()
			_, _, _, da := blockImg.Image.At(x, y+1).RGBA()
			if (a > common.TransparentThreshold && ra <= common.TransparentThreshold) ||
				(a <= common.TransparentThreshold && ra > common.TransparentThreshold) ||
				(a > common.TransparentThreshold && da <= common.TransparentThreshold) ||
				(a <= common.TransparentThreshold && da > common.TransparentThreshold) {
				mix := common.ColorMix(color.RGBA{R: 255, G: 255, B: 255, A: 220}, (oriRGBA.At(point.X+x, point.Y+y)).(color.RGBA))
				newImage.Set(x, y, color.White)
				oriRGBA.Set(point.X+x, point.Y+y, mix)
			}
		}
	}
	oriBase64, err := common.ImgToBase64(oriRGBA, oriImg.FileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}
	blockBase64, err := common.ImgToBase64(newImage, blockImg.FileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}

	//saveImage("/Users/yasin/tmp.png", "png", oriRGBA)
	//saveImage("/Users/yasin/block.png", "png", newImage)

	err = this.conf.redisCli.Set(token, point, this.conf.expireTime)
	if err != nil {
		return nil, err
	}
	return &common.Captcha{
		Token:       token,
		Type:        string(common.CaptchaTypeBlockPuzzle),
		OriImage:    oriBase64,
		JigsawImage: blockBase64,
	}, nil
}

func (this *SlideBlock) Check(token, pointJson string) (*common.RespMsg, error) {
	var cachedPoint image.Point
	var checkedPoint image.Point

	cachedBuff, err := this.conf.redisCli.Get(token)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cachedBuff, &cachedPoint)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	base64Buff, err := base64.StdEncoding.DecodeString(pointJson)
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	err = json.Unmarshal(base64Buff, &checkedPoint)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}

	status := 201
	msg := "验证失败"
	if (math.Abs(float64(cachedPoint.X-checkedPoint.X)) <= this.conf.threshold) &&
		(math.Abs(float64(cachedPoint.Y-checkedPoint.Y)) <= this.conf.threshold) {
		status = 200
		msg = "验证通过"
	}

	err = this.conf.redisCli.Client.Del(token)
	if err != nil {
		log.Printf("验证码缓存删除失败:%s", token)
	}
	return &common.RespMsg{Status: status, Message: msg}, nil
}

func (this *SlideBlock) interfereBlock(img *image.RGBA, point image.Point, srcBlockName string) {
	blockPath := this.conf.blockPath
	var blockName1 string
	for {
		blockName1, _ = common.RandomFileName(blockPath)
		if blockName1 != srcBlockName {
			break
		}
	}
	blockFileName1 := blockPath + "/" + blockName1
	this.doInterfere(blockFileName1, img, point, 1)

	var blockName2 string
	for {
		blockName2, _ = common.RandomFileName(blockPath)
		if blockName2 != srcBlockName && blockName2 != blockName1 {
			break
		}
	}
	blockFileName2 := blockPath + "/" + blockName2
	this.doInterfere(blockFileName2, img, point, 2)
}

func (this *SlideBlock) doInterfere(blockFileName string, img *image.RGBA, point image.Point, _type int) {
	blockFile, err := os.Open(blockFileName)
	if err != nil {
		log.Printf("open file error: %v", err)
		return
	}
	defer blockFile.Close()
	blockImg, _ := png.Decode(blockFile)
	originalWidth := img.Bounds().Dx()
	jigsawWidth := blockImg.Bounds().Dx()
	x := point.X
	position := 0
	switch _type {
	case 1:
		if originalWidth-x-5 > jigsawWidth*2 {
			position = common.RandInt(x+jigsawWidth+5, originalWidth-jigsawWidth)
		} else {
			position = common.RandInt(100, x-jigsawWidth-5)
		}
	case 2:
		position = common.RandInt(jigsawWidth, 100-jigsawWidth)
	}
	point = blockImg.Bounds().Min.Sub(image.Pt(-position, 0))
	jigsaw := this.cropJigsaw(blockImg, img, point)
	blur := imaging.Blur(jigsaw, this.conf.blur)
	blur = imaging.AdjustBrightness(blur, this.conf.brightness)
	blurRGB := common.Image2RGBA(blur)

	for x := 0; x < blockImg.Bounds().Dx(); x++ {
		for y := 0; y < blockImg.Bounds().Dy(); y++ {
			_, _, _, a := blockImg.At(x, y).RGBA()
			if a > common.TransparentThreshold {
				img.Set(x+point.X, y+point.Y, common.ColorMix((blurRGB.At(x, y)).(color.RGBA), (img.At(x+point.X, y+point.Y)).(color.RGBA)))
			}
			if x == (blockImg.Bounds().Dx()-1) || y == (blockImg.Bounds().Dy()-1) {
				continue
			}
			_, _, _, ra := blockImg.At(x+1, y).RGBA()
			_, _, _, da := blockImg.At(x, y+1).RGBA()
			if (a > common.TransparentThreshold && ra <= common.TransparentThreshold) ||
				(a <= common.TransparentThreshold && ra > common.TransparentThreshold) ||
				(a > common.TransparentThreshold && da <= common.TransparentThreshold) ||
				(a <= common.TransparentThreshold && da > common.TransparentThreshold) {
				mix := common.ColorMix(color.RGBA{R: 255, G: 255, B: 255, A: 220}, (img.At(point.X+x, point.Y+y)).(color.RGBA))
				img.Set(point.X+x, point.Y+y, mix)
			}
		}
	}
}

func (this *SlideBlock) cropJigsaw(blockImg, oriImg image.Image, point image.Point) *image.RGBA {
	newImage := image.NewRGBA(blockImg.Bounds())
	for x := 0; x < blockImg.Bounds().Dx(); x++ {
		for y := 0; y < blockImg.Bounds().Dy(); y++ {
			_, _, _, a := blockImg.At(x, y).RGBA()
			if a > common.TransparentThreshold {
				r, g, b, _ := oriImg.At(point.X+x, point.Y+y).RGBA()
				newImage.Set(x, y, common.ColorTransparent(r, g, b, a))
			}
			if x == (blockImg.Bounds().Dx()-1) || y == (blockImg.Bounds().Dy()-1) {
				continue
			}
		}
	}
	return newImage
}
