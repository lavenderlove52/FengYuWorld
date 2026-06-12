package repository

import (
	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserPushSettingRepository struct {
	db *gorm.DB
}

func NewUserPushSettingRepository(db *gorm.DB) *UserPushSettingRepository {
	return &UserPushSettingRepository{db: db}
}

func (r *UserPushSettingRepository) Get(userID int64) (*domain.UserPushSetting, error) {
	var setting domain.UserPushSetting
	err := r.db.Where("user_id = ?", userID).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		return &domain.UserPushSetting{UserID: userID}, nil
	}
	return &setting, err
}

func (r *UserPushSettingRepository) Upsert(setting *domain.UserPushSetting) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"bark_endpoint", "ntfy_topic", "ntfy_public_enabled", "ntfy_public_token", "updated_at"}),
	}).Create(setting).Error
}

func (r *UserPushSettingRepository) GetByPublicToken(token string) (*domain.UserPushSetting, error) {
	var setting domain.UserPushSetting
	err := r.db.Where("ntfy_public_enabled = ? AND ntfy_public_token = ?", true, token).First(&setting).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &setting, err
}
