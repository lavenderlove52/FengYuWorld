package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
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
	ErrInvalidSlug     = errors.New("invalid slug")
)

var slugValidRe = regexp.MustCompile(`^[a-z0-9-]+$`)

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

	baseSlug, err := s.prepareBaseSlug(req.Slug, req.Title)
	if err != nil {
		return nil, err
	}
	slug, needPostID, err := s.resolveUniqueSlug(baseSlug, 0)
	if err != nil {
		return nil, err
	}

	allowComment := true
	if req.AllowComment != nil {
		allowComment = *req.AllowComment
	}
	showTOC := true
	if req.ShowTOC != nil {
		showTOC = *req.ShowTOC
	}

	post := &domain.Post{
		AuthorID:      authorID,
		Title:         req.Title,
		Slug:          slug,
		CoverImageURL: req.CoverImageURL,
		Summary:       strings.TrimSpace(req.Summary),
		ContentMD:     req.ContentMD,
		ContentHTML:   html,
		CategoryID:    req.CategoryID,
		Status:        domain.PostStatus(req.Status),
		Visibility:    domain.PostVisibility(req.Visibility),
		AllowComment:  allowComment,
		ShowTOC:       showTOC,
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
	// 所有候选 slug 均被占用时，用 slug-{postID} 作为最终回退
	if needPostID {
		post.Slug = fmt.Sprintf("%s-%d", baseSlug, post.ID)
		if err := tx.Model(&domain.Post{}).Where("id = ?", post.ID).Update("slug", post.Slug).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
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

	baseSlug, err := s.prepareBaseSlug(req.Slug, req.Title)
	if err != nil {
		return nil, err
	}
	slug, needPostID, err := s.resolveUniqueSlug(baseSlug, post.ID)
	if err != nil {
		return nil, err
	}
	if needPostID {
		slug = fmt.Sprintf("%s-%d", baseSlug, post.ID)
	}

	allowComment := post.AllowComment
	if req.AllowComment != nil {
		allowComment = *req.AllowComment
	}
	showTOC := post.ShowTOC
	if req.ShowTOC != nil {
		showTOC = *req.ShowTOC
	}

	wasPublished := post.Status == domain.PostStatusPublished
	post.Title = req.Title
	post.Slug = slug
	post.CoverImageURL = req.CoverImageURL
	post.Summary = strings.TrimSpace(req.Summary)
	post.ContentMD = req.ContentMD
	post.ContentHTML = html
	post.CategoryID = req.CategoryID
	post.Status = domain.PostStatus(req.Status)
	post.Visibility = domain.PostVisibility(req.Visibility)
	post.AllowComment = allowComment
	post.ShowTOC = showTOC

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

func (s *PostService) ListPublic(keyword string, page, pageSize int) (*util.PageResult[dto.PostSummary], error) {
	p := util.NormalizePageParams(page, pageSize)
	posts, total, err := s.posts.SearchPublic(strings.TrimSpace(keyword), p)
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

func (s *PostService) SearchLinkable(userID int64, keyword string, page, pageSize int) (*util.PageResult[dto.PostSummary], error) {
	p := util.NormalizePageParams(page, pageSize)
	posts, total, err := s.posts.SearchLinkable(userID, strings.TrimSpace(keyword), p)
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

func (s *PostService) ListByUsername(viewerID *int64, username string, page, pageSize int) (*util.PageResult[dto.PostSummary], error) {
	p := util.NormalizePageParams(page, pageSize)
	posts, total, err := s.posts.ListByUsernameWithViewer(username, viewerID, p)
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

func (s *PostService) toDetail(post *domain.Post, viewerID int64, isAuthor bool) *dto.PostDetail {
	summary := s.toSummary(post)
	detail := &dto.PostDetail{
		PostSummary:  summary,
		ContentMD:    post.ContentMD,
		ContentHTML:  post.ContentHTML,
		CategoryID:   post.CategoryID,
		AllowComment: post.AllowComment,
	}
	if isAuthor && post.Visibility == domain.PostVisibilitySpecified {
		ids, _ := s.posts.GetVisibleUserIDs(post.ID)
		detail.VisibleUserIDs = ids
	}
	return detail
}

func deref(id *int64) int64 {
	if id == nil {
		return 0
	}
	return *id
}

// GetBySlug 通过 slug 查询文章（校验可见性）
func (s *PostService) GetBySlug(slug string, viewerID *int64) (*dto.PostDetail, error) {
	post, err := s.posts.FindBySlug(slug)
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
		_ = s.posts.IncrementViewCount(post.ID)
		post.ViewCount++
	}
	return s.toDetail(post, deref(viewerID), isAuthor), nil
}

// prepareBaseSlug 解析用户填写或标题生成的 slug；用户填写纯数字则拒绝。
func (s *PostService) prepareBaseSlug(reqSlug, title string) (string, error) {
	baseSlug := strings.TrimSpace(reqSlug)
	userProvided := baseSlug != ""
	if !userProvided {
		baseSlug = util.Slugify(title)
	}
	if baseSlug == "" {
		return "", ErrInvalidPostData
	}
	if !slugValidRe.MatchString(baseSlug) {
		return "", ErrInvalidSlug
	}
	if userProvided && util.SlugIsAllNumeric(baseSlug) {
		return "", ErrInvalidSlug
	}
	return baseSlug, nil
}

// resolveUniqueSlug 找到 base 的第一个可用候选 slug（排除 excludeID 自身）。
// 依次尝试 base、base-2 … base-10；若全部被占用则返回 (base, true, nil)，
// 调用方应在知道 postID 后拼接 base-{postID}。
func (s *PostService) resolveUniqueSlug(base string, excludeID int64) (string, bool, error) {
	if base == "" {
		return "", false, nil
	}
	exists, err := s.posts.SlugExists(base, excludeID)
	if err != nil {
		return "", false, err
	}
	if !exists {
		return base, false, nil
	}
	for i := 2; i <= 10; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		exists, err = s.posts.SlugExists(candidate, excludeID)
		if err != nil {
			return "", false, err
		}
		if !exists {
			return candidate, false, nil
		}
	}
	return base, true, nil
}
