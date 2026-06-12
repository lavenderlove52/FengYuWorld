package handler

import (
	"errors"
	"net/http"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type PushHandler struct {
	push          *service.PushService
	publicLimiter *service.RateLimiter
}

func NewPushHandler(push *service.PushService) *PushHandler {
	return &PushHandler{push: push, publicLimiter: service.NewRateLimiter(10, time.Minute)}
}

func (h *PushHandler) GetSetting(c *gin.Context) {
	setting, err := h.push.GetSetting(middleware.MustUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "system error"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(h.settingResponse(setting)))
}

func (h *PushHandler) UpdateSetting(c *gin.Context) {
	var req dto.UpdatePushSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid request"))
		return
	}
	setting, err := h.push.UpdateSetting(middleware.MustUserID(c), req.BarkEndpoint)
	if errors.Is(err, service.ErrInvalidPushEndpoint) {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid Bark endpoint"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "system error"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(h.settingResponse(setting)))
}

func (h *PushHandler) UpdateNtfyTopic(c *gin.Context) {
	var req dto.UpdateNtfySettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid topic"))
		return
	}
	setting, err := h.push.UpdateNtfyTopic(middleware.MustUserID(c), req.TopicSuffix)
	switch {
	case errors.Is(err, service.ErrInvalidNtfyTopic):
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "topic may only contain letters, numbers, hyphens and underscores"))
	case errors.Is(err, service.ErrNtfyNotConfigured):
		c.JSON(http.StatusServiceUnavailable, dto.Err(dto.CodeInternalError, "ntfy is not configured"))
	case errors.Is(err, service.ErrNtfyACLFailed):
		c.JSON(http.StatusBadGateway, dto.Err(dto.CodeInternalError, err.Error()))
	case err != nil:
		c.JSON(http.StatusConflict, dto.Err(dto.CodeConflict, "topic is unavailable"))
	default:
		c.JSON(http.StatusOK, dto.OK(h.settingResponse(setting)))
	}
}

func (h *PushHandler) UpdateNtfyPublic(c *gin.Context) {
	var req dto.UpdateNtfyPublicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid request"))
		return
	}
	setting, err := h.push.UpdateNtfyPublic(middleware.MustUserID(c), req.Enabled)
	if errors.Is(err, service.ErrPushNotConfigured) {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "save an ntfy topic first"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "system error"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(h.settingResponse(setting)))
}

func (h *PushHandler) SendNtfy(c *gin.Context) {
	var req dto.SendNtfyPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid request"))
		return
	}
	h.writeNtfyResult(c, h.push.SendNtfy(middleware.MustUserID(c), req.Mode, req.Title, req.Message, req.Priority, req.Tags))
}

func (h *PushHandler) GetPublicNtfy(c *gin.Context) {
	setting, err := h.push.GetPublicNtfy(c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "public push page not found"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(dto.PublicNtfyPushResponse{Topic: setting.NtfyTopic}))
}

func (h *PushHandler) SendPublicNtfy(c *gin.Context) {
	token := c.Param("token")
	if _, err := h.push.GetPublicNtfy(token); err != nil {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "public push page not found"))
		return
	}
	if !h.publicLimiter.Allow(token + ":" + c.RemoteIP()) {
		c.JSON(http.StatusTooManyRequests, dto.Err(dto.CodeBadRequest, "too many push requests"))
		return
	}
	var req dto.SendNtfyPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid request"))
		return
	}
	err := h.push.SendPublicNtfy(token, req.Mode, req.Title, req.Message, req.Priority, req.Tags)
	if errors.Is(err, service.ErrPublicPushNotFound) {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "public push page not found"))
		return
	}
	h.writeNtfyResult(c, err)
}

func (h *PushHandler) writeNtfyResult(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPushNotConfigured):
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "ntfy topic is not configured"))
	case errors.Is(err, service.ErrNtfyNotConfigured):
		c.JSON(http.StatusServiceUnavailable, dto.Err(dto.CodeInternalError, "ntfy is not configured"))
	case errors.Is(err, service.ErrInvalidPushMessage):
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid ntfy push request"))
	case err != nil:
		c.JSON(http.StatusBadGateway, dto.Err(dto.CodeInternalError, err.Error()))
	default:
		c.JSON(http.StatusOK, dto.OK(nil))
	}
}

func (h *PushHandler) settingResponse(setting *domain.UserPushSetting) dto.PushSettingResponse {
	return dto.PushSettingResponse{
		BarkEndpoint: setting.BarkEndpoint, NtfyBaseURL: h.push.NtfyBaseURL(), NtfyTopic: setting.NtfyTopic,
		NtfyPublicEnabled: setting.NtfyPublicEnabled, NtfyPublicToken: setting.NtfyPublicToken,
	}
}

func (h *PushHandler) SendBark(c *gin.Context) {
	var req dto.SendBarkPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid request"))
		return
	}
	err := h.push.SendBark(middleware.MustUserID(c), req.Mode, req.Title, req.Message)
	switch {
	case errors.Is(err, service.ErrPushNotConfigured):
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "Bark endpoint is not configured"))
	case errors.Is(err, service.ErrInvalidPushEndpoint), errors.Is(err, service.ErrInvalidPushMessage):
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "invalid Bark push request"))
	case err != nil:
		c.JSON(http.StatusBadGateway, dto.Err(dto.CodeInternalError, err.Error()))
	default:
		c.JSON(http.StatusOK, dto.OK(nil))
	}
}
