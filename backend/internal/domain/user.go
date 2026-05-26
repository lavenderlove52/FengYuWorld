package domain

import "time"

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "active"
	UserStatusLocked UserStatus = "locked"
)

type User struct {
	ID           int64      `gorm:"primaryKey"`
	Username     string     `gorm:"size:50;uniqueIndex;not null"`
	Email        string     `gorm:"size:120;uniqueIndex;not null"`
	PasswordHash string     `gorm:"size:255;not null"`
	Nickname     string     `gorm:"size:80;not null"`
	AvatarURL    string     `gorm:"size:500;not null;default:''"`
	Bio          string     `gorm:"type:text;not null;default:''"`
	Role         UserRole   `gorm:"size:20;not null;default:user"`
	Status       UserStatus `gorm:"size:20;not null;default:active;index"`
	CanSitePin   bool       `gorm:"not null;default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string { return "users" }
