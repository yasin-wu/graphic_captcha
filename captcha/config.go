package captcha

import (
	"time"

	redis2 "github.com/yasin-wu/graphic_captcha/redis"
	"github.com/yasin-wu/utils/redis"
)

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @description: 验证器配置项选择器
 */
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
	redisCli       *redis2.Client
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: path string
 * @return: Option
 * @description: 配置文字点选背景图路径
 */
func WithClickImagePath(path string) Option {
	return func(config *config) {
		config.clickImagePath = path
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: file string
 * @return: Option
 * @description: 配置文字点选文字文件路径
 */
func WithClickWordFile(file string) Option {
	return func(config *config) {
		config.clickWordFile = file
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: count int
 * @return: Option
 * @description: 配置文字点选个数
 */
func WithClickWordCount(count int) Option {
	return func(config *config) {
		config.clickWordCount = count
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: path string
 * @return: Option
 * @description: 配置滑块验证背景图路径
 */
func WithOriginalPath(path string) Option {
	return func(config *config) {
		config.originalPath = path
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: path string
 * @return: Option
 * @description: 配置滑块验证切块图路径
 */
func WithBlockPath(path string) Option {
	return func(config *config) {
		config.blockPath = path
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: threshold float64
 * @return: Option
 * @description: 配置滑块验证边界值
 */
func WithThreshold(threshold float64) Option {
	return func(config *config) {
		config.threshold = threshold
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: blur float64
 * @return: Option
 * @description: 配置滑块验证模糊值
 */
func WithBlur(blur float64) Option {
	return func(config *config) {
		config.blur = blur
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: brightness float64
 * @return: Option
 * @description: 配置滑块验证亮度值
 */
func WithBrightness(brightness float64) Option {
	return func(config *config) {
		config.brightness = brightness
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: file string
 * @return: Option
 * @description: 配置字体文件路径
 */
func WithFontFile(file string) Option {
	return func(config *config) {
		config.fontFile = file
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: size int
 * @return: Option
 * @description: 配置字体大小
 */
func WithFontSize(size int) Option {
	return func(config *config) {
		config.fontSize = size
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: text string
 * @return: Option
 * @description: 配置水印内容
 */
func WithWatermarkText(text string) Option {
	return func(config *config) {
		config.watermarkText = text
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: size int
 * @return: Option
 * @description: 配置水印大小
 */
func WithWatermarkSize(size int) Option {
	return func(config *config) {
		config.watermarkSize = size
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: dpi float64
 * @return: Option
 * @description: 配置DPI
 */
func WithDPI(dpi float64) Option {
	return func(config *config) {
		config.dpi = dpi
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:08
 * @params: options ...redis.Option
 * @return: Option
 * @description: 配置Redis
 */
func WithRedisOptions(options ...redis.Option) Option {
	return func(config *config) {
		config.redisOptions = options
	}
}

/**
 * @author: yasin
 * @date: 2022/1/13 14:14
 * @params: expireTime time.Duration
 * @return: Option
 * @description: 配置Redis key过期时间
 */
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
