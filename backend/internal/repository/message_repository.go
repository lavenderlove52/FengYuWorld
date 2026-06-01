package repository

import (
	"my-blog/backend/internal/domain"

	"gorm.io/gorm"
)

type MessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// GetConversation 游标分页获取对话历史（before=0 表示最新）
func (r *MessageRepository) GetConversation(userA, userB int64, before int64, limit int) ([]domain.Message, error) {
	q := r.db.Where(
		"LEAST(from_user, to_user) = LEAST(?::bigint, ?::bigint) AND GREATEST(from_user, to_user) = GREATEST(?::bigint, ?::bigint)",
		userA, userB, userA, userB,
	)
	if before > 0 {
		q = q.Where("id < ?", before)
	}
	var msgs []domain.Message
	err := q.Order("id DESC").Limit(limit).Find(&msgs).Error
	return msgs, err
}

// Create 保存消息
func (r *MessageRepository) Create(msg *domain.Message) error {
	return r.db.Create(msg).Error
}

// MarkRead 标记两人对话中发给 toUser 的所有消息为已读
func (r *MessageRepository) MarkRead(fromUser, toUser int64) error {
	return r.db.Model(&domain.Message{}).
		Where("from_user = ? AND to_user = ? AND is_read = false", fromUser, toUser).
		Update("is_read", true).Error
}

// GetUnreadCounts 获取 userID 的所有未读计数，返回 map[fromUserID]count
func (r *MessageRepository) GetUnreadCounts(userID int64) (map[int64]int64, error) {
	type row struct {
		FromUser int64
		Count    int64
	}
	var rows []row
	err := r.db.Model(&domain.Message{}).
		Select("from_user, COUNT(*) as count").
		Where("to_user = ? AND is_read = false", userID).
		Group("from_user").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]int64, len(rows))
	for _, r := range rows {
		result[r.FromUser] = r.Count
	}
	return result, nil
}
