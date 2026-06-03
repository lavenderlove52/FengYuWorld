package domain

import "time"

type RefreshToken struct {
	TokenHash string     `gorm:"primaryKey;size:64"`
	UserID    int64      `gorm:"not null;index"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	RevokedAt *time.Time `gorm:"index"`
	CreatedAt time.Time
}

func (RefreshToken) TableName() string { return "refresh_tokens" }
