package domain

import "time"

type PostTreeNodeType string

const (
	PostTreeNodeTypeFolder PostTreeNodeType = "folder"
	PostTreeNodeTypePost   PostTreeNodeType = "post"
)

type PostTreeNode struct {
	ID        int64            `gorm:"primaryKey"`
	UserID    int64            `gorm:"not null;index:idx_post_tree_nodes_user_parent,priority:1"`
	ParentID  *int64           `gorm:"index:idx_post_tree_nodes_user_parent,priority:2"`
	NodeType  PostTreeNodeType `gorm:"size:20;not null"`
	Name      string           `gorm:"size:120;not null;default:''"`
	PostID    *int64           `gorm:"index"`
	SortOrder int              `gorm:"not null;default:0;index:idx_post_tree_nodes_user_parent,priority:3"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Post *Post `gorm:"foreignKey:PostID"`
}

func (PostTreeNode) TableName() string { return "post_tree_nodes" }
