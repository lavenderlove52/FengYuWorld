package domain

import "time"

type Tag struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"not null;uniqueIndex:tags_unique_user_name,priority:1"`
	Name      string    `gorm:"size:80;not null;uniqueIndex:tags_unique_user_name,priority:2"`
	CreatedAt time.Time
}

func (Tag) TableName() string { return "tags" }

type PostTag struct {
	PostID int64 `gorm:"primaryKey"`
	TagID  int64 `gorm:"primaryKey"`
}

func (PostTag) TableName() string { return "post_tags" }
