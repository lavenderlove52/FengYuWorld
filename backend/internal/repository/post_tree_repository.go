package repository

import (
	"errors"

	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type PostTreeRepository struct {
	db *gorm.DB
}

func NewPostTreeRepository(db *gorm.DB) *PostTreeRepository {
	return &PostTreeRepository{db: db}
}

func (r *PostTreeRepository) ListByUser(userID int64) ([]domain.PostTreeNode, error) {
	var nodes []domain.PostTreeNode
	err := r.db.Preload("Post.Author").Preload("Post.Tags").
		Where("user_id = ?", userID).
		Order("sort_order ASC, id ASC").
		Find(&nodes).Error
	return nodes, err
}

func (r *PostTreeRepository) FindByIDAndUser(id, userID int64) (*domain.PostTreeNode, error) {
	var node domain.PostTreeNode
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &node, err
}

func (r *PostTreeRepository) Create(node *domain.PostTreeNode) error {
	return r.db.Create(node).Error
}

func (r *PostTreeRepository) CreateWithTx(tx *gorm.DB, node *domain.PostTreeNode) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(node).Error
}

func (r *PostTreeRepository) Update(node *domain.PostTreeNode) error {
	return r.db.Save(node).Error
}

func (r *PostTreeRepository) UpdateWithTx(tx *gorm.DB, node *domain.PostTreeNode) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Save(node).Error
}

func (r *PostTreeRepository) Delete(id, userID int64) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&domain.PostTreeNode{}).Error
}

func (r *PostTreeRepository) HasChildren(id, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.PostTreeNode{}).
		Where("parent_id = ? AND user_id = ?", id, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *PostTreeRepository) HasFolderChildren(id, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.PostTreeNode{}).
		Where("parent_id = ? AND user_id = ? AND node_type = ?", id, userID, domain.PostTreeNodeTypeFolder).
		Count(&count).Error
	return count > 0, err
}

func (r *PostTreeRepository) HasPostChildren(id, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&domain.PostTreeNode{}).
		Where("parent_id = ? AND user_id = ? AND node_type = ?", id, userID, domain.PostTreeNodeTypePost).
		Count(&count).Error
	return count > 0, err
}

func (r *PostTreeRepository) FindPostNodeByPostID(userID, postID int64) (*domain.PostTreeNode, error) {
	var node domain.PostTreeNode
	err := r.db.Where("user_id = ? AND post_id = ? AND node_type = ?", userID, postID, domain.PostTreeNodeTypePost).
		First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &node, err
}
