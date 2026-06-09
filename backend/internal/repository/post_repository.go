package repository

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/util"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(tx *gorm.DB, post *domain.Post) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(post).Error
}

func (r *PostRepository) Update(tx *gorm.DB, post *domain.Post) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Save(post).Error
}

func (r *PostRepository) FindByID(id int64) (*domain.Post, error) {
	var post domain.Post
	err := r.db.Preload("Author").Preload("Category").Preload("Tags").
		First(&post, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &post, err
}

func (r *PostRepository) SoftDelete(id, authorID int64) error {
	return r.db.Model(&domain.Post{}).
		Where("id = ? AND author_id = ?", id, authorID).
		Update("is_deleted", true).Error
}

func (r *PostRepository) ListPublic(page util.PageParams) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64
	q := r.db.Model(&domain.Post{}).
		Where("status = ? AND visibility = ? AND is_deleted = ?", domain.PostStatusPublished, domain.PostVisibilityPublic, false)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Author").Preload("Tags").
		Order("published_at DESC NULLS LAST, id DESC").
		Offset(util.Offset(page)).Limit(page.PageSize).
		Find(&posts).Error
	return posts, total, err
}

func (r *PostRepository) SearchPublic(keyword string, page util.PageParams) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64
	q := r.db.Model(&domain.Post{}).
		Where("status = ? AND visibility = ? AND is_deleted = ?", domain.PostStatusPublished, domain.PostVisibilityPublic, false)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title ILIKE ? OR summary ILIKE ?", like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Author").Preload("Tags").
		Order("published_at DESC NULLS LAST, id DESC").
		Offset(util.Offset(page)).Limit(page.PageSize).
		Find(&posts).Error
	return posts, total, err
}

func (r *PostRepository) SearchLinkable(viewerID int64, keyword string, page util.PageParams) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64
	q := r.db.Model(&domain.Post{}).
		Joins("JOIN users ON users.id = posts.author_id").
		Where("posts.is_deleted = ?", false).
		Where(`(
			posts.author_id = ? OR
			(users.status = ? AND posts.status = ? AND (
				posts.visibility = ? OR
				(posts.visibility = ? AND EXISTS (
					SELECT 1 FROM friendships
					WHERE status = ?
					AND ((user_id = posts.author_id AND friend_id = ?) OR (user_id = ? AND friend_id = posts.author_id))
				)) OR
				(posts.visibility = ? AND EXISTS (
					SELECT 1 FROM post_visible_users
					WHERE post_id = posts.id AND user_id = ?
				))
			))
		)`,
			viewerID,
			domain.UserStatusActive,
			domain.PostStatusPublished,
			domain.PostVisibilityPublic,
			domain.PostVisibilityFriends, string(domain.FriendshipStatusAccepted), viewerID, viewerID,
			domain.PostVisibilitySpecified, viewerID,
		)
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("posts.title ILIKE ? OR posts.summary ILIKE ? OR users.username ILIKE ? OR users.nickname ILIKE ?", like, like, like, like)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Author").Preload("Tags").
		Order("posts.updated_at DESC, posts.id DESC").
		Offset(util.Offset(page)).Limit(page.PageSize).
		Find(&posts).Error
	return posts, total, err
}

func (r *PostRepository) ListByAuthor(authorID int64, status, visibility string, page util.PageParams) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64
	q := r.db.Model(&domain.Post{}).Where("author_id = ? AND is_deleted = ?", authorID, false)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if visibility != "" {
		q = q.Where("visibility = ?", visibility)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Author").Preload("Tags").
		Order("updated_at DESC").
		Offset(util.Offset(page)).Limit(page.PageSize).
		Find(&posts).Error
	return posts, total, err
}

