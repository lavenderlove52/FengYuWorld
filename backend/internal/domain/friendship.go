package domain

import "time"

type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "pending"
	FriendshipStatusAccepted FriendshipStatus = "accepted"
	FriendshipStatusRejected FriendshipStatus = "rejected"
	FriendshipStatusBlocked  FriendshipStatus = "blocked"
)

type Friendship struct {
	ID        int64            `gorm:"primaryKey"`
	UserID    int64            `gorm:"not null;index:idx_friendships_user_status,priority:1"`
	FriendID  int64            `gorm:"not null;index:idx_friendships_friend_status,priority:1"`
	Status    FriendshipStatus `gorm:"size:20;not null;default:pending"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Friendship) TableName() string { return "friendships" }
