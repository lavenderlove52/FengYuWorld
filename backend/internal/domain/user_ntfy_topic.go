package domain

import "time"

type UserNtfyTopic struct {
	ID            int64     `gorm:"primaryKey" json:"id"`
	UserID        int64     `gorm:"not null;index" json:"user_id"`
	Topic         string    `gorm:"size:80;not null;uniqueIndex" json:"topic"`
	PublicEnabled bool      `gorm:"not null;default:false" json:"public_enabled"`
	PublicToken   string    `gorm:"size:100;not null;default:''" json:"public_token"`
	CreatedAt     time.Time `gorm:"not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (UserNtfyTopic) TableName() string { return "user_ntfy_topics" }
