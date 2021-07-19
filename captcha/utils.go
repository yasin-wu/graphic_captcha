package captcha

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/yasin-wu/captcha/redis"
)

func randomFileName(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var fileNames []string
	for _, v := range files {
		if strings.HasPrefix(v.Name(), ".") {
			continue
		}
		fileNames = append(fileNames, v.Name())
	}
	if len(fileNames) == 0 {
		return "", errors.New("dir is nil")
	}
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(fileNames))
	if index >= len(fileNames) {
		index = len(fileNames) - 1
	}
	return fileNames[index], nil
}

func image2RGBA(oriImg image.Image) *image.RGBA {
	if oriImg == nil {
		return nil
	}
	if rgba, ok := oriImg.(*image.RGBA); ok {
		return rgba
	}
	rgba := image.NewRGBA(oriImg.Bounds())
	draw.Draw(rgba, rgba.Bounds(), oriImg, oriImg.Bounds().Min, draw.Src)
	return rgba
}

func drawText(img *image.RGBA, watermarkText, fontFile string, watermarkSize int, dpi float64) error {
	watermarkLen := strings.Count(watermarkText, "") - 1
	pt := image.Pt(img.Bounds().Dx()-(watermarkSize*watermarkLen), img.Bounds().Dy()-watermarkLen)
	fontColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	fontBytes, err := ioutil.ReadFile(fontFile)
	if err != nil {
		return errors.New("read font file error:" + err.Error())
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return errors.New("parse font error:" + err.Error())
	}
	c := freetype.NewContext()
	//设置分辨率
	c.SetDPI(dpi)
	//设置字体
	c.SetFont(font)
	//设置尺寸
	c.SetFontSize(float64(watermarkSize))
	c.SetClip(img.Bounds())
	//设置输出的图片
	c.SetDst(img)
	c.SetSrc(image.NewUniform(fontColor))
	point := freetype.Pt(pt.X, pt.Y)
	_, err = c.DrawString(watermarkText, point)
	return err
}

func StringsContains(stringArray []string, substr string) bool {
	for _, v := range stringArray {
		if v == substr {
			return true
		}
	}
	return false
}

func drawTextOnBackground(bg *image.RGBA, pt image.Point, fontStyle *truetype.Font, text string, fontColor color.Color, fontSize int, angle float64) {
	fontPng := drawString2Png(fontStyle, fontColor, text, float64(fontSize))
	rotate := imaging.Rotate(fontPng, angle, color.Transparent)
	resize := imaging.Resize(rotate, fontSize, fontSize, imaging.Lanczos)
	resizePng := image2RGBA(resize)
	draw.Draw(bg, image.Rect(pt.X, pt.Y, pt.X+fontSize, pt.Y+fontSize), resizePng, image.ZP, draw.Over)
}

func drawString2Png(font *truetype.Font, c color.Color, str string, fontSize float64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, int(fontSize), int(fontSize)))
	ctx := freetype.NewContext()
	// default 72dpi
	ctx.SetDst(img)
	ctx.SetClip(img.Bounds())
	ctx.SetSrc(image.NewUniform(c))
	ctx.SetFontSize(fontSize)
	ctx.SetFont(font)
	// 写入文字的位置
	pt := freetype.Pt(0, int(-fontSize/6)+ctx.PointToFixed(fontSize).Ceil())
	ctx.DrawString(str, pt)
	return img
}

func imgToBase64(img image.Image, fileType string) (string, error) {
	var err error
	emptyBuff := bytes.NewBuffer(nil)
	switch fileType {
	case "png":
		err = png.Encode(emptyBuff, img)
	default:
		err = jpeg.Encode(emptyBuff, img, nil)
	}
	if err != nil {
		return "", err
	}
	dist := make([]byte, 20*1024*1024)
	base64.StdEncoding.Encode(dist, emptyBuff.Bytes())
	index := bytes.IndexByte(dist, 0)
	baseImage := dist[0:index]
	return *(*string)(unsafe.Pointer(&baseImage)), nil
}

func SetRedis(token, data interface{}, expireTime int) error {
	_, err := redis.ExecRedisCommand("SET", token, data, "EX", expireTime)
	if err != nil {
		return errors.New("存储至redis失败")
	}
	return nil
}

func generateJigsawPoint(originalWidth, originalHeight, jigsawWidth, jigsawHeight int) image.Point {
	rand.Seed(time.Now().UnixNano())
	widthDifference := originalWidth - jigsawWidth
	heightDifference := originalHeight - jigsawHeight
	x, y := 0, 0
	if widthDifference <= 0 {
		x = 5
	} else {
		x = rand.Intn(originalWidth-jigsawWidth-100) + 100
	}
	if heightDifference <= 0 {
		y = 5
	} else {
		y = rand.Intn(originalWidth-jigsawWidth) + 5
	}
	return image.Point{X: x, Y: y}
}

func randInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min) + min
}

func colorTransparent(r, g, b, a uint32) color.RGBA {
	rgba := image.NewRGBA(image.Rect(0, 0, 1, 1))
	convert := rgba.ColorModel().Convert(color.NRGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: uint16(a)})
	rr, gg, bb, aa := convert.RGBA()
	return color.RGBA{R: uint8(rr), G: uint8(gg), B: uint8(bb), A: uint8(aa)}
}

//色彩混合
func colorMix(fg, bg color.RGBA) color.RGBA {
	rgba := color.RGBA{}
	fa := FloatRound(float64(fg.A)/255, 2)
	ba := FloatRound(float64(bg.A)/255, 2)
	alpha := 1 - (1-fa)*(1-ba)
	rgba.R = uint8((float64(fg.R)*fa + float64(bg.R)*ba*(1-fa)) / alpha)
	rgba.G = uint8((float64(fg.G)*fa + float64(bg.G)*ba*(1-fa)) / alpha)
	rgba.B = uint8((float64(fg.B)*fa + float64(bg.B)*ba*(1-fa)) / alpha)
	rgba.A = uint8(alpha * 255)
	return rgba
}

func FloatRound(f float64, n int) float64 {
	format := "%." + strconv.Itoa(n) + "f"
	res, _ := strconv.ParseFloat(fmt.Sprintf(format, f), 64)
	return res
}

func saveImage(filePath string, img image.Image, opt int) bool {
	var file io.Writer
	if !strings.Contains(filePath, ".") {
		return false
	}
	_, e := os.Stat(filePath)
	if e != nil {
		f, err := os.Create(filePath)
		defer f.Close()
		file = f
		if err != nil {
			return false
		}
	} else {
		e := os.Remove(filePath)
		if e != nil {
			fmt.Println(e)
		}
		f, err := os.Create(filePath)
		defer f.Close()
		file = f
		if err != nil {
			return false
		}
	}
	if strings.HasSuffix(filePath, "jpg") || strings.HasSuffix(filePath, "jpeg") {
		e := jpeg.Encode(file, img, &jpeg.Options{opt})
		if e != nil {
			return false
		}
	} else if strings.HasSuffix(filePath, "png") {
		e := png.Encode(file, img)
		if e != nil {
			return false
		}
	} else if strings.HasSuffix(filePath, "gif") {
		e := gif.Encode(file, img, nil)
		if e != nil {
			return false
		}
	} else {
		fmt.Errorf("不支持的图片格式")
	}
	return true
}

func isFile(path string) bool {
	fi, e := os.Stat(path)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}
