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
}

func NewUserThemeService(repo *repository.UserThemeRepository, userRepo *repository.UserRepository) *UserThemeService {
	return &UserThemeService{repo: repo, userRepo: userRepo}
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
		if err := validateEffectConfig(effectConfig); err != nil {
			return nil, err
		}
		theme.EffectConfig = effectConfig
	}
	theme.UpdatedAt = time.Now()

	if err := s.repo.Upsert(theme); err != nil {
		return nil, err
	}
	return theme, nil
}

func validateEffectConfig(config domain.ThemeEffectConfig) error {
	if !allowedValue(config.Scope, "home", "post", "all") {
		return invalidEffectConfig("scope")
	}

	background := config.Background
	if !allowedValue(background.Type, "off", "floating-shapes") {
		return invalidEffectConfig("background.type")
	}
	if !validEffectColor(background.Color) ||
		!withinInt(background.Count, 1, 60) ||
		!withinFloat(background.Size, 4, 80) ||
		!withinFloat(background.Opacity, 0.05, 1) ||
		!withinFloat(background.Speed, 0.1, 2) {
		return invalidEffectConfig("background")
	}

	cursor := config.Cursor
	if !allowedValue(cursor.Type, "off", "particle-trail") {
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
	if !allowedValue(click.Type, "off", "heart", "star", "dot") {
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

func filterCSS(css string) string {
	for _, re := range dangerousCSS {
		css = re.ReplaceAllString(css, "")
	}
	return strings.TrimSpace(css)
}
