package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type InviteHandler struct {
	svc *service.InviteService
}

func NewInviteHandler(svc *service.InviteService) *InviteHandler {
	return &InviteHandler{svc: svc}
}

// ListInviteCodes GET /api/admin/invite-codes
func (h *InviteHandler) ListInviteCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	codes, total, err := h.svc.List(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"items":     codes,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}

// GenerateInviteCodes POST /api/admin/invite-codes
func (h *InviteHandler) GenerateInviteCodes(c *gin.Context) {
	var req struct {
		Count     int    `json:"count" binding:"required,min=1,max=100"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "expires_at 格式错误，请使用 RFC3339 格式"))
			return
		}
		expiresAt = &t
	}

	userID := c.GetInt64("userID")
	codes, err := h.svc.GenerateCodes(userID, req.Count, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(codes))
}

// DeleteInviteCode DELETE /api/admin/invite-codes/:id
func (h *InviteHandler) DeleteInviteCode(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "无效的 ID"))
		return
	}
	if err := h.svc.Delete(id); err != nil {
		if errors.Is(err, service.ErrInviteCodeNotFound) {
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeBadRequest, "邀请码不存在或已被使用"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}
