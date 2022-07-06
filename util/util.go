package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/golang/freetype"
)

func IsFile(path string) bool {
	fi, e := os.Stat(path)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}

func RandomFileName(dir string) (string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var fileNames []string //nolint:prealloc
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
	index := rand.Intn(len(fileNames)) //nolint:gosec
	if index >= len(fileNames) {
		index = len(fileNames) - 1
	}
	return fileNames[index], nil
}

func Image2RGBA(oriImg image.Image) *image.RGBA {
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

func DrawText(img draw.Image, watermarkText, fontFile string, watermarkSize int, dpi float64) error {
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
	c.SetDPI(dpi)
	c.SetFont(font)
	c.SetFontSize(float64(watermarkSize))
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(fontColor))
	point := freetype.Pt(pt.X, pt.Y)
	_, err = c.DrawString(watermarkText, point)
	return err
}

func Image2Base64(img image.Image, fileType string) (string, error) {
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

func RandInt(min, max int) int {
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return rand.Intn(max-min) + min //nolint:gosec
}

func SaveImage(fileName, fileType string, img image.Image) {
	var err error
	tmpFile, err := os.Create(fileName)
	if err != nil {
		log.Printf("create image file error: %v", err)
		return
	}
	defer tmpFile.Close()
	switch fileType {
	case "png":
		err = png.Encode(tmpFile, img)
	default:
		err = jpeg.Encode(tmpFile, img, &jpeg.Options{Quality: 100})
	}
	if err != nil {
		log.Printf("image encode error: %v", err)
	}
}

func Println(data interface{}) {
	j, _ := json.MarshalIndent(data, "", "\t")
	fmt.Println(string(j))
}
