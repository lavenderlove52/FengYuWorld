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

func (r *PostRepository) DB() *gorm.DB {
	return r.db
}
