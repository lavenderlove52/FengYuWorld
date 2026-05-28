package dto

type UpdateThemeRequest struct {
	Preset         string `json:"preset"`
	AccentColor    string `json:"accent_color"`
	FontFamily     string `json:"font_family"`
	BannerURL      string `json:"banner_url"`
	Layout         string `json:"layout"`
	ShowPinsPublic *bool  `json:"show_pins_public"`
	CustomCSS      string `json:"custom_css"`
	CustomJS       string `json:"custom_js"`
}

type UserThemeResponse struct {
	UserID         int64  `json:"user_id"`
	Preset         string `json:"preset"`
	AccentColor    string `json:"accent_color"`
	FontFamily     string `json:"font_family"`
	BannerURL      string `json:"banner_url"`
	Layout         string `json:"layout"`
	ShowPinsPublic bool   `json:"show_pins_public"`
	CustomCSS      string `json:"custom_css"`
	CustomJS       string `json:"custom_js"`
}
