package service

import (
	"errors"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/util"

	"gorm.io/gorm"
)

var (
	ErrDuplicateUser = errors.New("duplicate user")
	ErrLoginFailed   = errors.New("login failed")
	ErrUnauthorized  = errors.New("unauthorized")
)

type AuthService struct {
	db            *gorm.DB
	users         *repository.UserRepository
	jwt           *util.JWTManager
	refreshTokens *repository.RefreshTokenRepository
	refreshTTL    time.Duration
	emailCodeSvc  *EmailCodeService
	inviteSvc     *InviteService
}

func NewAuthService(db *gorm.DB, users *repository.UserRepository, jwt *util.JWTManager, refreshTokens *repository.RefreshTokenRepository, refreshTTL time.Duration, emailCodeSvc *EmailCodeService, inviteSvc *InviteService) *AuthService {
	return &AuthService{
		db:            db,
		users:         users,
		jwt:           jwt,
		refreshTokens: refreshTokens,
		refreshTTL:    refreshTTL,
		emailCodeSvc:  emailCodeSvc,
		inviteSvc:     inviteSvc,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	var user *domain.User

	err := s.db.Transaction(func(tx *gorm.DB) error {
		txUserRepo := repository.NewUserRepository(tx)
		txEmailCodeRepo := repository.NewEmailCodeRepository(tx)
		txInviteCodeRepo := repository.NewInviteCodeRepository(tx)

		// 1. 检查用户名/邮箱重复
		exists, err := txUserRepo.ExistsByUsernameOrEmail(req.Username, req.Email)
		if err != nil {
			return err
		}
		if exists {
			return ErrDuplicateUser
		}

		// 2. 原子消费邮箱验证码
		if err := txEmailCodeRepo.ConsumeByEmailAndCode(req.Email, req.EmailCode); err != nil {
			if errors.Is(err, repository.ErrCodeInvalid) {
				return ErrCodeInvalid
			}
			return err
		}

		// 3. 查找邀请码并校验
		ic, err := txInviteCodeRepo.FindByCode(req.InviteCode)
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

		// 4. 创建用户
		hash, err := util.HashPassword(req.Password)
		if err != nil {
			return err
		}
		user = &domain.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hash,
			Nickname:     req.Nickname,
			Role:         domain.UserRoleUser,
			Status:       domain.UserStatusActive,
		}
		if err := txUserRepo.Create(user); err != nil {
			return err
		}

		// 5. 原子标记邀请码已用（WHERE is_used=false，防止并发抢占）
		if err := txInviteCodeRepo.MarkUsed(ic.ID, user.ID); err != nil {
			if errors.Is(err, repository.ErrInviteCodeUsed) {
				return ErrInviteCodeUsed
			}
			return err
		}

		return nil
	})

	if err != nil {
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
	active, err := s.refreshTokens.IsActive(refreshToken, user.ID)
	if err != nil || !active {
		return nil, ErrUnauthorized
	}

	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokens.Rotate(refreshToken, tokens.RefreshToken, user.ID, time.Now().Add(s.refreshTTL)); err != nil {
		return nil, ErrUnauthorized
	}
	return &dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         toUserInfo(user, false),
	}, nil
}

func (s *AuthService) Logout(refreshToken string) {
	if refreshToken != "" {
		_ = s.refreshTokens.Revoke(refreshToken)
	}
}

// ResetPassword 用邮箱验证码重置密码
func (s *AuthService) ResetPassword(email, emailCode, newPassword string) error {
	user, err := s.users.FindByEmail(email)
	if err != nil || user == nil {
		return ErrUserNotFound
	}
	if err := s.emailCodeSvc.VerifyAndConsume(email, emailCode); err != nil {
		return ErrCodeInvalid
	}
	hash, err := util.HashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.users.UpdatePassword(user.ID, hash); err != nil {
		return err
	}
	_ = s.refreshTokens.RevokeUser(user.ID)
	return nil
}

func (s *AuthService) buildAuthResponse(user *domain.User) (*dto.AuthResponse, error) {
	tokens, err := s.jwt.GenerateTokenPair(user.ID, user.Username)
	if err != nil {
		return nil, err
	}
	if err := s.refreshTokens.Create(tokens.RefreshToken, user.ID, time.Now().Add(s.refreshTTL)); err != nil {
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
		info.CanCustomJS = user.CanCustomJS
		info.CreatedAt = user.CreatedAt
	}
	return info
}

// toPublicUserInfo 构建仅含公开字段的用户信息（用于文章 Author 等公开场景）
func toPublicUserInfo(user *domain.User) dto.PublicUserInfo {
	return dto.PublicUserInfo{
		Username:  user.Username,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
	}
}
