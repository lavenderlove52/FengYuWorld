package domain

import "time"

const AppSettingFriendChatEnabled = "friend_chat_enabled"

type AppSetting struct {
	Key       string `gorm:"size:80;primaryKey"`
	Value     string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AppSetting) TableName() string { return "app_settings" }
