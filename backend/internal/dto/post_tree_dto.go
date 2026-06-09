package dto

type PostTreeNodeResponse struct {
	ID        int64                   `json:"id"`
	ParentID  *int64                  `json:"parent_id"`
	NodeType  string                  `json:"node_type"`
	Name      string                  `json:"name"`
	PostID    *int64                  `json:"post_id,omitempty"`
	SortOrder int                     `json:"sort_order"`
	Post      *PostSummary            `json:"post,omitempty"`
	Children  []*PostTreeNodeResponse `json:"children,omitempty"`
}

type CreatePostTreeFolderRequest struct {
	ParentID  *int64 `json:"parent_id"`
	Name      string `json:"name" binding:"required,max=120"`
	SortOrder int    `json:"sort_order"`
}

type AddPostTreePostRequest struct {
	ParentID  *int64 `json:"parent_id"`
	PostID    int64  `json:"post_id" binding:"required"`
	SortOrder int    `json:"sort_order"`
}

type UpdatePostTreeNodeRequest struct {
	ParentID  *int64 `json:"parent_id"`
	Name      string `json:"name" binding:"max=120"`
	SortOrder int    `json:"sort_order"`
}
