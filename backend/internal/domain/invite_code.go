package domain

import "time"

type InviteCode struct {
	ID        int64      `gorm:"primaryKey"                    json:"id"`
	Code      string     `gorm:"size:20;not null;uniqueIndex"  json:"code"`
	CreatedBy int64      `gorm:"not null"                      json:"created_by"`
	IsUsed    bool       `gorm:"not null;default:false"        json:"is_used"`
	UsedBy    *int64     `                                     json:"used_by,omitempty"`
	UsedAt    *time.Time `                                     json:"used_at,omitempty"`
	ExpiresAt *time.Time `                                     json:"expires_at,omitempty"`
	CreatedAt time.Time  `gorm:"not null;default:now()"        json:"created_at"`

	Creator    *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	UsedByUser *User `gorm:"foreignKey:UsedBy"    json:"used_by_user,omitempty"`
}

func (InviteCode) TableName() string { return "invite_codes" }
