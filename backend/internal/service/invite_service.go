package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInviteCodeInvalid  = errors.New("invite code invalid")
	ErrInviteCodeUsed     = errors.New("invite code already used")
	ErrInviteCodeExpired  = errors.New("invite code expired")
	ErrInviteCodeNotFound = errors.New("invite code not found or already used")
)

const inviteCodeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type InviteService struct {
	repo *repository.InviteCodeRepository
}

func NewInviteService(repo *repository.InviteCodeRepository) *InviteService {
	return &InviteService{repo: repo}
}

// GenerateCodes 批量生成邀请码（8位大写字母数字，crypto/rand）
// 逐条插入并在唯一键冲突时最多重试3次
func (s *InviteService) GenerateCodes(createdBy int64, count int, expiresAt *time.Time) ([]domain.InviteCode, error) {
	result := make([]domain.InviteCode, 0, count)
	for i := 0; i < count; i++ {
		var saved *domain.InviteCode
		for retry := 0; retry < 3; retry++ {
			code, err := generateInviteCode(8)
			if err != nil {
				return nil, err
			}
			ic := &domain.InviteCode{Code: code, CreatedBy: createdBy, ExpiresAt: expiresAt}
			if err := s.repo.CreateOne(ic); err == nil {
				saved = ic
				break
			} else if !isUniqueConstraintError(err) {
				return nil, err
			}
		}
		if saved == nil {
			return nil, errors.New("生成邀请码冲突，请重试")
		}
		result = append(result, *saved)
	}
	return result, nil
}

// List 分页列表
func (s *InviteService) List(page, pageSize int) ([]domain.InviteCode, int64, error) {
	return s.repo.List(page, pageSize)
}

// Validate 仅校验邀请码有效性，不消费
func (s *InviteService) Validate(code string) error {
	ic, err := s.repo.FindByCode(code)
	if err != nil {
		return err
	}
	if ic == nil {
		return ErrInviteCodeInvalid
	}
	if ic.IsUsed {
		return ErrInviteCodeUsed
	}
	if ic.ExpiresAt != nil && time.Now().After(*ic.ExpiresAt) {
		return ErrInviteCodeExpired
	}
	return nil
}

// VerifyAndConsume 校验邀请码（未使用且未过期），成功后原子消费
func (s *InviteService) VerifyAndConsume(code string, usedBy int64) error {
	ic, err := s.repo.FindByCode(code)
	if err != nil {
		return err
	}
	if ic == nil {
		return ErrInviteCodeInvalid
	}
	if ic.IsUsed {
		return ErrInviteCodeUsed
	}
	if ic.ExpiresAt != nil && time.Now().After(*ic.ExpiresAt) {
		return ErrInviteCodeExpired
	}
	if err := s.repo.MarkUsed(ic.ID, usedBy); err != nil {
		if errors.Is(err, repository.ErrInviteCodeUsed) {
			return ErrInviteCodeUsed
		}
		return err
	}
	return nil
}

// Delete 删除未使用的邀请码
func (s *InviteService) Delete(id int64) error {
	if err := s.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrInviteCodeNotFound) {
			return ErrInviteCodeNotFound
		}
		return err
	}
	return nil
}

func generateInviteCode(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(inviteCodeCharset))))
		if err != nil {
			return "", err
		}
		result[i] = inviteCodeCharset[n.Int64()]
	}
	return string(result), nil
}

// isUniqueConstraintError 检查是否为 PostgreSQL 唯一键冲突错误（错误码 23505）
func isUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
