package repository

import (
	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserThemeRepository struct {
	db *gorm.DB
}

func NewUserThemeRepository(db *gorm.DB) *UserThemeRepository {
	return &UserThemeRepository{db: db}
}

func defaultTheme(userID int64) *domain.UserTheme {
	return &domain.UserTheme{
		UserID:                 userID,
		Preset:                 "default",
		AccentColor:            "",
		ThemeColorEnabled:      false,
		BackgroundColor:        "",
		BackgroundColorEnabled: false,
		FontFamily:             "",
		BannerURL:              "",
		Layout:                 "grid",
		ShowPinsPublic:         false,
		CustomCSS:              "",
		CustomJS:               "",
		EffectConfig:           domain.DefaultThemeEffectConfig(),
		HeroConfig:             domain.DefaultThemeHeroConfig(),
	}
}

// GetByUsername 通过用户名查找主题；若不存在返回默认主题（不返回错误）
func (r *UserThemeRepository) GetByUsername(username string) (*domain.UserTheme, error) {
	var theme domain.UserTheme
	err := r.db.
		Joins("JOIN users ON users.id = user_themes.user_id").
		Where("users.username = ?", username).
		First(&theme).Error
	if err == gorm.ErrRecordNotFound {
		var userID int64
		if e := r.db.Model(&domain.User{}).Select("id").Where("username = ?", username).Scan(&userID).Error; e != nil {
			return nil, e
		}
		return defaultTheme(userID), nil
	}
	return &theme, err
}

// GetByUserID 通过用户 ID 查找主题；若不存在返回默认主题（不返回错误）
func (r *UserThemeRepository) GetByUserID(userID int64) (*domain.UserTheme, error) {
	var theme domain.UserTheme
	err := r.db.Where("user_id = ?", userID).First(&theme).Error
	if err == gorm.ErrRecordNotFound {
		return defaultTheme(userID), nil
	}
	return &theme, err
}

// Upsert 插入或更新主题
func (r *UserThemeRepository) Upsert(theme *domain.UserTheme) error {
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"preset", "accent_color", "theme_color_enabled", "background_color", "background_color_enabled", "font_family", "banner_url",
			"layout", "show_pins_public", "custom_css", "custom_js", "effect_config", "hero_config", "updated_at",
		}),
	}).Create(theme).Error
}
