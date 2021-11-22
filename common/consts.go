package common

type CaptchaType string

const (
	CaptchaTypeClickWord   CaptchaType = "click_word"   //点选文字
	CaptchaTypeBlockPuzzle CaptchaType = "block_puzzle" //滑块图片
)

const (
	TokenFormat = "^CAPT:%s;CLI:%s;STAMP:%d#"
)

const (
	TransparentThreshold uint32 = 150 << 8
)
