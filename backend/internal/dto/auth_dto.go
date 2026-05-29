package dto

import "time"

type RegisterRequest struct {
	Username      string `json:"username" binding:"required,min=3,max=50"`
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=6,max=72"`
	Nickname      string `json:"nickname" binding:"required,min=1,max=80"`
	EmailCode     string `json:"email_code" binding:"required"`
	InviteCode    string `json:"invite_code" binding:"required"`
	CaptchaID     string `json:"captcha_id" binding:"required"`
	CaptchaAnswer string `json:"captcha_answer" binding:"required"`
}

type LoginRequest struct {
	Login         string `json:"login" binding:"required"`
	Password      string `json:"password" binding:"required"`
	CaptchaID     string `json:"captcha_id" binding:"required"`
	CaptchaAnswer string `json:"captcha_answer" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email,omitempty"`
	Nickname    string    `json:"nickname"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio,omitempty"`
	Role        string    `json:"role,omitempty"`
	Status      string    `json:"status,omitempty"`
	CanSitePin  bool      `json:"can_site_pin,omitempty"`
	CanCustomJS bool      `json:"can_custom_js,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

// PublicUserInfo 仅含公开字段，用于公开接口（公开主页、文章作者等）
type PublicUserInfo struct {
	ID        int64  `json:"id,omitempty"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio,omitempty"`
}

type UpdateProfileRequest struct {
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
	Bio       *string `json:"bio"`
}
