package handler

import (
	"net/http"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	settings *service.AppSettingService
}

func NewSettingsHandler(settings *service.AppSettingService) *SettingsHandler {
	return &SettingsHandler{settings: settings}
}

func (h *SettingsHandler) GetPublicSettings(c *gin.Context) {
	settings, err := h.settings.GetPublicSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "system error"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(dto.PublicSettingsResponse{
		FriendChatEnabled: settings.FriendChatEnabled,
	}))
}

func (h *SettingsHandler) UpdateFriendChat(c *gin.Context) {
	var req dto.UpdateFriendChatSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid request"))
		return
	}

	settings, err := h.settings.UpdateFriendChatEnabled(middleware.MustUserID(c), req.Enabled)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "user not found"))
		case service.ErrSuperAdminForbidden:
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "requires super admin permission"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "system error"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(dto.PublicSettingsResponse{
		FriendChatEnabled: settings.FriendChatEnabled,
	}))
}
