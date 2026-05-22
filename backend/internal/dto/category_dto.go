package dto

type CategoryRequest struct {
	Name      string `json:"name" binding:"required,max=80"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

type CategoryResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	SortOrder int    `json:"sort_order"`
}

type TagRequest struct {
	Name string `json:"name" binding:"required,max=80"`
}

type TagResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
