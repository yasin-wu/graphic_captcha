package captcha

type CaptchaType string

const (
	CaptchaTypeClickWord   CaptchaType = "click_word"
	CaptchaTypeBlockPuzzle CaptchaType = "block_puzzle"
)

const (
	TOKENFORMAT = "^CAPT:%s;CLI:%s;STAMP:%d#"
)

const (
	TransparentThreshold uint32 = 150 << 8
)
