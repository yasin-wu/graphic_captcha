package captcha

import (
	"errors"

	"github.com/yasin-wu/captcha/redis"
)

type CaptchaVO struct {
	Token               string   `json:"token"`                //每次验证请求唯一标识
	CaptchaType         string   `json:"captcha_type"`         //验证码类型:(click_word,block_puzzle)
	OriginalImageBase64 string   `json:"image_base_64"`        //原生图片base64
	JigsawImageBase64   string   `json:"jigsaw_image_base_64"` //滑块图片base64
	Words               []string `json:"words"`                //点选文字
}

type CaptchaConfig struct {
	ClickImagePath     string  `json:"click_image_path"`     //点选校验图片目录
	ClickWordFile      string  `json:"click_word_file"`      //点选文字文件
	FontFile           string  `json:"font_file"`            //字体文件
	ClickWordCount     int     `json:"click_word_count"`     //点选文字个数
	FontSize           int     `json:"font_size"`            //字体大小
	WatermarkText      string  `json:"watermark_text"`       //图片水印
	WatermarkSize      int     `json:"watermark_size"`       //水印大小
	DPI                float64 `json:"dpi"`                  //分辨率
	ExpireTime         int     `json:"expire_time"`          //校验过期时间
	JigsawOriginalPath string  `json:"jigsaw_original_path"` //滑块原图目录
	JigsawBlockPath    string  `json:"jigsaw_block_path"`    //滑块抠图目录
	JigsawThreshold    float64 `json:"jigsaw_threshold"`     //滑块容忍的偏差范围
	JigsawBlur         float64 `json:"jigsaw_blur"`          //滑块空缺的模糊度
	JigsawBrightness   float64 `json:"jigsaw_brightness"`    //滑块空缺亮度
}

type RespMsg struct {
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
}

type Captcha interface {
	Get(token string) (*CaptchaVO, error)
	Check(token, pointJson string) (*RespMsg, error)
}

var (
	captchaConf *CaptchaConfig
)

func New(captchaType CaptchaType, conf *CaptchaConfig, redisConf *redis.RedisConfig) (Captcha, error) {
	if conf == nil || redisConf == nil {
		return nil, errors.New("conf is nil")
	}
	newCaptchaConf(conf)
	redis.InitRedisPool(redisConf)
	switch captchaType {
	case CaptchaTypeClickWord:
		return &ClickWord{}, nil
	case CaptchaTypeBlockPuzzle:
		return &BlockPuzzle{}, nil
	}
	return nil, errors.New("验证类型错误")
}

func newCaptchaConf(conf *CaptchaConfig) {
	if conf.ClickImagePath == "" {
		conf.ClickImagePath = "../pic_click"
	}
	if conf.ClickWordFile == "" {
		conf.ClickWordFile = "../fonts/license.txt"
	}
	if conf.FontFile == "" {
		conf.FontFile = "../fonts/captcha.ttf"
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
		conf.ExpireTime = 30
	}
	if conf.JigsawOriginalPath == "" {
		conf.JigsawOriginalPath = "../jigsaw/original"
	}
	if conf.JigsawBlockPath == "" {
		conf.JigsawBlockPath = "../jigsaw/sliding_block"
	}
	if conf.JigsawThreshold == 0 {
		conf.JigsawThreshold = 8
	}
	if conf.JigsawBlur == 0 {
		conf.JigsawBlur = 1.3
	}
	if conf.JigsawBrightness == 0 {
		conf.JigsawBrightness = -30
	}
	captchaConf = conf
}
