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
	//LoadConfig()
}

type Config struct {
	ClickImagePath     string        //点选校验图片目录
	ClickWordFile      string        //点选文字文件
	ClickWordCount     int           //点选文字个数
	JigsawOriginalPath string        //滑块原图目录
	JigsawBlockPath    string        //滑块抠图目录
	JigsawThreshold    float64       //滑块容忍的偏差范围
	JigsawBlur         float64       //滑块空缺的模糊度
	JigsawBrightness   float64       //滑块空缺亮度
	FontFile           string        //字体文件
	FontSize           int           //字体大小
	WatermarkText      string        //图片水印
	WatermarkSize      int           //水印大小
	DPI                float64       //分辨率
	ExpireTime         time.Duration //校验过期时间
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
			originalPath:  config.JigsawOriginalPath,
			blockPath:     config.JigsawBlockPath,
			threshold:     config.JigsawThreshold,
			blur:          config.JigsawBlur,
			brightness:    config.JigsawBrightness,
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
	if config.JigsawOriginalPath == "" {
		config.JigsawOriginalPath = "../conf/jigsaw/original"
	}
	if config.JigsawBlockPath == "" {
		config.JigsawBlockPath = "../conf/jigsaw/sliding_block"
	}
	if config.JigsawThreshold == 0 {
		config.JigsawThreshold = 8
	}
	if config.JigsawBlur == 0 {
		config.JigsawBlur = 1.3
	}
	if config.JigsawBrightness == 0 {
		config.JigsawBrightness = -30
	}
}
