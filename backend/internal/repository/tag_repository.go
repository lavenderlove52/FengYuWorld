package repository

import (
	"errors"
	"strings"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type TagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) ListByUser(userID int64) ([]domain.Tag, error) {
	var list []domain.Tag
	err := r.db.Where("user_id = ?", userID).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *TagRepository) FindByIDAndUser(id, userID int64) (*domain.Tag, error) {
	var t domain.Tag
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

func (r *TagRepository) FindOrCreateByNames(tx *gorm.DB, userID int64, names []string) ([]domain.Tag, error) {
	db := r.db
	if tx != nil {
		db = tx
	}
	tags := make([]domain.Tag, 0, len(names))
	seen := map[string]struct{}{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		var tag domain.Tag
		err := db.Where("user_id = ? AND name = ?", userID, name).First(&tag).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tag = domain.Tag{UserID: userID, Name: name}
			if err := db.Create(&tag).Error; err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *TagRepository) ReplacePostTags(tx *gorm.DB, postID int64, tags []domain.Tag) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	if err := db.Where("post_id = ?", postID).Delete(&domain.PostTag{}).Error; err != nil {
		return err
	}
	for _, tag := range tags {
		if err := db.Create(&domain.PostTag{PostID: postID, TagID: tag.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *TagRepository) Create(t *domain.Tag) error {
	return r.db.Create(t).Error
}

func (r *TagRepository) Delete(id, userID int64) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&domain.Tag{}).Error
}
