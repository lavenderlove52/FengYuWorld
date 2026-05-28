package domain

import "time"

type UserTheme struct {
	UserID         int64     `gorm:"primaryKey"                        json:"user_id"`
	Preset         string    `gorm:"size:50;not null;default:default"  json:"preset"`
	AccentColor    string    `gorm:"size:20;not null;default:''"       json:"accent_color"`
	FontFamily     string    `gorm:"size:100;not null;default:''"      json:"font_family"`
	BannerURL      string    `gorm:"size:500;not null;default:''"      json:"banner_url"`
	Layout         string    `gorm:"size:20;not null;default:grid"     json:"layout"`
	ShowPinsPublic bool      `gorm:"not null;default:false"            json:"show_pins_public"`
	CustomCSS      string    `gorm:"type:text;not null;default:''"     json:"custom_css"`
	CustomJS       string    `gorm:"type:text;not null;default:''"     json:"custom_js"`
	UpdatedAt      time.Time `gorm:"not null;default:now()"            json:"updated_at"`
}

func (UserTheme) TableName() string { return "user_themes" }
