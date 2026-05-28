package repository

import (
	"errors"
	"time"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

var ErrCodeInvalid = errors.New("email code invalid or expired")

type EmailCodeRepository struct {
	db *gorm.DB
}

func NewEmailCodeRepository(db *gorm.DB) *EmailCodeRepository {
	return &EmailCodeRepository{db: db}
}

func (r *EmailCodeRepository) Create(code *domain.EmailVerifyCode) error {
	return r.db.Create(code).Error
}

func (r *EmailCodeRepository) FindLatestByEmail(email string) (*domain.EmailVerifyCode, error) {
	var code domain.EmailVerifyCode
	err := r.db.Where("email = ?", email).Order("created_at DESC").First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &code, err
}

func (r *EmailCodeRepository) FindValidByEmailAndCode(email, code string) (*domain.EmailVerifyCode, error) {
	var evc domain.EmailVerifyCode
	err := r.db.Where(
		"email = ? AND code = ? AND used_at IS NULL AND expires_at > ?",
		email, code, time.Now(),
	).First(&evc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &evc, err
}

func (r *EmailCodeRepository) MarkUsed(id int64) error {
	now := time.Now()
	return r.db.Model(&domain.EmailVerifyCode{}).
		Where("id = ?", id).
		Update("used_at", now).Error
}

// ConsumeByEmailAndCode 原子校验并消费验证码，返回 nil 表示成功
func (r *EmailCodeRepository) ConsumeByEmailAndCode(email, code string) error {
	result := r.db.Model(&domain.EmailVerifyCode{}).
		Where("email = ? AND code = ? AND used_at IS NULL AND expires_at > ?", email, code, time.Now()).
		Update("used_at", time.Now())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrCodeInvalid
	}
	return nil
}

// InvalidateByEmail 作废某邮箱所有未使用的验证码
func (r *EmailCodeRepository) InvalidateByEmail(email string) error {
	return r.db.Model(&domain.EmailVerifyCode{}).
		Where("email = ? AND used_at IS NULL", email).
		Update("used_at", time.Now()).Error
}

// DeleteExpired 清理已过期且未使用的验证码记录
func (r *EmailCodeRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ? AND used_at IS NULL", time.Now()).
		Delete(&domain.EmailVerifyCode{}).Error
}
