package captcha

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	mredis "github.com/gomodule/redigo/redis"

	"github.com/davecgh/go-spew/spew"

	"github.com/yasin-wu/captcha/redis"
	yredis "github.com/yasin-wu/utils/redis"
)

type CaptchaVO struct {
	Token               string   `json:"token"`                  //每次验证请求唯一标识
	CaptchaType         string   `json:"captcha_type"`           //验证码类型:(click_word,block_puzzle)
	OriginalImageBase64 string   `json:"original_image_base_64"` //原生图片base64
	JigsawImageBase64   string   `json:"jigsaw_image_base_64"`   //滑块图片base64
	Words               []string `json:"words"`                  //点选文字
}

type RespMsg struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Config struct {
	ClickImagePath     string  //点选校验图片目录
	ClickWordFile      string  //点选文字文件
	ClickWordCount     int     //点选文字个数
	JigsawOriginalPath string  //滑块原图目录
	JigsawBlockPath    string  //滑块抠图目录
	JigsawThreshold    float64 //滑块容忍的偏差范围
	JigsawBlur         float64 //滑块空缺的模糊度
	JigsawBrightness   float64 //滑块空缺亮度
	FontFile           string  //字体文件
	FontSize           int     //字体大小
	WatermarkText      string  //图片水印
	WatermarkSize      int     //水印大小
	DPI                float64 //分辨率
	ExpireTime         int     //校验过期时间
}

type Captcha interface {
	Get(token string) (*CaptchaVO, error)
	Check(token, pointJson string) (*RespMsg, error)
}

func New(captchaType CaptchaType, captchaConf *Config, redisConf *yredis.Config) (Captcha, error) {
	if captchaConf == nil || redisConf == nil {
		return nil, errors.New("conf is nil")
	}
	checkCaptchaConf(captchaConf)
	redis.InitRedis(redisConf)
	switch captchaType {
	case CaptchaTypeClickWord:
		return &ClickWord{
			imagePath:     captchaConf.ClickImagePath,
			wordFile:      captchaConf.ClickWordFile,
			wordCount:     captchaConf.ClickWordCount,
			fontFile:      captchaConf.FontFile,
			fontSize:      captchaConf.FontSize,
			watermarkText: captchaConf.WatermarkText,
			watermarkSize: captchaConf.WatermarkSize,
			dpi:           captchaConf.DPI,
			expireTime:    captchaConf.ExpireTime,
		}, nil
	case CaptchaTypeBlockPuzzle:
		return &BlockPuzzle{
			originalPath:  captchaConf.JigsawOriginalPath,
			blockPath:     captchaConf.JigsawBlockPath,
			threshold:     captchaConf.JigsawThreshold,
			blur:          captchaConf.JigsawBlur,
			brightness:    captchaConf.JigsawBrightness,
			fontFile:      captchaConf.FontFile,
			watermarkText: captchaConf.WatermarkText,
			watermarkSize: captchaConf.WatermarkSize,
			dpi:           captchaConf.DPI,
			expireTime:    captchaConf.ExpireTime,
		}, nil
	}
	return nil, errors.New("验证类型错误")
}

func checkCaptchaConf(config *Config) {
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
		config.ExpireTime = 60
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

//校验数据存入Redis,存入时进行base64
func setRedis(token, data interface{}, expireTime int) error {
	dataBuff, err := json.Marshal(data)
	if err != nil {
		return errors.New("json marshal error:" + err.Error())
	}
	data64 := base64.StdEncoding.EncodeToString(dataBuff)
	spew.Dump("数据:" + data64)
	_, err = redis.RedisClient.Exec("SET", token, data64, "EX", expireTime)
	if err != nil {
		return errors.New("存储至redis失败")
	}
	return nil
}

//从Redis获取待校验数据,并解base64
func getRedis(token string) ([]byte, error) {
	ttl, err := redis.RedisClient.Exec("TTL", token)
	if err != nil {
		return nil, err
	}
	if ttl.(int64) <= 0 {
		_, err = redis.RedisClient.Exec("DEL", token)
		return nil, errors.New("验证码已过期，请刷新重试")
	}
	cachedBuff, err := mredis.Bytes(redis.RedisClient.Exec("GET", token))
	if err != nil {
		return nil, errors.New("get captcha error:" + err.Error())
	}
	base64Buff, err := base64.StdEncoding.DecodeString(string(cachedBuff))
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	return base64Buff, nil
}
