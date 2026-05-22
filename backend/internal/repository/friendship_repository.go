package repository

import (
	"errors"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type FriendshipRepository struct {
	db *gorm.DB
}

func NewFriendshipRepository(db *gorm.DB) *FriendshipRepository {
	return &FriendshipRepository{db: db}
}

func (r *FriendshipRepository) Create(f *domain.Friendship) error {
	return r.db.Create(f).Error
}

func (r *FriendshipRepository) FindByID(id int64) (*domain.Friendship, error) {
	var f domain.Friendship
	err := r.db.First(&f, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &f, err
}

func (r *FriendshipRepository) FindPair(userID, friendID int64) (*domain.Friendship, error) {
	var f domain.Friendship
	err := r.db.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID,
	).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &f, err
}

func (r *FriendshipRepository) Update(f *domain.Friendship) error {
	return r.db.Save(f).Error
}

func (r *FriendshipRepository) DeletePair(userID, friendID int64) error {
	return r.db.Where(
		"(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID,
	).Delete(&domain.Friendship{}).Error
}

func (r *FriendshipRepository) ListReceived(userID int64) ([]domain.Friendship, error) {
	var list []domain.Friendship
	err := r.db.Where("friend_id = ? AND status = ?", userID, domain.FriendshipStatusPending).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *FriendshipRepository) ListSent(userID int64) ([]domain.Friendship, error) {
	var list []domain.Friendship
	err := r.db.Where("user_id = ? AND status = ?", userID, domain.FriendshipStatusPending).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *FriendshipRepository) ListFriends(userID int64) ([]domain.Friendship, error) {
	var list []domain.Friendship
	err := r.db.Where(
		"status = ? AND (user_id = ? OR friend_id = ?)",
		domain.FriendshipStatusAccepted, userID, userID,
	).Order("updated_at DESC").Find(&list).Error
	return list, err
}

func (r *FriendshipRepository) AreFriends(userA, userB int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Friendship{}).
		Where("status = ?", domain.FriendshipStatusAccepted).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userA, userB, userB, userA).
		Count(&count).Error
	return count > 0, err
}

func (r *FriendshipRepository) ListAcceptedFriendIDs(userID int64) ([]int64, error) {
	list, err := r.ListFriends(userID)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(list))
	for _, f := range list {
		if f.UserID == userID {
			ids = append(ids, f.FriendID)
		} else {
			ids = append(ids, f.UserID)
		}
	}
	return ids, nil
}
