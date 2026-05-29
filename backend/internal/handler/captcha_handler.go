package handler

import (
	"net/http"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CaptchaHandler struct {
	svc *service.CaptchaService
}

func NewCaptchaHandler(svc *service.CaptchaService) *CaptchaHandler {
	return &CaptchaHandler{svc: svc}
}

// GetCaptcha GET /api/auth/captcha
func (h *CaptchaHandler) GetCaptcha(c *gin.Context) {
	id, b64, err := h.svc.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "生成验证码失败"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "image": b64}))
}
