package domain

import "time"

type EmailVerifyCode struct {
	ID        int64     `gorm:"primaryKey"`
	Email     string    `gorm:"size:120;not null;index:idx_evc_email"`
	Code      string    `gorm:"size:10;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (EmailVerifyCode) TableName() string { return "email_verify_codes" }
