package dto

type FriendRequestCreate struct {
	FriendUsername string `json:"friend_username" binding:"required"`
}

type FriendshipItem struct {
	ID        int64    `json:"id"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
	User      UserInfo `json:"user"`
}

type FriendListItem struct {
	UserInfo
	FriendsSince string `json:"friends_since"`
}
