package domain

import "time"

type ThemeEffectConfig struct {
	Scope      string                `json:"scope"`
	Background ThemeBackgroundEffect `json:"background"`
	Cursor     ThemeCursorEffect     `json:"cursor"`
	Click      ThemeClickEffect      `json:"click"`
}

type ThemeBackgroundEffect struct {
	Type    string  `json:"type"`
	Color   string  `json:"color"`
	Count   int     `json:"count"`
	Size    float64 `json:"size"`
	Opacity float64 `json:"opacity"`
	Speed   float64 `json:"speed"`
}

type ThemeCursorEffect struct {
	Type     string  `json:"type"`
	Color    string  `json:"color"`
	Count    int     `json:"count"`
	Size     float64 `json:"size"`
	Opacity  float64 `json:"opacity"`
	Duration int     `json:"duration"`
}

type ThemeClickEffect struct {
	Type     string  `json:"type"`
	Color    string  `json:"color"`
	Count    int     `json:"count"`
	Size     float64 `json:"size"`
	Opacity  float64 `json:"opacity"`
	Duration int     `json:"duration"`
}

func DefaultThemeEffectConfig() ThemeEffectConfig {
	return ThemeEffectConfig{
		Scope: "all",
		Background: ThemeBackgroundEffect{
			Type: "off", Color: "#6366f1", Count: 20, Size: 20, Opacity: 0.25, Speed: 0.5,
		},
		Cursor: ThemeCursorEffect{
			Type: "off", Color: "#6366f1", Count: 16, Size: 5, Opacity: 0.7, Duration: 700,
		},
		Click: ThemeClickEffect{
			Type: "off", Color: "#ec4899", Count: 8, Size: 16, Opacity: 0.9, Duration: 900,
		},
	}
}

type UserTheme struct {
	UserID         int64             `gorm:"primaryKey"                        json:"user_id"`
	Preset         string            `gorm:"size:50;not null;default:default"  json:"preset"`
	AccentColor    string            `gorm:"size:20;not null;default:''"       json:"accent_color"`
	FontFamily     string            `gorm:"size:100;not null;default:''"      json:"font_family"`
	BannerURL      string            `gorm:"size:500;not null;default:''"      json:"banner_url"`
	Layout         string            `gorm:"size:20;not null;default:grid"     json:"layout"`
	ShowPinsPublic bool              `gorm:"not null;default:false"            json:"show_pins_public"`
	CustomCSS      string            `gorm:"type:text;not null;default:''"     json:"custom_css"`
	CustomJS       string            `gorm:"type:text;not null;default:''"     json:"custom_js"`
	EffectConfig   ThemeEffectConfig `gorm:"type:jsonb;serializer:json;not null" json:"effect_config"`
	UpdatedAt      time.Time         `gorm:"not null;default:now()"            json:"updated_at"`
}

func (UserTheme) TableName() string { return "user_themes" }
