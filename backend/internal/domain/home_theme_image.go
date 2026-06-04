package domain

import "time"

type HomeThemeImage struct {
	ID             int64     `gorm:"primaryKey" json:"id"`
	UserID         int64     `gorm:"not null;index" json:"-"`
	UploadID       int64     `gorm:"not null;uniqueIndex" json:"-"`
	URL            string    `gorm:"size:500;not null" json:"url"`
	MobileUploadID *int64    `gorm:"uniqueIndex" json:"-"`
	MobileURL      string    `gorm:"size:500;not null;default:''" json:"mobile_url,omitempty"`
	CreatedAt      time.Time `gorm:"not null;default:now()" json:"created_at"`
}

func (HomeThemeImage) TableName() string { return "home_theme_images" }
