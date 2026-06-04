package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
)

var (
	ErrInvalidPreset       = errors.New("invalid preset")
	ErrInvalidLayout       = errors.New("invalid layout")
	ErrCSSToLong           = errors.New("custom_css exceeds 10000 characters")
	ErrJSToLong            = errors.New("custom_js exceeds 10000 characters")
	ErrInvalidAccentColor  = errors.New("invalid accent_color format")
	ErrInvalidEffectConfig = errors.New("invalid effect_config")
	ErrInvalidHeroConfig   = errors.New("invalid hero_config")
	ErrForbidden           = errors.New("forbidden")

	allowedPresets = map[string]bool{"default": true, "minimal": true, "dark": true, "forest": true, "ocean": true, "warm": true}
	allowedLayouts = map[string]bool{"grid": true, "list": true, "magazine": true}
	accentColorRe  = regexp.MustCompile(`^#[0-9a-fA-F]{3,6}$`)
	effectColorRe  = regexp.MustCompile(`^(?:#[0-9a-fA-F]{3}|#[0-9a-fA-F]{6})$`)

	dangerousCSS = []*regexp.Regexp{
		regexp.MustCompile(`(?i)javascript\s*:`),
		regexp.MustCompile(`(?i)expression\s*\(`),
		regexp.MustCompile(`(?i)@import`),
		regexp.MustCompile(`(?i)-moz-binding`),
		regexp.MustCompile(`(?i)behavior\s*:`),
	}
)

type UserThemeService struct {
	repo     *repository.UserThemeRepository
	userRepo *repository.UserRepository
	images   *repository.HomeThemeImageRepository
}

func NewUserThemeService(repo *repository.UserThemeRepository, userRepo *repository.UserRepository, images *repository.HomeThemeImageRepository) *UserThemeService {
	return &UserThemeService{repo: repo, userRepo: userRepo, images: images}
}

// GetByUsername 公开接口，不存在时返回默认值
func (s *UserThemeService) GetByUsername(username string) (*domain.UserTheme, error) {
	return s.repo.GetByUsername(username)
}

// Update 更新主题，含 CSS 过滤和字段校验
func (s *UserThemeService) Update(userID int64, req dto.UpdateThemeRequest) (*domain.UserTheme, error) {
	if req.Preset != "" && !allowedPresets[req.Preset] {
		return nil, ErrInvalidPreset
	}
	if req.Layout != "" && !allowedLayouts[req.Layout] {
		return nil, ErrInvalidLayout
	}
	if len(req.CustomCSS) > 10000 {
		return nil, ErrCSSToLong
	}
	if len(req.CustomJS) > 10000 {
		return nil, ErrJSToLong
	}
	if req.AccentColor != "" && !accentColorRe.MatchString(req.AccentColor) {
		return nil, ErrInvalidAccentColor
	}
	if req.BackgroundColor != "" && !accentColorRe.MatchString(req.BackgroundColor) {
		return nil, ErrInvalidAccentColor
	}
	if req.CustomJS != "" {
		user, err := s.userRepo.FindByID(userID)
		if err != nil || user == nil {
			return nil, ErrForbidden
		}
		if user.Role != "admin" && !user.CanCustomJS {
			return nil, ErrForbidden
		}
	}
	filteredCSS := filterCSS(req.CustomCSS)

	theme, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if req.Preset != "" {
		theme.Preset = req.Preset
	}
	theme.AccentColor = req.AccentColor
	if req.ThemeColorEnabled != nil {
		theme.ThemeColorEnabled = *req.ThemeColorEnabled
	}
	theme.BackgroundColor = req.BackgroundColor
	if req.BackgroundColorEnabled != nil {
		theme.BackgroundColorEnabled = *req.BackgroundColorEnabled
	}
	theme.FontFamily = req.FontFamily
	theme.BannerURL = req.BannerURL
	if req.Layout != "" {
		theme.Layout = req.Layout
	}
	if req.ShowPinsPublic != nil {
		theme.ShowPinsPublic = *req.ShowPinsPublic
	}
	theme.CustomCSS = filteredCSS
	theme.CustomJS = req.CustomJS
	if req.EffectConfig != nil {
		effectConfig := domain.ThemeEffectConfig(*req.EffectConfig)
		if effectConfig.Orbit.Type == "" {
			effectConfig.Orbit = domain.DefaultThemeEffectConfig().Orbit
		}
		if err := validateEffectConfig(effectConfig); err != nil {
			return nil, err
		}
		theme.EffectConfig = effectConfig
	}
	if req.HeroConfig != nil {
		heroConfig := domain.ThemeHeroConfig(*req.HeroConfig)
		if err := validateHeroConfig(heroConfig); err != nil {
			return nil, err
		}
		images, err := s.images.ListByUserIDAndIDs(userID, heroConfig.ImageIDs)
		if err != nil {
			return nil, err
		}
		if len(images) != len(heroConfig.ImageIDs) {
			return nil, invalidHeroConfig("image_ids")
		}
		theme.HeroConfig = heroConfig
	}
	theme.UpdatedAt = time.Now()

	if err := s.repo.Upsert(theme); err != nil {
		return nil, err
	}
	return theme, nil
}

