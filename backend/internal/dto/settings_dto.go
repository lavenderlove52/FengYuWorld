package dto

type PublicSettingsResponse struct {
	FriendChatEnabled bool `json:"friend_chat_enabled"`
}

type UpdateFriendChatSettingRequest struct {
	Enabled bool `json:"enabled"`
}
