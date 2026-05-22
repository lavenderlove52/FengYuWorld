package domain

import "time"

type Upload struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"not null"`
	Path      string    `gorm:"size:500;not null"`
	URL       string    `gorm:"size:500;not null"`
	Mime      string    `gorm:"size:120;not null"`
	SizeBytes int64     `gorm:"not null"`
	CreatedAt time.Time
}

func (Upload) TableName() string { return "uploads" }
