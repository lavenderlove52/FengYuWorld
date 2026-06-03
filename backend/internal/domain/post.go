package domain

import "time"

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

type PostVisibility string

const (
	PostVisibilityPublic    PostVisibility = "public"
	PostVisibilityPrivate   PostVisibility = "private"
	PostVisibilityFriends   PostVisibility = "friends"
	PostVisibilitySpecified PostVisibility = "specified"
)

type Post struct {
	ID            int64          `gorm:"primaryKey"`
	AuthorID      int64          `gorm:"not null;index:idx_posts_author_status_deleted,priority:1"`
	Title         string         `gorm:"size:200;not null"`
	Slug          string         `gorm:"size:220;not null;default:''"`
	CoverImageURL string         `gorm:"size:500;not null;default:''"`
	Summary       string         `gorm:"type:text;not null;default:''"`
	ContentMD     string         `gorm:"type:text;not null"`
	ContentHTML   string         `gorm:"type:text;not null"`
	CategoryID    *int64         `gorm:"index"`
	Status        PostStatus     `gorm:"size:20;not null;default:draft;index:idx_posts_author_status_deleted,priority:2;index:idx_posts_public_list,priority:1"`
	Visibility    PostVisibility `gorm:"size:20;not null;default:private;index:idx_posts_public_list,priority:2"`
	AllowComment  bool           `gorm:"not null;default:true"`
	ShowTOC       bool           `gorm:"not null;default:true"`
	ViewCount     int64          `gorm:"not null;default:0"`
	IsDeleted     bool           `gorm:"not null;default:false;index:idx_posts_author_status_deleted,priority:3;index:idx_posts_public_list,priority:3"`
	PublishedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Author   *User     `gorm:"foreignKey:AuthorID"`
	Category *Category `gorm:"foreignKey:CategoryID"`
	Tags     []Tag     `gorm:"many2many:post_tags;"`
}

func (Post) TableName() string { return "posts" }

type PostVisibleUser struct {
	PostID    int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"primaryKey"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (PostVisibleUser) TableName() string { return "post_visible_users" }
