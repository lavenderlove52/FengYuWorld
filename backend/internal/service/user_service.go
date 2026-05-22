package service

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")

type UserService struct {
	users *repository.UserRepository
}

func NewUserService(users *repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetMe(userID int64) (*dto.UserInfo, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	info := toUserInfo(user, true)
	return &info, nil
}

func (s *UserService) UpdateMe(userID int64, req dto.UpdateProfileRequest) (*dto.UserInfo, error) {
	user, err := s.users.FindByID(userID)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if err := s.users.Update(user); err != nil {
		return nil, err
	}
	info := toUserInfo(user, true)
	return &info, nil
}

func (s *UserService) GetByUsername(username string) (*domain.User, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
