package domain

import "time"

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeEmoji MessageType = "emoji"
)

type Message struct {
	ID        int64       `gorm:"primaryKey"                    json:"id"`
	FromUser  int64       `gorm:"not null;index"                json:"from_user"`
	ToUser    int64       `gorm:"not null;index"                json:"to_user"`
	Content   string      `gorm:"type:text;not null"            json:"content"`
	Type      MessageType `gorm:"size:20;not null;default:text" json:"type"`
	IsRead    bool        `gorm:"not null;default:false"        json:"is_read"`
	CreatedAt time.Time   `gorm:"not null;default:now()"        json:"created_at"`
}

func (Message) TableName() string { return "messages" }
