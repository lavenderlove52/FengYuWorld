package domain

import "time"

type Upload struct {
	ID        int64        `gorm:"primaryKey"`
	UserID    int64        `gorm:"not null"`
	AssetID   *int64       `gorm:"index"`
	Asset     *UploadAsset `gorm:"foreignKey:AssetID"`
	Purpose   *string      `gorm:"size:50"`
	OwnerType *string      `gorm:"size:50"`
	OwnerKey  *string      `gorm:"size:200"`
	Path      string       `gorm:"size:500;not null"`
	URL       string       `gorm:"size:500;not null"`
	Mime      string       `gorm:"size:120;not null"`
	SizeBytes int64        `gorm:"not null"`
	CreatedAt time.Time
}

func (Upload) TableName() string { return "uploads" }

type UploadAsset struct {
	ID                int64  `gorm:"primaryKey"`
	ContentHash       string `gorm:"size:64;not null;uniqueIndex:upload_assets_content_hash_variant_key"`
	Variant           string `gorm:"size:50;not null;uniqueIndex:upload_assets_content_hash_variant_key"`
	Path              string `gorm:"size:500;not null"`
	URL               string `gorm:"size:500;not null"`
	Mime              string `gorm:"size:120;not null"`
	SizeBytes         int64  `gorm:"not null"`
	OriginalSizeBytes int64  `gorm:"not null"`
	CreatedAt         time.Time
}

func (UploadAsset) TableName() string { return "upload_assets" }
