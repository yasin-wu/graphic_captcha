package pkg

type Captchaer interface {
	Get(token string) (*Response, error)
	Check(token, pointJSON string) (*Response, error)
}
