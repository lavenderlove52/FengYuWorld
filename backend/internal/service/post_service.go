package service

import (
	"errors"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/markdown"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/util"
)

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrPostForbidden   = errors.New("post forbidden")
	ErrInvalidPostData = errors.New("invalid post data")
)

type PostService struct {
	posts       *repository.PostRepository
	tags        *repository.TagRepository
	categories  *repository.CategoryRepository
	friendships *repository.FriendshipRepository
	policy      *PostPolicy
	md          *markdown.Renderer
}

func NewPostService(
	posts *repository.PostRepository,
	tags *repository.TagRepository,
	categories *repository.CategoryRepository,
	friendships *repository.FriendshipRepository,
	policy *PostPolicy,
	md *markdown.Renderer,
) *PostService {
	return &PostService{
		posts:       posts,
		tags:        tags,
		categories:  categories,
		friendships: friendships,
		policy:      policy,
		md:          md,
	}
}

func (s *PostService) Create(authorID int64, req dto.CreatePostRequest) (*dto.PostDetail, error) {
	if err := s.validatePostInput(authorID, req.Visibility, req.VisibleUserIDs, req.CategoryID); err != nil {
		return nil, err
	}

	html, err := s.md.Render(req.ContentMD)
	if err != nil {
		return nil, err
	}

	slug := req.Slug
	if slug == "" {
		slug = util.Slugify(req.Title)
	}

	allowComment := true
	if req.AllowComment != nil {
		allowComment = *req.AllowComment
	}

	post := &domain.Post{
		AuthorID:      authorID,
		Title:         req.Title,
		Slug:          slug,
		CoverImageURL: req.CoverImageURL,
		ContentMD:     req.ContentMD,
		ContentHTML:   html,
		CategoryID:    req.CategoryID,
		Status:        domain.PostStatus(req.Status),
		Visibility:    domain.PostVisibility(req.Visibility),
		AllowComment:  allowComment,
	}
	if post.Status == domain.PostStatusPublished {
		now := time.Now()
		post.PublishedAt = &now
	}

	tx := s.posts.DB().Begin()
	if err := s.posts.Create(tx, post); err != nil {
		tx.Rollback()
		return nil, err
	}

	tags, err := s.tags.FindOrCreateByNames(tx, authorID, req.TagNames)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.tags.ReplacePostTags(tx, post.ID, tags); err != nil {
		tx.Rollback()
		return nil, err
	}

	if post.Visibility == domain.PostVisibilitySpecified {
		if err := s.posts.ReplaceVisibleUsers(tx, post.ID, req.VisibleUserIDs); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	created, err := s.posts.FindByID(post.ID)
	if err != nil || created == nil {
		return nil, ErrPostNotFound
	}
	return s.toDetail(created, authorID, true), nil
}

func (s *PostService) Update(authorID, postID int64, req dto.UpdatePostRequest) (*dto.PostDetail, error) {
	post, err := s.posts.FindByID(postID)
	if err != nil || post == nil || post.IsDeleted {
		return nil, ErrPostNotFound
	}
	if post.AuthorID != authorID {
		return nil, ErrPostForbidden
	}

	if err := s.validatePostInput(authorID, req.Visibility, req.VisibleUserIDs, req.CategoryID); err != nil {
		return nil, err
	}

	html, err := s.md.Render(req.ContentMD)
	if err != nil {
		return nil, err
	}

	slug := req.Slug
	if slug == "" {
		slug = util.Slugify(req.Title)
	}

	allowComment := post.AllowComment
	if req.AllowComment != nil {
		allowComment = *req.AllowComment
	}

	wasPublished := post.Status == domain.PostStatusPublished
	post.Title = req.Title
	post.Slug = slug
	post.CoverImageURL = req.CoverImageURL
	post.ContentMD = req.ContentMD
	post.ContentHTML = html
	post.CategoryID = req.CategoryID
	post.Status = domain.PostStatus(req.Status)
	post.Visibility = domain.PostVisibility(req.Visibility)
	post.AllowComment = allowComment

	if post.Status == domain.PostStatusPublished && (!wasPublished || post.PublishedAt == nil) {
		now := time.Now()
		post.PublishedAt = &now
	}

	tx := s.posts.DB().Begin()
	if err := s.posts.Update(tx, post); err != nil {
		tx.Rollback()
		return nil, err
	}

	tags, err := s.tags.FindOrCreateByNames(tx, authorID, req.TagNames)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.tags.ReplacePostTags(tx, post.ID, tags); err != nil {
		tx.Rollback()
		return nil, err
	}

	if post.Visibility == domain.PostVisibilitySpecified {
		if err := s.posts.ReplaceVisibleUsers(tx, post.ID, req.VisibleUserIDs); err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		if err := s.posts.ReplaceVisibleUsers(tx, post.ID, nil); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	updated, err := s.posts.FindByID(post.ID)
	if err != nil || updated == nil {
		return nil, ErrPostNotFound
	}
	return s.toDetail(updated, authorID, true), nil
}

func (s *PostService) Delete(authorID, postID int64) error {
	post, err := s.posts.FindByID(postID)
	if err != nil || post == nil || post.IsDeleted {
		return ErrPostNotFound
	}
	if post.AuthorID != authorID {
		return ErrPostForbidden
	}
	return s.posts.SoftDelete(postID, authorID)
}

func (s *PostService) GetDetail(viewerID *int64, postID int64) (*dto.PostDetail, error) {
	post, err := s.posts.FindByID(postID)
	if err != nil || post == nil {
		return nil, ErrPostNotFound
	}

	canRead, err := s.policy.CanReadPost(viewerID, post)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrPostNotFound
	}

	isAuthor := viewerID != nil && *viewerID == post.AuthorID
	if post.Status == domain.PostStatusPublished && !post.IsDeleted {
		_ = s.posts.IncrementViewCount(postID)
		post.ViewCount++
	}

	return s.toDetail(post, deref(viewerID), isAuthor), nil
}

func (s *PostService) ListPublic(page, pageSize int) (*util.PageResult[dto.PostSummary], error) {
	p := util.NormalizePageParams(page, pageSize)
	posts, total, err := s.posts.ListPublic(p)
	if err != nil {
		return nil, err
	}
	items := make([]dto.PostSummary, 0, len(posts))
	for i := range posts {
		items = append(items, s.toSummary(&posts[i]))
	}
	return &util.PageResult[dto.PostSummary]{
		Items: items, Total: total, Page: p.Page, PageSize: p.PageSize,
	}, nil
}

func (s *PostService) ListMine(userID int64, status, visibility string, page, pageSize int) (*util.PageResult[dto.PostSummary], error) {
	p := util.NormalizePageParams(page, pageSize)
	posts, total, err := s.posts.ListByAuthor(userID, status, visibility, p)
	if err != nil {
		return nil, err
	}
	items := make([]dto.PostSummary, 0, len(posts))
	for i := range posts {
		items = append(items, s.toSummary(&posts[i]))
	}
	return &util.PageResult[dto.PostSummary]{
		Items: items, Total: total, Page: p.Page, PageSize: p.PageSize,
	}, nil
}

func (s *PostService) ListByUsername(username string, page, pageSize int) (*util.PageResult[dto.PostSummary], error) {
	p := util.NormalizePageParams(page, pageSize)
	posts, total, err := s.posts.ListPublicByUsername(username, p)
	if err != nil {
		return nil, err
	}
	items := make([]dto.PostSummary, 0, len(posts))
	for i := range posts {
		items = append(items, s.toSummary(&posts[i]))
	}
	return &util.PageResult[dto.PostSummary]{
		Items: items, Total: total, Page: p.Page, PageSize: p.PageSize,
	}, nil
}

func (s *PostService) validatePostInput(authorID int64, visibility string, visibleUserIDs []int64, categoryID *int64) error {
	if visibility == string(domain.PostVisibilitySpecified) {
		if len(visibleUserIDs) == 0 {
			return ErrInvalidPostData
		}
		friendIDs, err := s.friendships.ListAcceptedFriendIDs(authorID)
		if err != nil {
			return err
		}
		friendSet := map[int64]struct{}{}
		for _, id := range friendIDs {
			friendSet[id] = struct{}{}
		}
		for _, uid := range visibleUserIDs {
			if _, ok := friendSet[uid]; !ok {
				return ErrInvalidPostData
			}
		}
	}
	if categoryID != nil {
		cat, err := s.categories.FindByIDAndUser(*categoryID, authorID)
		if err != nil || cat == nil {
			return ErrInvalidPostData
		}
	}
	return nil
}

func (s *PostService) toSummary(post *domain.Post) dto.PostSummary {
	summary := dto.PostSummary{
		ID:            post.ID,
		Title:         post.Title,
		Slug:          post.Slug,
		CoverImageURL: post.CoverImageURL,
		Summary:       markdown.Summary(post.ContentHTML, 160),
		Status:        string(post.Status),
		Visibility:    string(post.Visibility),
		ViewCount:     post.ViewCount,
		CreatedAt:     util.FormatTime(post.CreatedAt),
	}
	if post.PublishedAt != nil {
		t := util.FormatTime(*post.PublishedAt)
		summary.PublishedAt = &t
	}
	if post.Author != nil {
		summary.Author = toUserInfo(post.Author, false)
	}
	if len(post.Tags) > 0 {
		summary.Tags = make([]string, 0, len(post.Tags))
		for _, tag := range post.Tags {
			summary.Tags = append(summary.Tags, tag.Name)
		}
	}
	return summary
}

func (s *PostService) toDetail(post *domain.Post, viewerID int64, isAuthor bool) *dto.PostDetail {
	summary := s.toSummary(post)
	detail := &dto.PostDetail{
		PostSummary:  summary,
		ContentHTML:  post.ContentHTML,
		CategoryID:   post.CategoryID,
		AllowComment: post.AllowComment,
	}
	if isAuthor {
		detail.ContentMD = post.ContentMD
		if post.Visibility == domain.PostVisibilitySpecified {
			ids, _ := s.posts.GetVisibleUserIDs(post.ID)
			detail.VisibleUserIDs = ids
		}
	}
	return detail
}

func deref(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}
