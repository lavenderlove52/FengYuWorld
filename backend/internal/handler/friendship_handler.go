package handler

import (
	"net/http"
	"strconv"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/service"

	"github.com/gin-gonic/gin"
)

type FriendshipHandler struct {
	friends *service.FriendshipService
}

func NewFriendshipHandler(friends *service.FriendshipService) *FriendshipHandler {
	return &FriendshipHandler{friends: friends}
}

func (h *FriendshipHandler) SendRequest(c *gin.Context) {
	userID := middleware.MustUserID(c)
	var req dto.FriendRequestCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "请求参数错误"))
		return
	}
	item, err := h.friends.SendRequest(userID, req)
	if err == service.ErrFriendNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "用户不存在"))
		return
	}
	if err == service.ErrCannotFriendSelf {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "不能添加自己"))
		return
	}
	if err == service.ErrFriendConflict {
		c.JSON(http.StatusConflict, dto.Err(dto.CodeConflict, "好友申请冲突"))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *FriendshipHandler) ListReceived(c *gin.Context) {
	userID := middleware.MustUserID(c)
	list, err := h.friends.ListReceived(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(list))
}

func (h *FriendshipHandler) ListSent(c *gin.Context) {
	userID := middleware.MustUserID(c)
	list, err := h.friends.ListSent(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(list))
}

func (h *FriendshipHandler) Accept(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.friends.Accept(userID, id); err == service.ErrFriendNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "申请不存在"))
		return
	} else if err == service.ErrFriendForbidden {
		c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限"))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *FriendshipHandler) Reject(c *gin.Context) {
	userID := middleware.MustUserID(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.friends.Reject(userID, id); err == service.ErrFriendNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "申请不存在"))
		return
	} else if err == service.ErrFriendForbidden {
		c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "无权限"))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}

func (h *FriendshipHandler) ListFriends(c *gin.Context) {
	userID := middleware.MustUserID(c)
	list, err := h.friends.ListFriends(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(list))
}

func (h *FriendshipHandler) DeleteFriend(c *gin.Context) {
	userID := middleware.MustUserID(c)
	friendID, _ := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err := h.friends.DeleteFriend(userID, friendID); err == service.ErrFriendNotFound {
		c.JSON(http.StatusNotFound, dto.Err(dto.CodeNotFound, "好友不存在"))
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}
