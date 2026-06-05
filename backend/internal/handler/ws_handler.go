package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/service"
	"my-blog/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     sameOrigin,
}

type WSHandler struct {
	hub        *ws.Hub
	msgRepo    *repository.MessageRepository
	friendRepo *repository.FriendshipRepository
	tickets    *service.WSTicketStore
	settings   *service.AppSettingService
}

func NewWSHandler(hub *ws.Hub, msgRepo *repository.MessageRepository, friendRepo *repository.FriendshipRepository, tickets *service.WSTicketStore, settings *service.AppSettingService) *WSHandler {
	return &WSHandler{hub: hub, msgRepo: msgRepo, friendRepo: friendRepo, tickets: tickets, settings: settings}
}

func (h *WSHandler) HandleWS(c *gin.Context) {
	if enabled, err := h.settings.IsFriendChatEnabled(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "system error"})
		return
	} else if !enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "friend chat is disabled"})
		return
	}

	userID, ok := h.tickets.Consume(c.Query("ticket"))
	if !ok || userID == 0 {
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

	h.pushUnreadCounts(client)

	go func() {
		defer conn.Close()
		for data := range client.Send {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}()

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
	if enabled, err := h.settings.IsFriendChatEnabled(); err != nil || !enabled {
		client.Send <- marshalOut(ws.OutMsg{Type: "error", Payload: "friend chat is disabled"})
		return
	}

	if in.To == 0 || in.Content == "" {
		return
	}

	validTypes := map[string]bool{"text": true, "image": true, "emoji": true}
	if !validTypes[in.MsgType] {
		in.MsgType = "text"
	}

	if len(in.Content) > 4096 {
		client.Send <- marshalOut(ws.OutMsg{Type: "error", Payload: "消息内容过长"})
		return
	}

	if in.MsgType == "image" && !strings.HasPrefix(in.Content, "/uploads/") && !strings.HasPrefix(in.Content, "https://") {
		client.Send <- marshalOut(ws.OutMsg{Type: "error", Payload: "图片地址无效"})
		return
	}

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
	h.hub.SendToUser(in.To, out)
	h.hub.SendToUser(client.UserID, out)
}

func (h *WSHandler) handleRead(client *ws.Client, in ws.InMsg) {
	friendID := in.FriendID
	if friendID == 0 {
		return
	}

	ok, err := h.friendRepo.AreFriends(client.UserID, friendID)
	if err != nil || !ok {
		return
	}

	if err := h.msgRepo.MarkRead(friendID, client.UserID); err != nil {
		return
	}
	h.hub.SendToUser(friendID, ws.OutMsg{
		Type:    "unread_update",
		Payload: map[string]interface{}{"friend_id": client.UserID, "count": 0},
	})
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

func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}
