package repository

import (
	"errors"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrHomeThemeImageLimit = errors.New("home theme image limit reached")

type HomeThemeImageRepository struct {
	db *gorm.DB
}

func NewHomeThemeImageRepository(db *gorm.DB) *HomeThemeImageRepository {
	return &HomeThemeImageRepository{db: db}
}

func (r *HomeThemeImageRepository) CreateWithLimit(image *domain.HomeThemeImage, limit int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user domain.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&user, image.UserID).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&domain.HomeThemeImage{}).Where("user_id = ?", image.UserID).Count(&count).Error; err != nil {
			return err
		}
		if count >= limit {
			return ErrHomeThemeImageLimit
		}
		return tx.Create(image).Error
	})
}

func (r *HomeThemeImageRepository) CountByUserID(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&domain.HomeThemeImage{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *HomeThemeImageRepository) ListByUserID(userID int64) ([]domain.HomeThemeImage, error) {
	var images []domain.HomeThemeImage
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC, id DESC").Find(&images).Error
	return images, err
}

func (r *HomeThemeImageRepository) ListByUserIDAndIDs(userID int64, ids []int64) ([]domain.HomeThemeImage, error) {
	if len(ids) == 0 {
		return []domain.HomeThemeImage{}, nil
	}
	var images []domain.HomeThemeImage
	err := r.db.Where("user_id = ? AND id IN ?", userID, ids).Find(&images).Error
	return images, err
}

func (r *HomeThemeImageRepository) FindByIDAndUserID(id, userID int64) (*domain.HomeThemeImage, error) {
	var image domain.HomeThemeImage
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&image).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &image, err
}

func (r *HomeThemeImageRepository) DeleteAndRemoveThemeReference(image *domain.HomeThemeImage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var theme domain.UserTheme
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", image.UserID).
			First(&theme).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil {
			theme.HeroConfig.ImageIDs = removeImageID(theme.HeroConfig.ImageIDs, image.ID)
			if len(theme.HeroConfig.ImageIDs) == 0 {
				theme.HeroConfig.Enabled = false
			}
			if err := tx.Model(&domain.UserTheme{}).
				Where("user_id = ?", image.UserID).
				Update("hero_config", theme.HeroConfig).Error; err != nil {
				return err
			}
		}

		result := tx.Where("id = ? AND user_id = ?", image.ID, image.UserID).Delete(&domain.HomeThemeImage{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func removeImageID(ids []int64, target int64) []int64 {
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id != target {
			result = append(result, id)
		}
	}
	return result
}
