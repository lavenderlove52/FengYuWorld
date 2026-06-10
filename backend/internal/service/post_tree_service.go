package service

import (
	"errors"
	"sort"
	"strings"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/util"
)

var (
	ErrPostTreeNodeNotFound = errors.New("post tree node not found")
	ErrInvalidPostTreeNode  = errors.New("invalid post tree node")
	ErrPostAlreadyInTree    = errors.New("post already in tree")
)

type PostTreeService struct {
	nodes *repository.PostTreeRepository
	posts *repository.PostRepository
}

func NewPostTreeService(nodes *repository.PostTreeRepository, posts *repository.PostRepository) *PostTreeService {
	return &PostTreeService{nodes: nodes, posts: posts}
}

func (s *PostTreeService) List(userID int64) ([]*dto.PostTreeNodeResponse, error) {
	nodes, err := s.nodes.ListByUser(userID)
	if err != nil {
		return nil, err
	}

	byID := make(map[int64]*dto.PostTreeNodeResponse, len(nodes))
	roots := make([]*dto.PostTreeNodeResponse, 0)
	for i := range nodes {
		item := s.toResponse(&nodes[i])
		byID[item.ID] = item
	}
	for _, item := range byID {
		if item.ParentID != nil {
			if parent, ok := byID[*item.ParentID]; ok {
				parent.Children = append(parent.Children, item)
				continue
			}
		}
		roots = append(roots, item)
	}
	sortTree(roots)
	return roots, nil
}

func (s *PostTreeService) CreateFolder(userID int64, req dto.CreatePostTreeFolderRequest) (*dto.PostTreeNodeResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrInvalidPostTreeNode
	}
	if err := s.ensureFolderParent(userID, req.ParentID); err != nil {
		return nil, err
	}
	if req.ParentID != nil {
		hasPostChildren, err := s.nodes.HasPostChildren(*req.ParentID, userID)
		if err != nil {
			return nil, err
		}
		if hasPostChildren {
			return nil, ErrInvalidPostTreeNode
		}
	}
	node := &domain.PostTreeNode{
		UserID:    userID,
		ParentID:  req.ParentID,
		NodeType:  domain.PostTreeNodeTypeFolder,
		Name:      name,
		SortOrder: req.SortOrder,
	}
	if err := s.nodes.Create(node); err != nil {
		return nil, err
	}
	return s.toResponse(node), nil
}

func (s *PostTreeService) AddPost(userID int64, req dto.AddPostTreePostRequest) (*dto.PostTreeNodeResponse, error) {
	if err := s.ensureFolderParent(userID, req.ParentID); err != nil {
		return nil, err
	}
	if req.ParentID != nil {
		if err := s.validateLeafFolder(userID, *req.ParentID); err != nil {
			return nil, err
		}
	}
	post, err := s.posts.FindByID(req.PostID)
	if err != nil || post == nil || post.IsDeleted || post.AuthorID != userID {
		return nil, ErrInvalidPostTreeNode
	}
	postID := post.ID
	node := &domain.PostTreeNode{
		UserID:    userID,
		ParentID:  req.ParentID,
		NodeType:  domain.PostTreeNodeTypePost,
		PostID:    &postID,
		SortOrder: req.SortOrder,
	}
	if err := s.nodes.Create(node); err != nil {
		if strings.Contains(err.Error(), "idx_post_tree_nodes_user_post") {
			return nil, ErrPostAlreadyInTree
		}
		return nil, err
	}
	node.Post = post
	return s.toResponse(node), nil
}

func (s *PostTreeService) Update(userID, id int64, req dto.UpdatePostTreeNodeRequest) (*dto.PostTreeNodeResponse, error) {
	node, err := s.nodes.FindByIDAndUser(id, userID)
	if err != nil || node == nil {
		return nil, ErrPostTreeNodeNotFound
	}
	if err := s.ensureFolderParent(userID, req.ParentID); err != nil {
		return nil, err
	}
	if node.NodeType == domain.PostTreeNodeTypeFolder {
		name := strings.TrimSpace(req.Name)
		if name == "" {
			return nil, ErrInvalidPostTreeNode
		}
		if req.ParentID != nil && *req.ParentID == node.ID {
			return nil, ErrInvalidPostTreeNode
		}
		if err := s.ensureNotMovingIntoDescendant(userID, node.ID, req.ParentID); err != nil {
			return nil, err
		}
		node.Name = name
	}
	node.ParentID = req.ParentID
	node.SortOrder = req.SortOrder
	if err := s.nodes.Update(node); err != nil {
		return nil, err
	}
	return s.toResponse(node), nil
}

