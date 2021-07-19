package captcha

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/disintegration/imaging"
	mredis "github.com/gomodule/redigo/redis"
	"github.com/yasin-wu/captcha/redis"
)

type BlockPuzzle struct {
	originalPath  string  //滑块原图目录
	blockPath     string  //滑块抠图目录
	threshold     float64 //滑块容忍的偏差范围
	blur          float64 //滑块空缺的模糊度
	brightness    float64 //滑块空缺亮度
	fontFile      string  //字体文件
	watermarkText string  //水印信息
	watermarkSize int     //水印大小
	dpi           float64 //分辨率
	expireTime    int     //校验过期时间
}

func (this *BlockPuzzle) Get(token string) (*CaptchaVO, error) {
	//拼图原图
	oriImg, err := NewImage(this.originalPath)
	if err != nil {
		return nil, err
	}
	//拼图模板图
	blockImg, err := NewImage(this.blockPath)
	if err != nil {
		return nil, err
	}
	//拼图底图
	oriRGBA := image2RGBA(oriImg.Image)
	//模板图片宽高
	blockWidth := blockImg.Image.Bounds().Dx()
	blockHeight := blockImg.Image.Bounds().Dy()
	//随机拼图块出现的坐标
	p := generateJigsawPoint(oriImg.Image.Bounds().Dx(), oriImg.Image.Bounds().Dy(), blockWidth, blockHeight)
	//绘制水印
	err = drawText(oriRGBA, this.watermarkText, this.fontFile, this.watermarkSize, this.dpi)
	if err != nil {
		return nil, err
	}
	//添加干扰图像
	this.interfereBlock(oriRGBA, p, blockImg.FileName)

	//处理拼图块中模糊部分
	jigsaw := this.cropJigsaw(blockImg.Image, oriImg.Image, p)
	blur := imaging.Blur(jigsaw, this.blur)
	blur = imaging.AdjustBrightness(blur, this.brightness)
	blurRGB := image2RGBA(blur)

	newImage := image.NewRGBA(blockImg.Image.Bounds())
	for x := 0; x < blockWidth; x++ {
		for y := 0; y < blockHeight; y++ {
			// 如果模板图像当前像素点不是透明色 copy源文件信息到目标图片中
			_, _, _, a := blockImg.Image.At(x, y).RGBA()
			if a > TransparentThreshold {
				r, g, b, _ := oriRGBA.At(p.X+x, p.Y+y).RGBA()
				newImage.Set(x, y, colorTransparent(r, g, b, a))
				oriRGBA.Set(x+p.X, y+p.Y, colorMix((blurRGB.At(x, y)).(color.RGBA), (oriRGBA.At(x+p.X, y+p.Y)).(color.RGBA)))
			}
			//防止数组越界判断
			if x == (blockWidth-1) || y == (blockHeight-1) {
				continue
			}
			_, _, _, ra := blockImg.Image.At(x+1, y).RGBA()
			_, _, _, da := blockImg.Image.At(x, y+1).RGBA()
			//描边处理,取带像素和无像素的界点，判断该点是不是临界轮廓点,如果是设置该坐标像素是白色
			//为都抗锯齿，需要提高透明度阈值
			if (a > TransparentThreshold && ra <= TransparentThreshold) ||
				(a <= TransparentThreshold && ra > TransparentThreshold) ||
				(a > TransparentThreshold && da <= TransparentThreshold) ||
				(a <= TransparentThreshold && da > TransparentThreshold) {
				//如果模板当前像素点不透明，但右侧像素点透明
				//如果模板当前像素点透明，但右侧像素点不透明
				//如果模板当前像素点不透明，但下侧侧像素点透明
				//如果模板当前像素点透明，但下侧侧像素点不透明
				mix := colorMix(color.RGBA{R: 255, G: 255, B: 255, A: 220}, (oriRGBA.At(p.X+x, p.Y+y)).(color.RGBA))
				newImage.Set(x, y, color.White)
				oriRGBA.Set(p.X+x, p.Y+y, mix)
			}
		}
	}
	oriBase64, err := imgToBase64(oriRGBA, oriImg.FileType)
	blockBase64, err := imgToBase64(newImage, blockImg.FileType)

	//saveImage("/Users/yasin/tmp.png", oriRGBA, 100)
	//saveImage("/Users/yasin/block.png", newImage, 100)

	//校验数据base64后存入Redis
	pBuff, err := json.Marshal(p)
	if err != nil {
		return nil, errors.New("json marshal error:" + err.Error())
	}
	data64 := base64.StdEncoding.EncodeToString(pBuff)
	fmt.Println(data64)
	err = SetRedis(token, data64, this.expireTime)
	if err != nil {
		return nil, err
	}

	return &CaptchaVO{
		Token:               token,
		CaptchaType:         string(CaptchaTypeBlockPuzzle),
		OriginalImageBase64: oriBase64,
		JigsawImageBase64:   blockBase64,
	}, nil
}

