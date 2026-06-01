package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// InMsg 客户端发来的消息
type InMsg struct {
	Type     string `json:"type"` // "msg" | "read"
	To       int64  `json:"to"`
	Content  string `json:"content"`
	MsgType  string `json:"msg_type"`  // "text"|"image"|"emoji"
	FriendID int64  `json:"friend_id"` // 用于 read 类型
}

// OutMsg 服务端推送的消息
type OutMsg struct {
	Type    string      `json:"type"` // "msg" | "unread_update" | "error"
	Payload interface{} `json:"payload"`
}

// Client 代表一个 WebSocket 连接
type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub 管理所有在线连接（支持同一用户多连接）
type Hub struct {
	clients map[int64][]*Client
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int64][]*Client),
	}
}

// Register 注册新连接
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.UserID] = append(h.clients[c.UserID], c)
}

// Unregister 注销连接
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	list := h.clients[c.UserID]
	newList := list[:0]
	for _, cl := range list {
		if cl != c {
			newList = append(newList, cl)
		}
	}
	if len(newList) == 0 {
		delete(h.clients, c.UserID)
	} else {
		h.clients[c.UserID] = newList
	}
	close(c.Send)
}

// SendToUser 向指定用户的所有连接推送消息，返回是否在线
func (h *Hub) SendToUser(userID int64, msg OutMsg) bool {
	h.mu.RLock()
	list := h.clients[userID]
	h.mu.RUnlock()

	if len(list) == 0 {
		return false
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return false
	}

	for _, c := range list {
		safeSend(c.Send, data)
	}
	return true
}

// safeSend 向 channel 发送数据，安全处理 channel 已关闭和缓冲区满的情况。
func safeSend(ch chan []byte, data []byte) {
	defer func() { recover() }() // 防止往已关闭 channel 发数据时 panic
	select {
	case ch <- data:
	default:
		// 发送缓冲区满，跳过该连接
	}
}
