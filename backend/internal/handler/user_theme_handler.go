package handler

import (
	"errors"
	"net/http"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserThemeHandler struct {
	svc *service.UserThemeService
}

func NewUserThemeHandler(svc *service.UserThemeService) *UserThemeHandler {
	return &UserThemeHandler{svc: svc}
}

// GetTheme GET /api/users/:username/theme — 公开
func (h *UserThemeHandler) GetTheme(c *gin.Context) {
	username := c.Param("username")
	theme, err := h.svc.GetByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, err.Error()))
		return
	}
	heroImages, err := h.svc.HeroImages(theme)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, dto.OK(dto.UserThemeResponse{
		UserID:                 theme.UserID,
		Preset:                 theme.Preset,
		AccentColor:            theme.AccentColor,
		ThemeColorEnabled:      theme.ThemeColorEnabled,
		BackgroundColor:        theme.BackgroundColor,
		BackgroundColorEnabled: theme.BackgroundColorEnabled,
		FontFamily:             theme.FontFamily,
		BannerURL:              theme.BannerURL,
		Layout:                 theme.Layout,
		ShowPinsPublic:         theme.ShowPinsPublic,
		CustomCSS:              theme.CustomCSS,
		CustomJS:               theme.CustomJS,
		EffectConfig:           theme.EffectConfig,
		HeroConfig:             theme.HeroConfig,
		HeroImages:             heroImages,
	}))
}

// UpdateTheme PUT /api/me/theme — 需登录
func (h *UserThemeHandler) UpdateTheme(c *gin.Context) {
	userID := middleware.MustUserID(c)

	var req dto.UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, err.Error()))
		return
	}

	theme, err := h.svc.Update(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidPreset),
			errors.Is(err, service.ErrInvalidLayout),
			errors.Is(err, service.ErrCSSToLong),
			errors.Is(err, service.ErrJSToLong),
			errors.Is(err, service.ErrInvalidAccentColor),
			errors.Is(err, service.ErrInvalidEffectConfig),
			errors.Is(err, service.ErrInvalidHeroConfig):
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, err.Error()))
		case errors.Is(err, service.ErrForbidden):
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无自定义 JS 权限"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, err.Error()))
		}
		return
	}
	heroImages, err := h.svc.HeroImages(theme)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.OK(dto.UserThemeResponse{
		UserID:                 theme.UserID,
		Preset:                 theme.Preset,
		AccentColor:            theme.AccentColor,
		ThemeColorEnabled:      theme.ThemeColorEnabled,
		BackgroundColor:        theme.BackgroundColor,
		BackgroundColorEnabled: theme.BackgroundColorEnabled,
		FontFamily:             theme.FontFamily,
		BannerURL:              theme.BannerURL,
		Layout:                 theme.Layout,
		ShowPinsPublic:         theme.ShowPinsPublic,
		CustomCSS:              theme.CustomCSS,
		CustomJS:               theme.CustomJS,
		EffectConfig:           theme.EffectConfig,
		HeroConfig:             theme.HeroConfig,
		HeroImages:             heroImages,
	}))
}
