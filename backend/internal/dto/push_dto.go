package dto

type PushSettingResponse struct {
	BarkEndpoint      string `json:"bark_endpoint"`
	NtfyBaseURL       string `json:"ntfy_base_url"`
	NtfyTopic         string `json:"ntfy_topic"`
	NtfyPublicEnabled bool   `json:"ntfy_public_enabled"`
	NtfyPublicToken   string `json:"ntfy_public_token,omitempty"`
}

type UpdateNtfySettingRequest struct {
	TopicSuffix string `json:"topic_suffix" binding:"required,min=1,max=64"`
}
type UpdateNtfyPublicRequest struct {
	Enabled bool `json:"enabled"`
}
type SendNtfyPushRequest struct {
	Mode     string `json:"mode" binding:"required,oneof=message title urgent priority tags"`
	Title    string `json:"title" binding:"max=200"`
	Message  string `json:"message" binding:"required,max=2000"`
	Priority int    `json:"priority" binding:"omitempty,min=1,max=5"`
	Tags     string `json:"tags" binding:"max=200"`
}
type PublicNtfyPushResponse struct {
	Topic string `json:"topic"`
}

type UpdatePushSettingRequest struct {
	BarkEndpoint string `json:"bark_endpoint" binding:"required,max=1000"`
}

type SendBarkPushRequest struct {
	Mode    string `json:"mode" binding:"required,oneof=message title sound call critical"`
	Title   string `json:"title" binding:"max=200"`
	Message string `json:"message" binding:"required,max=2000"`
}
