package service

import (
	"errors"
	"time"

	"github.com/mojocn/base64Captcha"
)

var ErrCaptchaInvalid = errors.New("captcha invalid or expired")

type CaptchaService struct {
	captcha *base64Captcha.Captcha
}

func NewCaptchaService() *CaptchaService {
	store := base64Captcha.NewMemoryStore(10240, 5*time.Minute)
	driver := base64Captcha.NewDriverString(
		80,  // height
		240, // width
		10,  // noise count
		base64Captcha.OptionShowHollowLine|base64Captcha.OptionShowSlimeLine,
		5, // length
		"ABCDEFGHJKMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789", // 排除易混淆字符 0/O/1/I/l
		nil, // bgColor（随机浅色）
		base64Captcha.DefaultEmbeddedFonts,
		nil, // 使用所有内置字体
	)
	return &CaptchaService{
		captcha: base64Captcha.NewCaptcha(driver, store),
	}
}

// Generate 生成验证码，返回 id 和 base64 图片字符串
func (s *CaptchaService) Generate() (id, b64Image string, err error) {
	id, b64Image, _, err = s.captcha.Generate()
	return
}

// Verify 校验验证码（不区分大小写），校验后立即失效
func (s *CaptchaService) Verify(id, answer string) error {
	if !s.captcha.Verify(id, answer, true) {
		return ErrCaptchaInvalid
	}
	return nil
}
