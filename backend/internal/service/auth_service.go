package service

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/util"
)

var (
	ErrDuplicateUser = errors.New("duplicate user")
	ErrLoginFailed   = errors.New("login failed")
	ErrUnauthorized  = errors.New("unauthorized")
)

type AuthService struct {
	users *repository.UserRepository
	jwt   *util.JWTManager
}

func NewAuthService(users *repository.UserRepository, jwt *util.JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	exists, err := s.users.ExistsByUsernameOrEmail(req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateUser
	}

	hash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hash,
		Nickname:     req.Nickname,
		Role:         domain.UserRoleUser,
		Status:       domain.UserStatusActive,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}

	return s.buildAuthResponse(user)
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.users.FindByLogin(req.Login)
	if err != nil || user == nil {
		return nil, ErrLoginFailed
	}
	if user.Status != domain.UserStatusActive {
		return nil, ErrLoginFailed
	}
	if !util.CheckPassword(user.PasswordHash, req.Password) {
		return nil, ErrLoginFailed
	}
	return s.buildAuthResponse(user)
}

func (s *AuthService) Refresh(refreshToken string) (*dto.AuthResponse, error) {
	claims, err := s.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrUnauthorized
	}
	user, err := s.users.FindByID(claims.UserID)
	if err != nil || user == nil || user.Status != domain.UserStatusActive {
		return nil, ErrUnauthorized
	}
	return s.buildAuthResponse(user)
}

func (s *AuthService) buildAuthResponse(user *domain.User) (*dto.AuthResponse, error) {
	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	return &dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         toUserInfo(user, false),
	}, nil
}

func toUserInfo(user *domain.User, full bool) dto.UserInfo {
	info := dto.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}
	if full {
		info.Email = user.Email
		info.Bio = user.Bio
		info.Role = string(user.Role)
		info.Status = string(user.Status)
		info.CanSitePin = user.CanSitePin
		info.CreatedAt = user.CreatedAt
	}
	return info
}
