package handler

import (
	"net/http"
	"strconv"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users *service.UserService
	posts *service.PostService
}

func NewUserHandler(users *service.UserService, posts *service.PostService) *UserHandler {
	return &UserHandler{users: users, posts: posts}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := middleware.MustUserID(c)
	info, err := h.users.GetMe(userID)
	if err == service.ErrUserNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(info))
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID := middleware.MustUserID(c)
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	info, err := h.users.UpdateMe(userID, req)
	if err == service.ErrUserNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(info))
}

func (h *UserHandler) GetUserPosts(c *gin.Context) {
	username := c.Param("username")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	if _, err := h.users.GetByUsername(username); err == service.ErrUserNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		return
	}

	result, err := h.posts.ListByUsername(username, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

func (h *UserHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	user, err := h.users.GetByUsername(username)
	if err == service.ErrUserNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	info := dto.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Bio:       user.Bio,
	}
	c.JSON(http.StatusOK, dto.OK(info))
}

// GrantCustomJS POST /api/admin/users/:user_id/custom-js
func (h *UserHandler) GrantCustomJS(c *gin.Context) {
	adminID := middleware.MustUserID(c)
	targetUserID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.users.GrantCustomJS(adminID, targetUserID); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		case service.ErrAdminForbidden:
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

// RevokeCustomJS DELETE /api/admin/users/:user_id/custom-js
func (h *UserHandler) RevokeCustomJS(c *gin.Context) {
	adminID := middleware.MustUserID(c)
	targetUserID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	if err := h.users.RevokeCustomJS(adminID, targetUserID); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		case service.ErrAdminForbidden:
			c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限"))
		default:
			c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		}
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.users.ListUsers(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}
