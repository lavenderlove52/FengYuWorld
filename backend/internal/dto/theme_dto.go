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

type UpdateThemeRequest struct {
	Preset         string                    `json:"preset"`
	AccentColor    string                    `json:"accent_color"`
	FontFamily     string                    `json:"font_family"`
	BannerURL      string                    `json:"banner_url"`
	Layout         string                    `json:"layout"`
	ShowPinsPublic *bool                     `json:"show_pins_public"`
	CustomCSS      string                    `json:"custom_css"`
	CustomJS       string                    `json:"custom_js"`
	EffectConfig   *ThemeEffectConfigRequest `json:"effect_config"`
}

type UserThemeResponse struct {
	UserID         int64                    `json:"user_id"`
	Preset         string                   `json:"preset"`
	AccentColor    string                   `json:"accent_color"`
	FontFamily     string                   `json:"font_family"`
	BannerURL      string                   `json:"banner_url"`
	Layout         string                   `json:"layout"`
	ShowPinsPublic bool                     `json:"show_pins_public"`
	CustomCSS      string                   `json:"custom_css"`
	CustomJS       string                   `json:"custom_js"`
	EffectConfig   domain.ThemeEffectConfig `json:"effect_config"`
}
