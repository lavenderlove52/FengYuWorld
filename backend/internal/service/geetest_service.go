package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const geetestValidateURL = "https://gcaptcha4.geetest.com/validate"

type GeetestService struct {
	captchaID  string
	captchaKey string
	client     *http.Client
}

func NewGeetestService(captchaID, captchaKey string) *GeetestService {
	return &GeetestService{
		captchaID:  captchaID,
		captchaKey: captchaKey,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *GeetestService) signToken(lotNumber string) string {
	mac := hmac.New(sha256.New, []byte(s.captchaKey))
	mac.Write([]byte(lotNumber))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify 返回 nil 表示验证通过，否则返回错误
// 若 captchaID 或 captchaKey 为空，则跳过验证（开发环境）
func (s *GeetestService) Verify(lotNumber, captchaOutput, passToken, genTime string) error {
	if s.captchaID == "" || s.captchaKey == "" {
		return nil
	}

	signToken := s.signToken(lotNumber)

	formData := url.Values{}
	formData.Set("lot_number", lotNumber)
	formData.Set("captcha_output", captchaOutput)
	formData.Set("pass_token", passToken)
	formData.Set("gen_time", genTime)
	formData.Set("captcha_id", s.captchaID)
	formData.Set("sign_token", signToken)

	resp, err := s.client.Post(geetestValidateURL, "application/x-www-form-urlencoded",
		strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("geetest request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("geetest read response failed: %w", err)
	}

	var result struct {
		Result string `json:"result"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("geetest parse response failed: %w", err)
	}

	if result.Result != "success" {
		return errors.New("geetest verification failed: " + result.Reason)
	}
	return nil
}
