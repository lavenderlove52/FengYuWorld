package repository

import (
	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type UserNtfyTopicRepository struct{ db *gorm.DB }

func NewUserNtfyTopicRepository(db *gorm.DB) *UserNtfyTopicRepository {
	return &UserNtfyTopicRepository{db: db}
}

func (r *UserNtfyTopicRepository) List(userID int64) ([]domain.UserNtfyTopic, error) {
	var topics []domain.UserNtfyTopic
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&topics).Error
	return topics, err
}

func (r *UserNtfyTopicRepository) Create(topic *domain.UserNtfyTopic) error {
	return r.db.Create(topic).Error
}

func (r *UserNtfyTopicRepository) GetOwned(userID, topicID int64) (*domain.UserNtfyTopic, error) {
	var topic domain.UserNtfyTopic
	err := r.db.Where("id = ? AND user_id = ?", topicID, userID).First(&topic).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &topic, err
}

func (r *UserNtfyTopicRepository) Update(topic *domain.UserNtfyTopic) error {
	return r.db.Save(topic).Error
}

func (r *UserNtfyTopicRepository) GetByPublicToken(token string) (*domain.UserNtfyTopic, error) {
	var topic domain.UserNtfyTopic
	err := r.db.Where("public_enabled = ? AND public_token = ?", true, token).First(&topic).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &topic, err
}
