package repository

import (
	"errors"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByUsername(username string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) FindByLogin(login string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("username = ? OR email = ?", login, login).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) ExistsByUsernameOrEmail(username, email string) (bool, error) {
	var count int64
	err := r.db.Model(&domain.User{}).
		Where("username = ? OR email = ?", username, email).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) UpdatePassword(userID int64, hash string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("password_hash", hash).Error
}

func (r *UserRepository) UpdateCanCustomJS(userID int64, value bool) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("can_custom_js", value).Error
}

func (r *UserRepository) ListUsers(keyword string, page, pageSize int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.Model(&domain.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR nickname LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}
