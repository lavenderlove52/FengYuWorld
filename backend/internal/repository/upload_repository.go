package repository

import (
	"errors"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrUploadReferenceNotFound = errors.New("upload reference not found")

type UploadRepository struct {
	db *gorm.DB
}

func NewUploadRepository(db *gorm.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

func (r *UploadRepository) Create(u *domain.Upload) error {
	return r.db.Create(u).Error
}

// Acquire creates or reuses a physical asset and its independently owned logical reference.
// materialize runs while the asset row is locked, before the transaction is committed.
func (r *UploadRepository) Acquire(asset *domain.UploadAsset, ref *domain.Upload, materialize func() error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "content_hash"}, {Name: "variant"}},
			DoNothing: true,
		}).Create(asset).Error; err != nil {
			return err
		}
		asset.ID = 0
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("content_hash = ? AND variant = ?", asset.ContentHash, asset.Variant).
			First(asset).Error; err != nil {
			return err
		}
		if err := materialize(); err != nil {
			return err
		}

		ref.AssetID = &asset.ID
		ref.Path = asset.Path
		ref.URL = asset.URL
		ref.Mime = asset.Mime
		ref.SizeBytes = asset.SizeBytes
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(ref).Error; err != nil {
			return err
		}
		ref.ID = 0
		return tx.Where(
			"asset_id = ? AND user_id = ? AND purpose = ? AND owner_type = ? AND owner_key = ?",
			asset.ID, ref.UserID, *ref.Purpose, *ref.OwnerType, *ref.OwnerKey,
		).First(ref).Error
	})
}

// Release removes one logical reference and deletes the asset only when no references remain.
// Physical deletion happens only after the database transaction commits successfully.
func (r *UploadRepository) Release(userID, referenceID int64, remove func(path string) error) error {
	var unusedPath string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var ref domain.Upload
		if err := tx.Where("id = ? AND user_id = ? AND asset_id IS NOT NULL", referenceID, userID).
			First(&ref).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUploadReferenceNotFound
			}
			return err
		}

		var asset domain.UploadAsset
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&asset, *ref.AssetID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&ref).Error; err != nil {
			return err
		}

		var count int64
		if err := tx.Model(&domain.Upload{}).Where("asset_id = ?", asset.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		if err := tx.Delete(&asset).Error; err != nil {
			return err
		}
		unusedPath = asset.Path
		return nil
	})
	if err != nil || unusedPath == "" {
		return err
	}
	return remove(unusedPath)
}
