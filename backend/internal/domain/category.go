package domain

import "time"

type Category struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"not null;uniqueIndex:categories_unique_user_name,priority:1"`
	Name      string    `gorm:"size:80;not null;uniqueIndex:categories_unique_user_name,priority:2"`
	Icon      string    `gorm:"size:200;not null;default:''"`
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Category) TableName() string { return "categories" }
