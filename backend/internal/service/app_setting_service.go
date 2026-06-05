package service

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"
)

var ErrSuperAdminForbidden = errors.New("super admin forbidden")

type AppSettings struct {
	FriendChatEnabled bool `json:"friend_chat_enabled"`
}

type AppSettingService struct {
	settings *repository.AppSettingRepository
	users    *repository.UserRepository
}

func NewAppSettingService(settings *repository.AppSettingRepository, users *repository.UserRepository) *AppSettingService {
	return &AppSettingService{settings: settings, users: users}
}

func (s *AppSettingService) GetPublicSettings() (*AppSettings, error) {
	enabled, err := s.IsFriendChatEnabled()
	if err != nil {
		return nil, err
	}
	return &AppSettings{FriendChatEnabled: enabled}, nil
}

func (s *AppSettingService) IsFriendChatEnabled() (bool, error) {
	setting, err := s.settings.Get(domain.AppSettingFriendChatEnabled)
	if err != nil {
		return false, err
	}
	if setting == nil {
		return true, nil
	}
	return setting.Value != "false", nil
}

func (s *AppSettingService) UpdateFriendChatEnabled(actorID int64, enabled bool) (*AppSettings, error) {
	user, err := s.users.FindByID(actorID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	if !domain.IsSuperAdminRole(user.Role) {
		return nil, ErrSuperAdminForbidden
	}

	value := "false"
	if enabled {
		value = "true"
	}
	if err := s.settings.Upsert(domain.AppSettingFriendChatEnabled, value); err != nil {
		return nil, err
	}
	return &AppSettings{FriendChatEnabled: enabled}, nil
}
