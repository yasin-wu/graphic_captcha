package config

import (
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisOptions redis.Options

type Option func(conf *Config)

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

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: path string
 * @return: Option
 * @description: 配置文字点选背景图路径
 */
func WithClickImagePath(path string) Option {
	return func(config *Config) {
		config.ClickImagePath = path
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: file string
 * @return: Option
 * @description: 配置文字点选文字文件路径
 */
func WithClickWordFile(file string) Option {
	return func(config *Config) {
		config.ClickWordFile = file
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: count int
 * @return: Option
 * @description: 配置文字点选个数
 */
func WithClickWordCount(count int) Option {
	return func(config *Config) {
		config.ClickWordCount = count
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: path string
 * @return: Option
 * @description: 配置滑块验证背景图路径
 */
func WithOriginalPath(path string) Option {
	return func(config *Config) {
		config.OriginalPath = path
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: path string
 * @return: Option
 * @description: 配置滑块验证切块图路径
 */
func WithBlockPath(path string) Option {
	return func(config *Config) {
		config.BlockPath = path
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: Threshold float64
 * @return: Option
 * @description: 配置滑块验证边界值
 */
func WithThreshold(threshold float64) Option {
	return func(config *Config) {
		config.Threshold = threshold
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: Blur float64
 * @return: Option
 * @description: 配置滑块验证模糊值
 */
func WithBlur(blur float64) Option {
	return func(config *Config) {
		config.Blur = blur
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: Brightness float64
 * @return: Option
 * @description: 配置滑块验证亮度值
 */
func WithBrightness(brightness float64) Option {
	return func(config *Config) {
		config.Brightness = brightness
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: file string
 * @return: Option
 * @description: 配置字体文件路径
 */
func WithFontFile(file string) Option {
	return func(config *Config) {
		config.FontFile = file
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: size int
 * @return: Option
 * @description: 配置字体大小
 */
func WithFontSize(size int) Option {
	return func(config *Config) {
		config.FontSize = size
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: text string
 * @return: Option
 * @description: 配置水印内容
 */
func WithWatermarkText(text string) Option {
	return func(config *Config) {
		config.WatermarkText = text
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: size int
 * @return: Option
 * @description: 配置水印大小
 */
func WithWatermarkSize(size int) Option {
	return func(config *Config) {
		config.WatermarkSize = size
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:08
 * @params: DPI float64
 * @return: Option
 * @description: 配置DPI
 */
func WithDPI(dpi float64) Option {
	return func(config *Config) {
		config.DPI = dpi
	}
}

/**
 * @author: yasinWu
 * @date: 2022/1/13 14:14
 * @params: ExpireTime time.Duration
 * @return: Option
 * @description: 配置Redis key过期时间
 */
func WithExpireTime(expireTime time.Duration) Option {
	return func(config *Config) {
		config.ExpireTime = expireTime
	}
}

func CheckConf(conf *Config) {
	if conf.ClickImagePath == "" {
		conf.ClickImagePath = "../config/pic_click"
	}
	if conf.ClickWordFile == "" {
		conf.ClickWordFile = "../config/fonts/license.txt"
	}
	if conf.FontFile == "" {
		conf.FontFile = "../config/fonts/captcha.ttf"
	}
	if conf.ClickWordCount == 0 {
		conf.ClickWordCount = 4
	}
	if conf.FontSize == 0 {
		conf.FontSize = 26
	}
	if conf.WatermarkText == "" {
		conf.WatermarkText = "yasin"
	}
	if conf.WatermarkSize == 0 {
		conf.WatermarkSize = 14
	}
	if conf.DPI == 0 {
		conf.DPI = 72
	}
	if conf.ExpireTime == 0 {
		conf.ExpireTime = time.Minute
	}
	if conf.OriginalPath == "" {
		conf.OriginalPath = "../config/jigsaw/original"
	}
	if conf.BlockPath == "" {
		conf.BlockPath = "../config/jigsaw/sliding_block"
	}
	if conf.Threshold == 0 {
		conf.Threshold = 8
	}
	if conf.Blur == 0 {
		conf.Blur = 1.3
	}
	if conf.Brightness == 0 {
		conf.Brightness = -30
	}
}
