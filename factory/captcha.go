package factory

import "github.com/yasin-wu/graphic_captcha/v2/entity"

type Captcha interface {
	Get(token string) (*entity.Response, error)
	Check(token, pointJSON string) (*entity.Response, error)
}
