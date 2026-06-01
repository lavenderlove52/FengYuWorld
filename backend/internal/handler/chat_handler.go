package handler

import (
	"net/http"
	"strconv"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	msgRepo    *repository.MessageRepository
	friendRepo *repository.FriendshipRepository
}

func NewChatHandler(msgRepo *repository.MessageRepository, friendRepo *repository.FriendshipRepository) *ChatHandler {
	return &ChatHandler{msgRepo: msgRepo, friendRepo: friendRepo}
}

// GetHistory GET /api/chat/:friend_id?before=<id>&limit=50
func (h *ChatHandler) GetHistory(c *gin.Context) {
	userID := middleware.MustUserID(c)
	friendID, err := strconv.ParseInt(c.Param("friend_id"), 10, 64)
	if err != nil || friendID == 0 {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "无效的好友ID"))
		return
	}

	ok, err := h.friendRepo.AreFriends(userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	if !ok {
		c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "非好友关系"))
		return
	}

	before, _ := strconv.ParseInt(c.Query("before"), 10, 64)
	limit := 50

	msgs, err := h.msgRepo.GetConversation(userID, friendID, before, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}

	// 反转为正序（DB 返回的是倒序）
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	c.JSON(http.StatusOK, dto.OK(msgs))
}

// GetUnreadCounts GET /api/chat/unread
func (h *ChatHandler) GetUnreadCounts(c *gin.Context) {
	userID := middleware.MustUserID(c)
	counts, err := h.msgRepo.GetUnreadCounts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(counts))
}

// MarkRead POST /api/chat/:friend_id/read
func (h *ChatHandler) MarkRead(c *gin.Context) {
	userID := middleware.MustUserID(c)
	friendID, err := strconv.ParseInt(c.Param("friend_id"), 10, 64)
	if err != nil || friendID == 0 {
		c.JSON(http.StatusBadRequest, dto.Err(dto.CodeBadRequest, "无效的好友ID"))
		return
	}

	ok, err := h.friendRepo.AreFriends(userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	if !ok {
		c.JSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "非好友关系"))
		return
	}

	if err := h.msgRepo.MarkRead(friendID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
		return
	}
	c.JSON(http.StatusOK, dto.OK(nil))
}
