package repository

import (
	"errors"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) ListByUser(userID int64) ([]domain.Category, error) {
	var list []domain.Category
	err := r.db.Where("user_id = ?", userID).Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

func (r *CategoryRepository) FindByIDAndUser(id, userID int64) (*domain.Category, error) {
	var c domain.Category
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *CategoryRepository) Create(c *domain.Category) error {
	return r.db.Create(c).Error
}

func (r *CategoryRepository) Update(c *domain.Category) error {
	return r.db.Save(c).Error
}

func (r *CategoryRepository) Delete(id, userID int64) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&domain.Category{}).Error
}
