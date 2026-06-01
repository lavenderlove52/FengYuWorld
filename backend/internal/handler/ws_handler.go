package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub        *ws.Hub
	msgRepo    *repository.MessageRepository
	friendRepo *repository.FriendshipRepository
}

func NewWSHandler(hub *ws.Hub, msgRepo *repository.MessageRepository, friendRepo *repository.FriendshipRepository) *WSHandler {
	return &WSHandler{hub: hub, msgRepo: msgRepo, friendRepo: friendRepo}
}

// HandleWS GET /api/ws/chat（需登录）
func (h *WSHandler) HandleWS(c *gin.Context) {
	userID := middleware.MustUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &ws.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// 推送未读消息数
	h.pushUnreadCounts(client)

	// 写协程：从 Send 通道取数据写入 WebSocket
	go func() {
		defer conn.Close()
		for data := range client.Send {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()

	// 读循环
	conn.SetReadLimit(64 * 1024)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var in ws.InMsg
		if err := conn.ReadJSON(&in); err != nil {
			break
		}
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))

		switch in.Type {
		case "msg":
			h.handleMsg(client, in)
		case "read":
			h.handleRead(client, in)
		}
	}
}

func (h *WSHandler) handleMsg(client *ws.Client, in ws.InMsg) {
	if in.To == 0 || in.Content == "" {
		return
	}

	// 校验消息类型，无效类型回退为 text
	validTypes := map[string]bool{"text": true, "image": true, "emoji": true}
	if !validTypes[in.MsgType] {
		in.MsgType = "text"
	}

	// 内容长度限制
	if len(in.Content) > 4096 {
		client.Send <- marshalOut(ws.OutMsg{Type: "error", Payload: "消息内容过长"})
		return
	}

	// image 类型校验 URL（只允许本站 /uploads/ 路径或 http/https 开头）
	if in.MsgType == "image" && !strings.HasPrefix(in.Content, "/uploads/") && !strings.HasPrefix(in.Content, "http") {
		client.Send <- marshalOut(ws.OutMsg{Type: "error", Payload: "图片地址无效"})
		return
	}

	// 校验是否好友
	ok, err := h.friendRepo.AreFriends(client.UserID, in.To)
	if err != nil || !ok {
		client.Send <- marshalOut(ws.OutMsg{Type: "error", Payload: "非好友关系"})
		return
	}

	msgType := domain.MessageType(in.MsgType)
	if msgType == "" {
		msgType = domain.MessageTypeText
	}

	msg := &domain.Message{
		FromUser: client.UserID,
		ToUser:   in.To,
		Content:  in.Content,
		Type:     msgType,
	}
	if err := h.msgRepo.Create(msg); err != nil {
		client.Send <- marshalOut(ws.OutMsg{Type: "error", Payload: "消息发送失败"})
		return
	}

	out := ws.OutMsg{Type: "msg", Payload: msg}
	// 推送给接收方
	h.hub.SendToUser(in.To, out)
	// 回显给发送方（同步给其他设备）
	h.hub.SendToUser(client.UserID, out)
}

func (h *WSHandler) handleRead(client *ws.Client, in ws.InMsg) {
	friendID := in.FriendID
	if friendID == 0 {
		return
	}

	// 校验是否好友，防止越权标记
	ok, err := h.friendRepo.AreFriends(client.UserID, friendID)
	if err != nil || !ok {
		return
	}

	if err := h.msgRepo.MarkRead(friendID, client.UserID); err != nil {
		return
	}
	// 推送 unread_update 给对方
	h.hub.SendToUser(friendID, ws.OutMsg{
		Type:    "unread_update",
		Payload: map[string]interface{}{"friend_id": client.UserID, "count": 0},
	})
	// 更新自己的未读数
	h.pushUnreadCounts(client)
}

func (h *WSHandler) pushUnreadCounts(client *ws.Client) {
	counts, err := h.msgRepo.GetUnreadCounts(client.UserID)
	if err != nil {
		return
	}
	client.Send <- marshalOut(ws.OutMsg{Type: "unread_counts", Payload: counts})
}

func marshalOut(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"type":"error","payload":"marshal error"}`)
	}
	return data
}
