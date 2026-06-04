package dto

import (
	"bytes"
	"encoding/json"

	"my-blog/backend/internal/domain"
)

type ThemeEffectConfigRequest domain.ThemeEffectConfig

func (c *ThemeEffectConfigRequest) UnmarshalJSON(data []byte) error {
	type configAlias ThemeEffectConfigRequest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*configAlias)(c))
}

type ThemeHeroConfigRequest domain.ThemeHeroConfig

func (c *ThemeHeroConfigRequest) UnmarshalJSON(data []byte) error {
	type configAlias ThemeHeroConfigRequest
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode((*configAlias)(c))
}

type UpdateThemeRequest struct {
	Preset                 string                    `json:"preset"`
	AccentColor            string                    `json:"accent_color"`
	ThemeColorEnabled      *bool                     `json:"theme_color_enabled"`
	BackgroundColor        string                    `json:"background_color"`
	BackgroundColorEnabled *bool                     `json:"background_color_enabled"`
	FontFamily             string                    `json:"font_family"`
	BannerURL              string                    `json:"banner_url"`
	Layout                 string                    `json:"layout"`
	ShowPinsPublic         *bool                     `json:"show_pins_public"`
	CustomCSS              string                    `json:"custom_css"`
	CustomJS               string                    `json:"custom_js"`
	EffectConfig           *ThemeEffectConfigRequest `json:"effect_config"`
	HeroConfig             *ThemeHeroConfigRequest   `json:"hero_config"`
}

type UserThemeResponse struct {
	UserID                 int64                    `json:"user_id"`
	Preset                 string                   `json:"preset"`
	AccentColor            string                   `json:"accent_color"`
	ThemeColorEnabled      bool                     `json:"theme_color_enabled"`
	BackgroundColor        string                   `json:"background_color"`
	BackgroundColorEnabled bool                     `json:"background_color_enabled"`
	FontFamily             string                   `json:"font_family"`
	BannerURL              string                   `json:"banner_url"`
	Layout                 string                   `json:"layout"`
	ShowPinsPublic         bool                     `json:"show_pins_public"`
	CustomCSS              string                   `json:"custom_css"`
	CustomJS               string                   `json:"custom_js"`
	EffectConfig           domain.ThemeEffectConfig `json:"effect_config"`
	HeroConfig             domain.ThemeHeroConfig   `json:"hero_config"`
	HeroImages             []domain.HomeThemeImage  `json:"hero_images"`
}
