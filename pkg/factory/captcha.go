package factory

import (
	"github.com/yasin-wu/graphic_captcha/v2/pkg/entity"
)

type Captchaer interface {
	Get(token string) (*entity.Response, error)
	Check(token, pointJSON string) (*entity.Response, error)
}
