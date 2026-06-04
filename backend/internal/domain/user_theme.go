package domain

import "time"

type ThemeEffectConfig struct {
	Scope      string                `json:"scope"`
	Background ThemeBackgroundEffect `json:"background"`
	Orbit      ThemeOrbitEffect      `json:"orbit"`
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

type ThemeOrbitEffect struct {
	Type    string  `json:"type"`
	Color   string  `json:"color"`
	Count   int     `json:"count"`
	Size    float64 `json:"size"`
	Opacity float64 `json:"opacity"`
	Speed   float64 `json:"speed"`
	Radius  float64 `json:"radius"`
}

type ThemeClickEffect struct {
	Type     string  `json:"type"`
	Color    string  `json:"color"`
	Count    int     `json:"count"`
	Size     float64 `json:"size"`
	Opacity  float64 `json:"opacity"`
	Duration int     `json:"duration"`
}

type ThemeHeroConfig struct {
	Enabled        bool    `json:"enabled"`
	ImageIDs       []int64 `json:"image_ids"`
	Transition     string  `json:"transition"`
	IntervalMS     int     `json:"interval_ms"`
	OverlayOpacity float64 `json:"overlay_opacity"`
	HeroHeightVH   int     `json:"hero_height_vh"`
	// PageWidthPX controls the personal homepage article content width.
	PageWidthPX int `json:"page_width_px"`
	// ArticleListWidthPX controls the post detail body width.
	ArticleListWidthPX int    `json:"article_list_width_px"`
	ArticleLayout      string `json:"article_layout"`
}

func DefaultThemeHeroConfig() ThemeHeroConfig {
	return ThemeHeroConfig{
		Enabled:            false,
		ImageIDs:           []int64{},
		Transition:         "fade",
		IntervalMS:         6000,
		OverlayOpacity:     0.4,
		HeroHeightVH:       100,
		PageWidthPX:        1280,
		ArticleListWidthPX: 960,
		ArticleLayout:      "list",
	}
}

func DefaultThemeEffectConfig() ThemeEffectConfig {
	return ThemeEffectConfig{
		Scope: "all",
		Background: ThemeBackgroundEffect{
			Type: "off", Color: "#6366f1", Count: 20, Size: 20, Opacity: 0.25, Speed: 0.5,
		},
		Orbit: ThemeOrbitEffect{
			Type: "off", Color: "#6366f1", Count: 8, Size: 5, Opacity: 0.8, Speed: 1, Radius: 48,
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
	UserID                 int64             `gorm:"primaryKey"                        json:"user_id"`
	Preset                 string            `gorm:"size:50;not null;default:default"  json:"preset"`
	AccentColor            string            `gorm:"size:20;not null;default:''"            json:"accent_color"`
	ThemeColorEnabled      bool              `gorm:"not null;default:false" json:"theme_color_enabled"`
	BackgroundColor        string            `gorm:"size:20;not null;default:''" json:"background_color"`
	BackgroundColorEnabled bool              `gorm:"not null;default:false" json:"background_color_enabled"`
	FontFamily             string            `gorm:"size:100;not null;default:''" json:"font_family"`
	BannerURL              string            `gorm:"size:500;not null;default:''"      json:"banner_url"`
	Layout                 string            `gorm:"size:20;not null;default:grid"     json:"layout"`
	ShowPinsPublic         bool              `gorm:"not null;default:false"            json:"show_pins_public"`
	CustomCSS              string            `gorm:"type:text;not null;default:''"     json:"custom_css"`
	CustomJS               string            `gorm:"type:text;not null;default:''"     json:"custom_js"`
	EffectConfig           ThemeEffectConfig `gorm:"type:jsonb;serializer:json;not null" json:"effect_config"`
	HeroConfig             ThemeHeroConfig   `gorm:"type:jsonb;serializer:json;not null" json:"hero_config"`
	UpdatedAt              time.Time         `gorm:"not null;default:now()"            json:"updated_at"`
}

func (UserTheme) TableName() string { return "user_themes" }
