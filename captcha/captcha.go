package captcha

import (
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/davecgh/go-spew/spew"

	mredis "github.com/gomodule/redigo/redis"
	"github.com/yasin-wu/captcha/redis"
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

type CaptchaConfig struct {
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

func New(captchaType CaptchaType, conf *CaptchaConfig, redisConf *redis.RedisConfig) (Captcha, error) {
	if conf == nil || redisConf == nil {
		return nil, errors.New("conf is nil")
	}
	checkCaptchaConf(conf)
	redis.InitRedisPool(redisConf)
	switch captchaType {
	case CaptchaTypeClickWord:
		return &ClickWord{
			imagePath:     conf.ClickImagePath,
			wordFile:      conf.ClickWordFile,
			wordCount:     conf.ClickWordCount,
			fontFile:      conf.FontFile,
			fontSize:      conf.FontSize,
			watermarkText: conf.WatermarkText,
			watermarkSize: conf.WatermarkSize,
			dpi:           conf.DPI,
			expireTime:    conf.ExpireTime,
		}, nil
	case CaptchaTypeBlockPuzzle:
		return &BlockPuzzle{
			originalPath:  conf.JigsawOriginalPath,
			blockPath:     conf.JigsawBlockPath,
			threshold:     conf.JigsawThreshold,
			blur:          conf.JigsawBlur,
			brightness:    conf.JigsawBrightness,
			fontFile:      conf.FontFile,
			watermarkText: conf.WatermarkText,
			watermarkSize: conf.WatermarkSize,
			dpi:           conf.DPI,
			expireTime:    conf.ExpireTime,
		}, nil
	}
	return nil, errors.New("验证类型错误")
}

func checkCaptchaConf(conf *CaptchaConfig) {
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
		conf.ExpireTime = 60
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
}

//校验数据存入Redis,存入时进行base64
func setRedis(token, data interface{}, expireTime int) error {
	dataBuff, err := json.Marshal(data)
	if err != nil {
		return errors.New("json marshal error:" + err.Error())
	}
	data64 := base64.StdEncoding.EncodeToString(dataBuff)
	spew.Dump("数据:" + data64)
	_, err = redis.ExecRedisCommand("SET", token, data64, "EX", expireTime)
	if err != nil {
		return errors.New("存储至redis失败")
	}
	return nil
}

//从Redis获取待校验数据,并解base64
func getRedis(token string) ([]byte, error) {
	ttl, err := redis.ExecRedisCommand("TTL", token)
	if err != nil {
		return nil, err
	}
	if ttl.(int64) <= 0 {
		_, err = redis.ExecRedisCommand("DEL", token)
		return nil, errors.New("验证码已过期，请刷新重试")
	}
	cachedBuff, err := mredis.Bytes(redis.ExecRedisCommand("GET", token))
	if err != nil {
		return nil, errors.New("get captcha error:" + err.Error())
	}
	base64Buff, err := base64.StdEncoding.DecodeString(string(cachedBuff))
	if err != nil {
		return nil, errors.New("base64 decode error:" + err.Error())
	}
	return base64Buff, nil
}
