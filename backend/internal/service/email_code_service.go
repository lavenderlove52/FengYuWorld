package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"
)

var (
	ErrCodeTooFrequent = errors.New("email code too frequent")
	ErrCodeInvalid     = errors.New("email code invalid or expired")
)

const emailCodeCharset = "0123456789"

type EmailCodeService struct {
	repo  *repository.EmailCodeRepository
	email *EmailService
}

func NewEmailCodeService(repo *repository.EmailCodeRepository, email *EmailService) *EmailCodeService {
	return &EmailCodeService{repo: repo, email: email}
}

// SendCode 限流+生成+发送：60秒内同邮箱有记录则返回 ErrCodeTooFrequent
// 发送前作废旧码，并异步清理过期记录
func (s *EmailCodeService) SendCode(emailAddr string) error {
	latest, err := s.repo.FindLatestByEmail(emailAddr)
	if err != nil {
		return err
	}
	if latest != nil && time.Since(latest.CreatedAt) < 60*time.Second {
		return ErrCodeTooFrequent
	}

	// 作废该邮箱所有未使用的旧验证码
	if err := s.repo.InvalidateByEmail(emailAddr); err != nil {
		return err
	}

	// 异步清理全局过期记录
	go func() { _ = s.repo.DeleteExpired() }()

	code, err := generateNumericCode(6)
	if err != nil {
		return fmt.Errorf("generate code: %w", err)
	}

	evc := &domain.EmailVerifyCode{
		Email:     emailAddr,
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := s.repo.Create(evc); err != nil {
		return err
	}

	return s.email.SendVerifyCode(emailAddr, code)
}

// VerifyAndConsume 原子校验并消费验证码
func (s *EmailCodeService) VerifyAndConsume(emailAddr, code string) error {
	if err := s.repo.ConsumeByEmailAndCode(emailAddr, code); err != nil {
		if errors.Is(err, repository.ErrCodeInvalid) {
			return ErrCodeInvalid
		}
		return err
	}
	return nil
}

func generateNumericCode(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(emailCodeCharset))))
		if err != nil {
			return "", err
		}
		result[i] = emailCodeCharset[n.Int64()]
	}
	return string(result), nil
}