func (r *PostRepository) ListPublicByUsername(username string, page util.PageParams) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64
	q := r.db.Model(&domain.Post{}).
		Joins("JOIN users ON users.id = posts.author_id").
		Where("users.username = ?", username).
		Where("posts.status = ? AND posts.visibility = ? AND posts.is_deleted = ?",
			domain.PostStatusPublished, domain.PostVisibilityPublic, false)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Author").Preload("Tags").
		Order("posts.published_at DESC NULLS LAST, posts.id DESC").
		Offset(util.Offset(page)).Limit(page.PageSize).
		Find(&posts).Error
	return posts, total, err
}

// ListByUsernameWithViewer 返回指定用户的文章，根据 viewerID 过滤可见性。
// viewerID == nil：只返回 public；viewerID != nil：返回 public + 好友可见（如果是好友）+ 指定可见（如果在列表中）+ 作者自己的所有 published。
func (r *PostRepository) ListByUsernameWithViewer(username string, viewerID *int64, page util.PageParams) ([]domain.Post, int64, error) {
	var posts []domain.Post
	var total int64
	q := r.db.Model(&domain.Post{}).
		Joins("JOIN users ON users.id = posts.author_id").
		Where("users.username = ?", username).
		Where("posts.status = ? AND posts.is_deleted = ?", domain.PostStatusPublished, false)

	if viewerID != nil {
		// 作者本人可看自己所有文章；其他人只能看 active 账号的文章
		q = q.Where(`(
			posts.author_id = ? OR
			(users.status = ? AND (
				posts.visibility = ? OR
				(posts.visibility = ? AND EXISTS (
					SELECT 1 FROM friendships
					WHERE status = ?
					AND ((user_id = posts.author_id AND friend_id = ?) OR (user_id = ? AND friend_id = posts.author_id))
				)) OR
				(posts.visibility = ? AND EXISTS (
					SELECT 1 FROM post_visible_users
					WHERE post_id = posts.id AND user_id = ?
				))
			))
		)`,
			*viewerID,
			domain.UserStatusActive,
			domain.PostVisibilityPublic,
			domain.PostVisibilityFriends, string(domain.FriendshipStatusAccepted), *viewerID, *viewerID,
			domain.PostVisibilitySpecified, *viewerID,
		)
	} else {
		q = q.Where("posts.visibility = ? AND users.status = ?", domain.PostVisibilityPublic, domain.UserStatusActive)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Preload("Author").Preload("Tags").
		Order("posts.published_at DESC NULLS LAST, posts.id DESC").
		Offset(util.Offset(page)).Limit(page.PageSize).
		Find(&posts).Error
	return posts, total, err
}

func (r *PostRepository) IncrementViewCount(id int64) error {
	return r.db.Model(&domain.Post{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (r *PostRepository) ReplaceVisibleUsers(tx *gorm.DB, postID int64, userIDs []int64) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	if err := db.Where("post_id = ?", postID).Delete(&domain.PostVisibleUser{}).Error; err != nil {
		return err
	}
	for _, uid := range userIDs {
		if err := db.Create(&domain.PostVisibleUser{PostID: postID, UserID: uid}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *PostRepository) GetVisibleUserIDs(postID int64) ([]int64, error) {
	var rows []domain.PostVisibleUser
	if err := r.db.Where("post_id = ?", postID).Find(&rows).Error; err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	return ids, nil
}

func (r *PostRepository) IsVisibleToUser(postID, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.PostVisibleUser{}).
		Where("post_id = ? AND user_id = ?", postID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *PostRepository) FindBySlug(slug string) (*domain.Post, error) {
	var post domain.Post
	err := r.db.Preload("Author").Preload("Category").Preload("Tags").
		Where("slug = ? AND is_deleted = ?", slug, false).
		First(&post).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &post, err
}

func (r *PostRepository) SlugExists(slug string, excludeID int64) (bool, error) {
	var count int64
	q := r.db.Model(&domain.Post{}).Where("slug = ? AND is_deleted = ?", slug, false)
	if excludeID > 0 {
		q = q.Where("id != ?", excludeID)
	}
	err := q.Count(&count).Error
	return count > 0, err
}

func (r *PostRepository) DB() *gorm.DB {
	return r.db
}
