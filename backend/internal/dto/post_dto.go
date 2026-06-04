package dto

type CreatePostRequest struct {
	Title          string   `json:"title" binding:"required,max=200"`
	Slug           string   `json:"slug"`
	CoverImageURL  string   `json:"cover_image_url"`
	Summary        string   `json:"summary" binding:"max=500"`
	ContentMD      string   `json:"content_md" binding:"required"`
	CategoryID     *int64   `json:"category_id"`
	TagNames       []string `json:"tag_names"`
	Status         string   `json:"status" binding:"required,oneof=draft published"`
	Visibility     string   `json:"visibility" binding:"required,oneof=public private friends specified"`
	VisibleUserIDs []int64  `json:"visible_user_ids"`
	AllowComment   *bool    `json:"allow_comment"`
	ShowTOC        *bool    `json:"show_toc"`
}

type UpdatePostRequest struct {
	Title          string   `json:"title" binding:"required,max=200"`
	Slug           string   `json:"slug"`
	CoverImageURL  string   `json:"cover_image_url"`
	Summary        string   `json:"summary" binding:"max=500"`
	ContentMD      string   `json:"content_md" binding:"required"`
	CategoryID     *int64   `json:"category_id"`
	TagNames       []string `json:"tag_names"`
	Status         string   `json:"status" binding:"required,oneof=draft published"`
	Visibility     string   `json:"visibility" binding:"required,oneof=public private friends specified"`
	VisibleUserIDs []int64  `json:"visible_user_ids"`
	AllowComment   *bool    `json:"allow_comment"`
	ShowTOC        *bool    `json:"show_toc"`
}

type PostSummary struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Slug          string         `json:"slug"`
	CoverImageURL string         `json:"cover_image_url"`
	Summary       string         `json:"summary"`
	Status        string         `json:"status"`
	Visibility    string         `json:"visibility"`
	ShowTOC       bool           `json:"show_toc"`
	ViewCount     int64          `json:"view_count"`
	PublishedAt   *string        `json:"published_at"`
	CreatedAt     string         `json:"created_at"`
	Author        PublicUserInfo `json:"author"`
	Tags          []string       `json:"tags,omitempty"`
}

type PostDetail struct {
	PostSummary
	ContentMD      string  `json:"content_md,omitempty"`
	ContentHTML    string  `json:"content_html"`
	CategoryID     *int64  `json:"category_id"`
	AllowComment   bool    `json:"allow_comment"`
	VisibleUserIDs []int64 `json:"visible_user_ids,omitempty"`
}

type UploadResponse struct {
	ReferenceID int64  `json:"reference_id"`
	URL         string `json:"url"`
	MobileURL   string `json:"mobile_url,omitempty"`
}