func (this *BlockPuzzle) Check(token, pointJson string) (*RespMsg, error) {
	var cachedPoint image.Point
	var checkedPoint image.Point
	ttl, err := redis.ExecRedisCommand("TTL", token)
	if err != nil {
		return nil, err
	}
	if ttl.(int64) <= 0 {
		_, err = redis.ExecRedisCommand("DEL", token)
		return nil, errors.New("验证码已过期，请刷新重试")
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
	err = json.Unmarshal(base64Buff, &cachedPoint)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}
	//待校验数据
	base64Buff, err = base64.StdEncoding.DecodeString(pointJson)
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	err = json.Unmarshal(base64Buff, &checkedPoint)
	if err != nil {
		return nil, errors.New("json unmarshal error:" + err.Error())
	}

	success := false
	msg := "验证失败"
	if (math.Abs(float64(cachedPoint.X-checkedPoint.X)) <= this.threshold) &&
		(math.Abs(float64(cachedPoint.Y-checkedPoint.Y)) <= this.threshold) {
		success = true
		msg = "验证通过"
	}

	//验证后将缓存删除，同一个验证码只能用于验证一次
	_, err = redis.ExecRedisCommand("DEL", token)
	if err != nil {
		log.Printf("验证码缓存删除失败:%s", token)
	}
	return &RespMsg{Success: success, Message: msg}, nil
}

func (this *BlockPuzzle) interfereBlock(img *image.RGBA, point image.Point, srcBlockName string) {
	//干扰1
	blockPath := this.blockPath
	var blockName1 string
	for {
		blockName1, _ = randomFileName(blockPath)
		if blockName1 != srcBlockName {
			break
		}
	}
	blockFileName1 := blockPath + "/" + blockName1
	this.doInterfere(blockFileName1, img, point, 1)
	//干扰2
	var blockName2 string
	for {
		blockName2, _ = randomFileName(blockPath)
		if blockName2 != srcBlockName && blockName2 != blockName1 {
			break
		}
	}
	blockFileName2 := blockPath + "/" + blockName2
	this.doInterfere(blockFileName2, img, point, 2)
}

func (this *BlockPuzzle) doInterfere(blockFileName string, img *image.RGBA, point image.Point, _type int) {
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
			//在原扣图右边插入干扰图
			position = randInt(x+jigsawWidth+5, originalWidth-jigsawWidth)
		} else {
			//在原扣图左边插入干扰图
			position = randInt(100, x-jigsawWidth-5)
		}
	case 2:
		position = randInt(jigsawWidth, 100-jigsawWidth)
	}
	point = blockImg.Bounds().Min.Sub(image.Pt(-position, 0))
	//处理拼图块中模糊部分
	jigsaw := this.cropJigsaw(blockImg, img, point)
	blur := imaging.Blur(jigsaw, this.blur)
	blur = imaging.AdjustBrightness(blur, this.brightness)
	blurRGB := image2RGBA(blur)

	for x := 0; x < blockImg.Bounds().Dx(); x++ {
		for y := 0; y < blockImg.Bounds().Dy(); y++ {
			// 如果模板图像当前像素点不是透明色 copy源文件信息到目标图片中
			_, _, _, a := blockImg.At(x, y).RGBA()
			if a > TransparentThreshold {
				img.Set(x+point.X, y+point.Y, colorMix((blurRGB.At(x, y)).(color.RGBA), (img.At(x+point.X, y+point.Y)).(color.RGBA)))
			}
			//防止数组越界判断
			if x == (blockImg.Bounds().Dx()-1) || y == (blockImg.Bounds().Dy()-1) {
				continue
			}
			_, _, _, ra := blockImg.At(x+1, y).RGBA()
			_, _, _, da := blockImg.At(x, y+1).RGBA()
			//描边处理,取带像素和无像素的界点，判断该点是不是临界轮廓点,如果是设置该坐标像素是白色
			//为都抗锯齿，需要提高透明度阈值
			if (a > TransparentThreshold && ra <= TransparentThreshold) ||
				(a <= TransparentThreshold && ra > TransparentThreshold) ||
				(a > TransparentThreshold && da <= TransparentThreshold) ||
				(a <= TransparentThreshold && da > TransparentThreshold) {
				//如果模板当前像素点不透明，但右侧像素点透明
				//如果模板当前像素点透明，但右侧像素点不透明
				//如果模板当前像素点不透明，但下侧侧像素点透明
				//如果模板当前像素点透明，但下侧侧像素点不透明
				mix := colorMix(color.RGBA{R: 255, G: 255, B: 255, A: 220}, (img.At(point.X+x, point.Y+y)).(color.RGBA))
				img.Set(point.X+x, point.Y+y, mix)
			}
		}
	}
}

func (this *BlockPuzzle) cropJigsaw(blockImg, oriImg image.Image, point image.Point) *image.RGBA {
	newImage := image.NewRGBA(blockImg.Bounds())
	for x := 0; x < blockImg.Bounds().Dx(); x++ {
		for y := 0; y < blockImg.Bounds().Dy(); y++ {
			// 如果模板图像当前像素点不是透明色 copy源文件信息到目标图片中
			_, _, _, a := blockImg.At(x, y).RGBA()
			if a > TransparentThreshold {
				r, g, b, _ := oriImg.At(point.X+x, point.Y+y).RGBA()
				newImage.Set(x, y, colorTransparent(r, g, b, a))
			}
			//防止数组越界判断
			if x == (blockImg.Bounds().Dx()-1) || y == (blockImg.Bounds().Dy()-1) {
				continue
			}
		}
	}
	return newImage
}
