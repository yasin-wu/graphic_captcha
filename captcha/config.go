package captcha

import (
	"time"

	"github.com/yasin-wu/utils/redis"
)

type Option func(conf *config)

type config struct {
	clickImagePath string
	clickWordFile  string
	clickWordCount int
	originalPath   string
	blockPath      string
	threshold      float64
	blur           float64
	brightness     float64
	fontFile       string
	fontSize       int
	watermarkText  string
	watermarkSize  int
	dpi            float64
	redisOptions   []redis.Option
	expireTime     time.Duration
}

func WithClickImagePath(path string) Option {
	return func(config *config) {
		config.clickImagePath = path
	}
}

func WithClickWordFile(file string) Option {
	return func(config *config) {
		config.clickWordFile = file
	}
}

func WithClickWordCount(count int) Option {
	return func(config *config) {
		config.clickWordCount = count
	}
}

func WithOriginalPath(path string) Option {
	return func(config *config) {
		config.originalPath = path
	}
}

func WithBlockPath(path string) Option {
	return func(config *config) {
		config.blockPath = path
	}
}

func WithThreshold(threshold float64) Option {
	return func(config *config) {
		config.threshold = threshold
	}
}

func WithBlur(blur float64) Option {
	return func(config *config) {
		config.blur = blur
	}
}

func WithBrightness(brightness float64) Option {
	return func(config *config) {
		config.brightness = brightness
	}
}

func WithFontFile(file string) Option {
	return func(config *config) {
		config.fontFile = file
	}
}

func WithFontSize(size int) Option {
	return func(config *config) {
		config.fontSize = size
	}
}

func WithWatermarkText(text string) Option {
	return func(config *config) {
		config.watermarkText = text
	}
}

func WithWatermarkSize(size int) Option {
	return func(config *config) {
		config.watermarkSize = size
	}
}

func WithDPI(dpi float64) Option {
	return func(config *config) {
		config.dpi = dpi
	}
}

func WithRedisOptions(options ...redis.Option) Option {
	return func(config *config) {
		config.redisOptions = options
	}
}

func WithExpireTime(expireTime time.Duration) Option {
	return func(config *config) {
		config.expireTime = expireTime
	}
}

func checkConf(conf *config) {
	if conf.clickImagePath == "" {
		conf.clickImagePath = "../conf/pic_click"
	}
	if conf.clickWordFile == "" {
		conf.clickWordFile = "../conf/fonts/license.txt"
	}
	if conf.fontFile == "" {
		conf.fontFile = "../conf/fonts/captcha.ttf"
	}
	if conf.clickWordCount == 0 {
		conf.clickWordCount = 4
	}
	if conf.fontSize == 0 {
		conf.fontSize = 26
	}
	if conf.watermarkText == "" {
		conf.watermarkText = "yasin"
	}
	if conf.watermarkSize == 0 {
		conf.watermarkSize = 14
	}
	if conf.dpi == 0 {
		conf.dpi = 72
	}
	if conf.expireTime == 0 {
		conf.expireTime = time.Minute
	}
	if conf.originalPath == "" {
		conf.originalPath = "../conf/jigsaw/original"
	}
	if conf.blockPath == "" {
		conf.blockPath = "../conf/jigsaw/sliding_block"
	}
	if conf.threshold == 0 {
		conf.threshold = 8
	}
	if conf.blur == 0 {
		conf.blur = 1.3
	}
	if conf.brightness == 0 {
		conf.brightness = -30
	}
}
