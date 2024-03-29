package image

import (
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"
	"strings"

	"github.com/yasin-wu/graphic_captcha/v2/internal/util"
)

type Image struct {
	FileType string
	FileName string
	Image    image.Image
}

func New(dir string) (*Image, error) {
	var err error
	var fileName string
	var staticImg image.Image
	filePath := dir
	if !util.IsFile(dir) {
		fileName, err = util.RandomFileName(dir)
		if err != nil {
			log.Printf("random file name error: %v", err)
			return nil, err
		}
		filePath = dir + "/" + fileName
	}
	imgFile, err := os.Open(filePath)
	if err != nil {
		log.Printf("open file error: %v", err)
		return nil, err
	}
	defer imgFile.Close()
	fileType := strings.Replace(path.Ext(path.Base(imgFile.Name())), ".", "", -1)
	switch fileType {
	case "png":
		staticImg, err = png.Decode(imgFile)
	default:
		staticImg, err = jpeg.Decode(imgFile)
	}
	if err != nil {
		log.Printf("image decode error: %v", err)
		return nil, err
	}
	return &Image{
		FileType: fileType,
		FileName: fileName,
		Image:    staticImg,
	}, nil
}
