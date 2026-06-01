package handler

import (
	"errors"
	"net/http"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth      *service.AuthService
	emailCode *service.EmailCodeService
	captcha   *service.CaptchaService
}

func NewAuthHandler(auth *service.AuthService, emailCode *service.EmailCodeService, captcha *service.CaptchaService) *AuthHandler {
	return &AuthHandler{auth: auth, emailCode: emailCode, captcha: captcha}
}

func (h *AuthHandler) SendEmailCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.emailCode.SendCode(req.Email); err != nil {
		if errors.Is(err, service.ErrCodeTooFrequent) {
			c.JSON(http.StatusTooManyRequests, dto.Err(dto.CodeBadRequest, "发送过于频繁，请60秒后重试"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "发送失败"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}

	if err := h.captcha.Verify(req.CaptchaID, req.CaptchaAnswer); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "验证码错误或已过期"))
		return
	}

	resp, err := h.auth.Register(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDuplicateUser):
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeDuplicateUser, "用户名或邮箱已存在"))
		case errors.Is(err, service.ErrCodeInvalid):
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "邮箱验证码错误或已过期"))
		case errors.Is(err, service.ErrInviteCodeInvalid):
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "邀请码无效"))
		case errors.Is(err, service.ErrInviteCodeUsed):
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "邀请码已被使用"))
		case errors.Is(err, service.ErrInviteCodeExpired):
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "邀请码已过期"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(resp))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}

	if err := h.captcha.Verify(req.CaptchaID, req.CaptchaAnswer); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "验证码错误或已过期"))
		return
	}

	resp, err := h.auth.Login(req)
	if errors.Is(err, service.ErrLoginFailed) {
		c.JSON(http.StatusUnauthorized, dto.Err(dto.CodeLoginFailed, "登录失败"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(resp))
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	resp, err := h.auth.Refresh(req.RefreshToken)
	if errors.Is(err, service.ErrUnauthorized) {
		c.JSON(http.StatusUnauthorized, dto.Err(dto.CodeUnauthorized, "未登录"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(resp))
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(nil))
}

// ResetPassword POST /api/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.captcha.Verify(req.CaptchaID, req.CaptchaAnswer); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "验证码错误或已过期"))
		return
	}
	if err := h.auth.ResetPassword(req.Email, req.EmailCode, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "该邮箱未注册"))
		case errors.Is(err, service.ErrCodeInvalid):
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "邮箱验证码错误或已过期"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}
