package repository

import (
	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type UploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

func (r *UploadRepository) Create(u *domain.Upload) error {
	return r.db.Create(u).Error
}
