package service

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
)

var (
	ErrInvalidPreset      = errors.New("invalid preset")
	ErrInvalidLayout      = errors.New("invalid layout")
	ErrCSSToLong          = errors.New("custom_css exceeds 10000 characters")
	ErrInvalidAccentColor = errors.New("invalid accent_color format")
	ErrJSTooLong          = errors.New("custom_js exceeds 50000 characters")
	ErrForbidden          = errors.New("forbidden")

	allowedPresets = map[string]bool{"default": true, "minimal": true, "dark": true, "forest": true, "ocean": true, "warm": true}
	allowedLayouts = map[string]bool{"grid": true, "list": true, "magazine": true}
	accentColorRe  = regexp.MustCompile(`^#[0-9a-fA-F]{3,6}$`)

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
	if req.AccentColor != "" && !accentColorRe.MatchString(req.AccentColor) {
		return nil, ErrInvalidAccentColor
	}
	if len(req.CustomJS) > 50000 {
		return nil, ErrJSTooLong
	}

	if req.CustomJS != "" {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, ErrUserNotFound
		}
		if !user.CanCustomJS && user.Role != domain.UserRoleAdmin {
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
	theme.UpdatedAt = time.Now()

	if err := s.repo.Upsert(theme); err != nil {
		return nil, err
	}
	return theme, nil
}

func filterCSS(css string) string {
	for _, re := range dangerousCSS {
		css = re.ReplaceAllString(css, "")
	}
	return strings.TrimSpace(css)
}