func (s *UserThemeService) HeroImages(theme *domain.UserTheme) ([]domain.HomeThemeImage, error) {
	images, err := s.images.ListByUserIDAndIDs(theme.UserID, theme.HeroConfig.ImageIDs)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]domain.HomeThemeImage, len(images))
	for _, image := range images {
		byID[image.ID] = image
	}
	ordered := make([]domain.HomeThemeImage, 0, len(theme.HeroConfig.ImageIDs))
	for _, id := range theme.HeroConfig.ImageIDs {
		if image, ok := byID[id]; ok {
			ordered = append(ordered, image)
		}
	}
	return ordered, nil
}

func validateHeroConfig(config domain.ThemeHeroConfig) error {
	if len(config.ImageIDs) > 5 {
		return invalidHeroConfig("image_ids")
	}
	if config.Enabled && len(config.ImageIDs) == 0 {
		return invalidHeroConfig("image_ids")
	}
	seen := make(map[int64]struct{}, len(config.ImageIDs))
	for _, id := range config.ImageIDs {
		if id <= 0 {
			return invalidHeroConfig("image_ids")
		}
		if _, ok := seen[id]; ok {
			return invalidHeroConfig("image_ids")
		}
		seen[id] = struct{}{}
	}
	if !allowedValue(config.Transition, "off", "fade", "slices", "interlaced") {
		return invalidHeroConfig("transition")
	}
	if !withinInt(config.IntervalMS, 3000, 30000) {
		return invalidHeroConfig("interval_ms")
	}
	if allowedValue(config.Transition, "slices", "interlaced") && config.IntervalMS < 5000 {
		return invalidHeroConfig("interval_ms")
	}
	if !withinFloat(config.OverlayOpacity, 0, 0.9) {
		return invalidHeroConfig("overlay_opacity")
	}
	if !withinInt(config.HeroHeightVH, 60, 100) {
		return invalidHeroConfig("hero_height_vh")
	}
	if !withinInt(config.PageWidthPX, 600, 1800) {
		return invalidHeroConfig("page_width_px")
	}
	if !withinInt(config.ArticleListWidthPX, 600, 1600) {
		return invalidHeroConfig("article_list_width_px")
	}
	if !allowedValue(config.ArticleLayout, "list", "grid", "magazine") {
		return invalidHeroConfig("article_layout")
	}
	return nil
}

func validateEffectConfig(config domain.ThemeEffectConfig) error {
	if !allowedValue(config.Scope, "home", "post", "all") {
		return invalidEffectConfig("scope")
	}

	background := config.Background
	if !allowedValue(background.Type, "off", "floating-shapes", "starfield", "fireflies", "bubbles", "constellation", "rain") {
		return invalidEffectConfig("background.type")
	}
	if !validEffectColor(background.Color) ||
		!withinInt(background.Count, 1, 60) ||
		!withinFloat(background.Size, 4, 80) ||
		!withinFloat(background.Opacity, 0.05, 1) ||
		!withinFloat(background.Speed, 0.1, 2) {
		return invalidEffectConfig("background")
	}

	orbit := config.Orbit
	if !allowedValue(orbit.Type, "off", "rainbow-orbit", "atom-orbit", "spiral-orbit") {
		return invalidEffectConfig("orbit.type")
	}
	if !validEffectColor(orbit.Color) ||
		!withinInt(orbit.Count, 3, 24) ||
		!withinFloat(orbit.Size, 1, 16) ||
		!withinFloat(orbit.Opacity, 0.05, 1) ||
		!withinFloat(orbit.Speed, 0.1, 3) ||
		!withinFloat(orbit.Radius, 20, 120) {
		return invalidEffectConfig("orbit")
	}

	cursor := config.Cursor
	if !allowedValue(cursor.Type, "off", "particle-trail", "comet-trail", "sparkle-trail") {
		return invalidEffectConfig("cursor.type")
	}
	if !validEffectColor(cursor.Color) ||
		!withinInt(cursor.Count, 1, 50) ||
		!withinFloat(cursor.Size, 1, 20) ||
		!withinFloat(cursor.Opacity, 0.05, 1) ||
		!withinInt(cursor.Duration, 100, 2000) {
		return invalidEffectConfig("cursor")
	}

	click := config.Click
	if !allowedValue(click.Type, "off", "heart", "star", "dot", "ring", "diamond", "flower", "sparkle") {
		return invalidEffectConfig("click.type")
	}
	if !validEffectColor(click.Color) ||
		!withinInt(click.Count, 1, 24) ||
		!withinFloat(click.Size, 4, 60) ||
		!withinFloat(click.Opacity, 0.05, 1) ||
		!withinInt(click.Duration, 200, 3000) {
		return invalidEffectConfig("click")
	}

	return nil
}

func allowedValue(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func validEffectColor(color string) bool {
	return effectColorRe.MatchString(color)
}

func withinInt(value, min, max int) bool {
	return value >= min && value <= max
}

func withinFloat(value, min, max float64) bool {
	return value >= min && value <= max
}

func invalidEffectConfig(field string) error {
	return fmt.Errorf("%w: %s", ErrInvalidEffectConfig, field)
}

func invalidHeroConfig(field string) error {
	return fmt.Errorf("%w: %s", ErrInvalidHeroConfig, field)
}

func filterCSS(css string) string {
	for _, re := range dangerousCSS {
		css = re.ReplaceAllString(css, "")
	}
	return strings.TrimSpace(css)
}
