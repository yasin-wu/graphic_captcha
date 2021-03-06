package consts

type CaptchaType string

const (
	CaptchaTypeClickWord   CaptchaType = "click_word"   // 点选验证
	CaptchaTypeBlockPuzzle CaptchaType = "block_puzzle" // 滑块验证
)

const (
	TokenFormat = "^CAPT:%s;CLI:%s;STAMP:%d#" //nolint:gosec
)
