package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(token string, userID int64, expiresAt time.Time) error {
	return r.db.Create(&domain.RefreshToken{
		TokenHash: tokenHash(token),
		UserID:    userID,
		ExpiresAt: expiresAt,
	}).Error
}

func (r *RefreshTokenRepository) IsActive(token string, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.RefreshToken{}).
		Where("token_hash = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash(token), userID, time.Now()).
		Count(&count).Error
	return count > 0, err
}

func (r *RefreshTokenRepository) Rotate(oldToken, newToken string, userID int64, newExpiresAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Model(&domain.RefreshToken{}).
			Where("token_hash = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", tokenHash(oldToken), userID, now).
			Update("revoked_at", now)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(&domain.RefreshToken{
			TokenHash: tokenHash(newToken),
			UserID:    userID,
			ExpiresAt: newExpiresAt,
		}).Error
	})
}

func (r *RefreshTokenRepository) Revoke(token string) error {
	now := time.Now()
	return r.db.Model(&domain.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash(token)).
		Update("revoked_at", now).Error
}

func (r *RefreshTokenRepository) RevokeUser(userID int64) error {
	now := time.Now()
	return r.db.Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

func tokenHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
