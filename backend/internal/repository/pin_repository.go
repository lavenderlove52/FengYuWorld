package repository

import (
	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type PinRepository struct {
	db *gorm.DB
}

func NewPinRepository(db *gorm.DB) *PinRepository {
	return &PinRepository{db: db}
}

func (r *PinRepository) GetSitePins() ([]domain.SitePin, error) {
	var pins []domain.SitePin
	err := r.db.Preload("Post").Preload("Post.Author").
		Joins("JOIN posts ON posts.id = site_pins.post_id").
		Where("posts.is_deleted = ? AND posts.status = ? AND posts.visibility = ?",
			false, domain.PostStatusPublished, domain.PostVisibilityPublic).
		Order("display_order ASC, id ASC").
		Find(&pins).Error
	return pins, err
}

func (r *PinRepository) AddSitePin(postID int64, pinnedBy int64, order int) error {
	pin := &domain.SitePin{
		PostID:       postID,
		PinnedBy:     pinnedBy,
		DisplayOrder: order,
	}
	return r.db.Create(pin).Error
}

func (r *PinRepository) RemoveSitePin(postID int64) error {
	return r.db.Where("post_id = ?", postID).Delete(&domain.SitePin{}).Error
}

func (r *PinRepository) ReorderSitePins(pins []domain.SitePin) error {
	for i := range pins {
		if err := r.db.Model(&domain.SitePin{}).
			Where("id = ?", pins[i].ID).
			Update("display_order", pins[i].DisplayOrder).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *PinRepository) CountSitePins() (int64, error) {
	var count int64
	err := r.db.Model(&domain.SitePin{}).
		Joins("JOIN posts ON posts.id = site_pins.post_id").
		Where("posts.is_deleted = ? AND posts.status = ? AND posts.visibility = ?",
			false, domain.PostStatusPublished, domain.PostVisibilityPublic).
		Count(&count).Error
	return count, err
}

func (r *PinRepository) SitePinExists(postID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.SitePin{}).Where("post_id = ?", postID).Count(&count).Error
	return count > 0, err
}

func (r *PinRepository) GetUserPins(userID int64) ([]domain.UserPin, error) {
	var pins []domain.UserPin
	err := r.db.Preload("Post").Preload("Post.Author").
		Joins("JOIN posts ON posts.id = user_pins.post_id").
		Where("user_id = ?", userID).
		Where("posts.is_deleted = ?", false).
		Order("display_order ASC, id ASC").
		Find(&pins).Error
	return pins, err
}

func (r *PinRepository) AddUserPin(userID int64, postID int64, order int) error {
	pin := &domain.UserPin{
		UserID:       userID,
		PostID:       postID,
		DisplayOrder: order,
	}
	return r.db.Create(pin).Error
}

func (r *PinRepository) RemoveUserPin(userID int64, postID int64) error {
	return r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&domain.UserPin{}).Error
}

func (r *PinRepository) UserPinExists(userID int64, postID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.UserPin{}).
		Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

func (r *PinRepository) CountUserPins(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&domain.UserPin{}).
		Joins("JOIN posts ON posts.id = user_pins.post_id").
		Where("user_id = ?", userID).
		Where("posts.is_deleted = ?", false).
		Count(&count).Error
	return count, err
}
