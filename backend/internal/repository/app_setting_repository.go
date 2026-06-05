package repository

import (
	"errors"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AppSettingRepository struct {
	db *gorm.DB
}

func NewAppSettingRepository(db *gorm.DB) *AppSettingRepository {
	return &AppSettingRepository{db: db}
}

func (r *AppSettingRepository) Get(key string) (*domain.AppSetting, error) {
	var setting domain.AppSetting
	err := r.db.First(&setting, "key = ?", key).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &setting, err
}

func (r *AppSettingRepository) Upsert(key, value string) error {
	setting := &domain.AppSetting{Key: key, Value: value}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(setting).Error
}
