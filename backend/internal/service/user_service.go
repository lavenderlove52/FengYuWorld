package service

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrAdminForbidden = errors.New("admin forbidden")
)

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

func (s *UserService) GrantCustomJS(adminID int64, targetUserID int64) error {
	admin, err := s.users.FindByID(adminID)
	if err != nil || admin == nil || admin.Role != domain.UserRoleAdmin {
		return ErrAdminForbidden
	}
	target, err := s.users.FindByID(targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	return s.users.UpdateCanCustomJS(targetUserID, true)
}

func (s *UserService) RevokeCustomJS(adminID int64, targetUserID int64) error {
	admin, err := s.users.FindByID(adminID)
	if err != nil || admin == nil || admin.Role != domain.UserRoleAdmin {
		return ErrAdminForbidden
	}
	target, err := s.users.FindByID(targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	return s.users.UpdateCanCustomJS(targetUserID, false)
}

func (s *UserService) ListUsers(keyword string, page, pageSize int) ([]dto.UserInfo, int64, error) {
	users, total, err := s.users.ListUsers(keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	infos := make([]dto.UserInfo, len(users))
	for i, u := range users {
		infos[i] = toUserInfo(&u, true)
	}
	return infos, total, nil
}