func (s *PostTreeService) Delete(userID, id int64) error {
	node, err := s.nodes.FindByIDAndUser(id, userID)
	if err != nil || node == nil {
		return ErrPostTreeNodeNotFound
	}
	return s.nodes.Delete(id, userID)
}

func (s *PostTreeService) ensureFolderParent(userID int64, parentID *int64) error {
	if parentID == nil {
		return nil
	}
	parent, err := s.nodes.FindByIDAndUser(*parentID, userID)
	if err != nil || parent == nil {
		return ErrInvalidPostTreeNode
	}
	if parent.NodeType != domain.PostTreeNodeTypeFolder {
		return ErrInvalidPostTreeNode
	}
	return nil
}

func (s *PostTreeService) validateLeafFolder(userID, folderID int64) error {
	folder, err := s.nodes.FindByIDAndUser(folderID, userID)
	if err != nil || folder == nil || folder.NodeType != domain.PostTreeNodeTypeFolder {
		return ErrInvalidPostTreeNode
	}
	hasFolderChildren, err := s.nodes.HasFolderChildren(folderID, userID)
	if err != nil {
		return err
	}
	if hasFolderChildren {
		return ErrInvalidPostTreeNode
	}
	return nil
}

func (s *PostTreeService) ensureNotMovingIntoDescendant(userID, nodeID int64, parentID *int64) error {
	if parentID == nil {
		return nil
	}
	currentID := *parentID
	for {
		if currentID == nodeID {
			return ErrInvalidPostTreeNode
		}
		current, err := s.nodes.FindByIDAndUser(currentID, userID)
		if err != nil || current == nil || current.ParentID == nil {
			return err
		}
		currentID = *current.ParentID
	}
}

func (s *PostTreeService) toResponse(node *domain.PostTreeNode) *dto.PostTreeNodeResponse {
	resp := &dto.PostTreeNodeResponse{
		ID:        node.ID,
		ParentID:  node.ParentID,
		NodeType:  string(node.NodeType),
		Name:      node.Name,
		PostID:    node.PostID,
		SortOrder: node.SortOrder,
	}
	if node.Post != nil {
		resp.Name = node.Post.Title
		summary := s.postSummary(node.Post)
		resp.Post = &summary
	}
	return resp
}

func (s *PostTreeService) postSummary(post *domain.Post) dto.PostSummary {
	summary := dto.PostSummary{
		ID:            post.ID,
		Title:         post.Title,
		Slug:          post.Slug,
		CoverImageURL: post.CoverImageURL,
		Summary:       post.Summary,
		Status:        string(post.Status),
		Visibility:    string(post.Visibility),
		ShowTOC:       post.ShowTOC,
		ViewCount:     post.ViewCount,
		CreatedAt:     util.FormatTime(post.CreatedAt),
	}
	if post.PublishedAt != nil {
		t := util.FormatTime(*post.PublishedAt)
		summary.PublishedAt = &t
	}
	if post.Author != nil {
		summary.Author = toPublicUserInfo(post.Author)
	}
	if len(post.Tags) > 0 {
		summary.Tags = make([]string, 0, len(post.Tags))
		for _, tag := range post.Tags {
			summary.Tags = append(summary.Tags, tag.Name)
		}
	}
	return summary
}

func sortTree(nodes []*dto.PostTreeNodeResponse) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].SortOrder == nodes[j].SortOrder {
			return nodes[i].ID < nodes[j].ID
		}
		return nodes[i].SortOrder < nodes[j].SortOrder
	})
	for _, node := range nodes {
		sortTree(node.Children)
	}
}
