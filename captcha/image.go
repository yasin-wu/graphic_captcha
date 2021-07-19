package captcha

import (
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"
	"strings"
)

type Image struct {
	FileType string
	FileName string
	Image    image.Image
}

func NewImage(dir string) (*Image, error) {
	var err error
	var fileName string
	var staticImg image.Image
	filePath := dir
	if !isFile(dir) {
		fileName, err = randomFileName(dir)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		filePath = dir + "/" + fileName
	}
	imgFile, err := os.Open(filePath)
	if err != nil {
		log.Print(err)
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
		log.Print(err)
		return nil, err
	}
	return &Image{
		FileType: fileType,
		FileName: fileName,
		Image:    staticImg,
	}, nil
}
