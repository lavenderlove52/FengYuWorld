package repository

import (
	"errors"
	"time"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

var (
	ErrInviteCodeUsed     = errors.New("invite code already used")
	ErrInviteCodeNotFound = errors.New("invite code not found or already used")
)

type InviteCodeRepository struct {
	db *gorm.DB
}

func NewInviteCodeRepository(db *gorm.DB) *InviteCodeRepository {
	return &InviteCodeRepository{db: db}
}

func (r *InviteCodeRepository) Create(codes []*domain.InviteCode) error {
	return r.db.Create(codes).Error
}

// CreateOne 创建单条邀请码，供批量生成时逐条重试
func (r *InviteCodeRepository) CreateOne(code *domain.InviteCode) error {
	return r.db.Create(code).Error
}

func (r *InviteCodeRepository) FindByCode(code string) (*domain.InviteCode, error) {
	var ic domain.InviteCode
	err := r.db.Where("code = ?", code).First(&ic).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ic, err
}

func (r *InviteCodeRepository) List(page, pageSize int) ([]domain.InviteCode, int64, error) {
	var codes []domain.InviteCode
	var total int64

	q := r.db.Model(&domain.InviteCode{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := q.Preload("Creator").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&codes).Error
	return codes, total, err
}

// MarkUsed 原子标记邀请码为已使用，防止并发重复消费
func (r *InviteCodeRepository) MarkUsed(id int64, usedBy int64) error {
	now := time.Now()
	result := r.db.Model(&domain.InviteCode{}).
		Where("id = ? AND is_used = false", id).
		Updates(map[string]interface{}{
			"is_used": true,
			"used_by": usedBy,
			"used_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInviteCodeUsed
	}
	return nil
}

// Delete 删除未使用的邀请码，若不存在或已使用返回 ErrInviteCodeNotFound
func (r *InviteCodeRepository) Delete(id int64) error {
	result := r.db.Where("id = ? AND is_used = false", id).Delete(&domain.InviteCode{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInviteCodeNotFound
	}
	return nil
}
