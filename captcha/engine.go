package captcha

import (
	"errors"
	"time"

	"github.com/yasin-wu/captcha/redis"

	"github.com/yasin-wu/captcha/common"
	yredis "github.com/yasin-wu/utils/redis"
)

type Engine interface {
	Get(token string) (*common.Captcha, error)
	Check(token, pointJson string) (*common.RespMsg, error)
}

type Config struct {
	ClickImagePath string
	ClickWordFile  string
	ClickWordCount int
	OriginalPath   string
	BlockPath      string
	Threshold      float64
	Blur           float64
	Brightness     float64
	FontFile       string
	FontSize       int
	WatermarkText  string
	WatermarkSize  int
	DPI            float64
	ExpireTime     time.Duration
}

func New(captchaType common.CaptchaType, config *Config, host string, options ...yredis.Option) (Engine, error) {
	if config == nil {
		return nil, errors.New("conf is nil")
	}
	checkConf(config)
	redisCli := redis.New(host, options...)
	if redisCli == nil {
		return nil, errors.New("redis client is nil")
	}
	switch captchaType {
	case common.CaptchaTypeClickWord:
		return &ClickWord{
			imagePath:     config.ClickImagePath,
			wordFile:      config.ClickWordFile,
			wordCount:     config.ClickWordCount,
			fontFile:      config.FontFile,
			fontSize:      config.FontSize,
			watermarkText: config.WatermarkText,
			watermarkSize: config.WatermarkSize,
			dpi:           config.DPI,
			expireTime:    config.ExpireTime,
			redisCli:      redisCli,
		}, nil
	case common.CaptchaTypeBlockPuzzle:
		return &SlideBlock{
			originalPath:  config.OriginalPath,
			blockPath:     config.BlockPath,
			threshold:     config.Threshold,
			blur:          config.Blur,
			brightness:    config.Brightness,
			fontFile:      config.FontFile,
			fontSize:      config.FontSize,
			watermarkText: config.WatermarkText,
			watermarkSize: config.WatermarkSize,
			dpi:           config.DPI,
			expireTime:    config.ExpireTime,
			redisCli:      redisCli,
		}, nil
	}
	return nil, errors.New("验证类型错误")
}

func checkConf(config *Config) {
	if config.ClickImagePath == "" {
		config.ClickImagePath = "../conf/pic_click"
	}
	if config.ClickWordFile == "" {
		config.ClickWordFile = "../conf/fonts/license.txt"
	}
	if config.FontFile == "" {
		config.FontFile = "../conf/fonts/captcha.ttf"
	}
	if config.ClickWordCount == 0 {
		config.ClickWordCount = 4
	}
	if config.FontSize == 0 {
		config.FontSize = 26
	}
	if config.WatermarkText == "" {
		config.WatermarkText = "yasin"
	}
	if config.WatermarkSize == 0 {
		config.WatermarkSize = 14
	}
	if config.DPI == 0 {
		config.DPI = 72
	}
	if config.ExpireTime == 0 {
		config.ExpireTime = time.Minute
	}
	if config.OriginalPath == "" {
		config.OriginalPath = "../conf/jigsaw/original"
	}
	if config.BlockPath == "" {
		config.BlockPath = "../conf/jigsaw/sliding_block"
	}
	if config.Threshold == 0 {
		config.Threshold = 8
	}
	if config.Blur == 0 {
		config.Blur = 1.3
	}
	if config.Brightness == 0 {
		config.Brightness = -30
	}
}
