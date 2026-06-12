package domain

import "time"

type UserPushSetting struct {
	UserID            int64     `gorm:"primaryKey" json:"user_id"`
	BarkEndpoint      string    `gorm:"size:1000;not null;default:''" json:"bark_endpoint"`
	NtfyTopic         string    `gorm:"size:80;not null;default:''" json:"ntfy_topic"`
	NtfyPublicEnabled bool      `gorm:"not null;default:false" json:"ntfy_public_enabled"`
	NtfyPublicToken   string    `gorm:"size:100;not null;default:''" json:"ntfy_public_token"`
	UpdatedAt         time.Time `gorm:"not null;default:now()" json:"updated_at"`
}

func (UserPushSetting) TableName() string { return "user_push_settings" }
