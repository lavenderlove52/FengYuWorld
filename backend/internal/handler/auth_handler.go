package handler

import (
	"net/http"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	resp, err := h.auth.Register(req)
	if err == service.ErrDuplicateUser {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeDuplicateUser, "用户名或邮箱已存在"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
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
	resp, err := h.auth.Login(req)
	if err == service.ErrLoginFailed {
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
	if err == service.ErrUnauthorized {
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
