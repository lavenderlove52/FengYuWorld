package domain

import "time"

type SitePin struct {
	ID           int64     `gorm:"primaryKey"`
	PostID       int64     `gorm:"not null;uniqueIndex"`
	DisplayOrder int       `gorm:"not null;default:0"`
	PinnedBy     int64     `gorm:"not null"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`

	Post         *Post `gorm:"foreignKey:PostID"`
	PinnedByUser *User `gorm:"foreignKey:PinnedBy"`
}

func (SitePin) TableName() string { return "site_pins" }

type UserPin struct {
	ID           int64     `gorm:"primaryKey"`
	UserID       int64     `gorm:"not null;index:idx_user_pins_user_post,priority:1"`
	PostID       int64     `gorm:"not null;index:idx_user_pins_user_post,priority:2"`
	DisplayOrder int       `gorm:"not null;default:0"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`

	Post *Post `gorm:"foreignKey:PostID"`
}

func (UserPin) TableName() string { return "user_pins" }
