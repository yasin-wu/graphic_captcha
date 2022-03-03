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

	"github.com/disintegration/imaging"
)

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:19
 * @description: 滑块验证
 */
type SlideBlock struct {
	conf *config
}

var _ Engine = (*SlideBlock)(nil)

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:19
 * @params: token string
 * @return: *Captcha, error
 * @description: 获取滑块待验证信息
 */
func (s *SlideBlock) Get(token string) (*Captcha, error) {
	oriImg, err := newImage(s.conf.originalPath)
	if err != nil {
		return nil, err
	}
	blockImg, err := newImage(s.conf.blockPath)
	if err != nil {
		return nil, err
	}
	oriRGBA := image2RGBA(oriImg.Image)
	blockWidth := blockImg.Image.Bounds().Dx()
	blockHeight := blockImg.Image.Bounds().Dy()
	point := generateJigsawPoint(oriImg.Image.Bounds().Dx(), oriImg.Image.Bounds().Dy(), blockWidth, blockHeight)
	err = drawText(oriRGBA, s.conf.watermarkText, s.conf.fontFile, s.conf.watermarkSize, s.conf.dpi)
	if err != nil {
		return nil, err
	}
	s.interfereBlock(oriRGBA, point, blockImg.FileName)
	jigsaw := s.cropJigsaw(blockImg.Image, oriImg.Image, point)
	blur := imaging.Blur(jigsaw, s.conf.blur)
	blur = imaging.AdjustBrightness(blur, s.conf.brightness)
	blurRGB := image2RGBA(blur)

	newImage := image.NewRGBA(blockImg.Image.Bounds())
	for x := 0; x < blockWidth; x++ {
		for y := 0; y < blockHeight; y++ {
			_, _, _, a := blockImg.Image.At(x, y).RGBA()
			if a > TransparentThreshold {
				r, g, b, _ := oriRGBA.At(point.X+x, point.Y+y).RGBA()
				newImage.Set(x, y, colorTransparent(r, g, b, a))
				oriRGBA.Set(x+point.X, y+point.Y, colorMix((blurRGB.At(x, y)).(color.RGBA),
					(oriRGBA.At(x+point.X, y+point.Y)).(color.RGBA)))
			}
			if x == (blockWidth-1) || y == (blockHeight-1) {
				continue
			}
			_, _, _, ra := blockImg.Image.At(x+1, y).RGBA()
			_, _, _, da := blockImg.Image.At(x, y+1).RGBA()
			if (a > TransparentThreshold && ra <= TransparentThreshold) ||
				(a <= TransparentThreshold && ra > TransparentThreshold) ||
				(a > TransparentThreshold && da <= TransparentThreshold) ||
				(a <= TransparentThreshold && da > TransparentThreshold) {
				mix := colorMix(color.RGBA{R: 255, G: 255, B: 255, A: 220}, (oriRGBA.At(point.X+x, point.Y+y)).(color.RGBA))
				newImage.Set(x, y, color.White)
				oriRGBA.Set(point.X+x, point.Y+y, mix)
			}
		}
	}
	oriBase64, err := image2Base64(oriRGBA, oriImg.FileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}
	blockBase64, err := image2Base64(newImage, blockImg.FileType)
	if err != nil {
		return nil, errors.New("image to base64 error:" + err.Error())
	}

	//saveImage("/Users/yasin/tmp.png", "png", oriRGBA)
	//saveImage("/Users/yasin/block.png", "png", newImage)

	err = s.conf.redisCli.Set(token, point, s.conf.expireTime)
	if err != nil {
		return nil, err
	}
	return &Captcha{
		Token:       token,
		Type:        string(CaptchaTypeBlockPuzzle),
		OriImage:    oriBase64,
		JigsawImage: blockBase64,
	}, nil
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:20
 * @params: token, pointJson string;pointJson为滑块图片base64值
 * @return: *RespMsg, error
 * @description: 校验用户操作结果
 */
func (s *SlideBlock) Check(token, pointJson string) (*RespMsg, error) {
	var cachedPoint image.Point
	var checkedPoint image.Point

	cachedBuff, err := s.conf.redisCli.Get(token)
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
	if (math.Abs(float64(cachedPoint.X-checkedPoint.X)) <= s.conf.threshold) &&
		(math.Abs(float64(cachedPoint.Y-checkedPoint.Y)) <= s.conf.threshold) {
		status = 200
		msg = "验证通过"
	}

	err = s.conf.redisCli.client.Del(s.conf.redisCli.ctx, token).Err()
	if err != nil {
		log.Printf("验证码缓存删除失败:%s", token)
	}
	return &RespMsg{Status: status, Message: msg}, nil
}

func (s *SlideBlock) interfereBlock(img *image.RGBA, point image.Point, srcBlockName string) {
	blockPath := s.conf.blockPath
	var blockName1 string
	for {
		blockName1, _ = randomFileName(blockPath)
		if blockName1 != srcBlockName {
			break
		}
	}
	blockFileName1 := blockPath + "/" + blockName1
	s.doInterfere(blockFileName1, img, point, 1)

	var blockName2 string
	for {
		blockName2, _ = randomFileName(blockPath)
		if blockName2 != srcBlockName && blockName2 != blockName1 {
			break
		}
	}
	blockFileName2 := blockPath + "/" + blockName2
	s.doInterfere(blockFileName2, img, point, 2)
}

func (s *SlideBlock) doInterfere(blockFileName string, img *image.RGBA, point image.Point, _type int) {
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
			position = randInt(x+jigsawWidth+5, originalWidth-jigsawWidth)
		} else {
			position = randInt(100, x-jigsawWidth-5)
		}
	case 2:
		position = randInt(jigsawWidth, 100-jigsawWidth)
	}
	point = blockImg.Bounds().Min.Sub(image.Pt(-position, 0))
	jigsaw := s.cropJigsaw(blockImg, img, point)
	blur := imaging.Blur(jigsaw, s.conf.blur)
	blur = imaging.AdjustBrightness(blur, s.conf.brightness)
	blurRGB := image2RGBA(blur)

	for x := 0; x < blockImg.Bounds().Dx(); x++ {
		for y := 0; y < blockImg.Bounds().Dy(); y++ {
			_, _, _, a := blockImg.At(x, y).RGBA()
			if a > TransparentThreshold {
				img.Set(x+point.X, y+point.Y, colorMix((blurRGB.At(x, y)).(color.RGBA), (img.At(x+point.X, y+point.Y)).(color.RGBA)))
			}
			if x == (blockImg.Bounds().Dx()-1) || y == (blockImg.Bounds().Dy()-1) {
				continue
			}
			_, _, _, ra := blockImg.At(x+1, y).RGBA()
			_, _, _, da := blockImg.At(x, y+1).RGBA()
			if (a > TransparentThreshold && ra <= TransparentThreshold) ||
				(a <= TransparentThreshold && ra > TransparentThreshold) ||
				(a > TransparentThreshold && da <= TransparentThreshold) ||
				(a <= TransparentThreshold && da > TransparentThreshold) {
				mix := colorMix(color.RGBA{R: 255, G: 255, B: 255, A: 220}, (img.At(point.X+x, point.Y+y)).(color.RGBA))
				img.Set(point.X+x, point.Y+y, mix)
			}
		}
	}
}

func (s *SlideBlock) cropJigsaw(blockImg, oriImg image.Image, point image.Point) *image.RGBA {
	newImage := image.NewRGBA(blockImg.Bounds())
	for x := 0; x < blockImg.Bounds().Dx(); x++ {
		for y := 0; y < blockImg.Bounds().Dy(); y++ {
			_, _, _, a := blockImg.At(x, y).RGBA()
			if a > TransparentThreshold {
				r, g, b, _ := oriImg.At(point.X+x, point.Y+y).RGBA()
				newImage.Set(x, y, colorTransparent(r, g, b, a))
			}
			if x == (blockImg.Bounds().Dx()-1) || y == (blockImg.Bounds().Dy()-1) {
				continue
			}
		}
	}
	return newImage
}
