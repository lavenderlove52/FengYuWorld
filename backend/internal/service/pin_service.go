package service

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/markdown"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/util"
)

const maxSitePins = 20

var (
	ErrPinNotFound      = errors.New("pin not found")
	ErrPinForbidden     = errors.New("no permission to pin")
	ErrPinLimitExceeded = errors.New("site pin limit exceeded")
	ErrAlreadyPinned    = errors.New("already pinned")
	ErrPinNotOwner      = errors.New("can only pin your own posts")
)

type PinSummary struct {
	PostID       int64           `json:"post_id"`
	DisplayOrder int             `json:"display_order"`
	Post         dto.PostSummary `json:"post"`
}

type PinService struct {
	pins  *repository.PinRepository
	posts *repository.PostRepository
	users *repository.UserRepository
}

func NewPinService(
	pins *repository.PinRepository,
	posts *repository.PostRepository,
	users *repository.UserRepository,
) *PinService {
	return &PinService{pins: pins, posts: posts, users: users}
}

func (s *PinService) GetSitePins() ([]PinSummary, error) {
	pins, err := s.pins.GetSitePins()
	if err != nil {
		return nil, err
	}
	result := make([]PinSummary, 0, len(pins))
	for i := range pins {
		ps := sitePinToSummary(&pins[i])
		result = append(result, ps)
	}
	return result, nil
}

func (s *PinService) SitePinPost(operatorID int64, postID int64) error {
	operator, err := s.users.FindByID(operatorID)
	if err != nil || operator == nil {
		return ErrUserNotFound
	}
	if operator.Role != domain.UserRoleAdmin && !operator.CanSitePin {
		return ErrPinForbidden
	}

	post, err := s.posts.FindByID(postID)
	if err != nil || post == nil || post.IsDeleted {
		return ErrPostNotFound
	}
	if post.Status != domain.PostStatusPublished || post.Visibility != domain.PostVisibilityPublic {
		return ErrPinForbidden
	}

	exists, err := s.pins.SitePinExists(postID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyPinned
	}

	count, err := s.pins.CountSitePins()
	if err != nil {
		return err
	}
	if count >= maxSitePins {
		return ErrPinLimitExceeded
	}

	return s.pins.AddSitePin(postID, operatorID, int(count)+1)
}

func (s *PinService) SiteUnpinPost(operatorID int64, postID int64) error {
	operator, err := s.users.FindByID(operatorID)
	if err != nil || operator == nil {
		return ErrUserNotFound
	}
	if operator.Role != domain.UserRoleAdmin && !operator.CanSitePin {
		return ErrPinForbidden
	}

	exists, err := s.pins.SitePinExists(postID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPinNotFound
	}

	return s.pins.RemoveSitePin(postID)
}

func (s *PinService) GetUserPins(userID int64) ([]PinSummary, error) {
	pins, err := s.pins.GetUserPins(userID)
	if err != nil {
		return nil, err
	}
	result := make([]PinSummary, 0, len(pins))
	for i := range pins {
		ps := userPinToSummary(&pins[i])
		result = append(result, ps)
	}
	return result, nil
}

func (s *PinService) UserPinPost(userID int64, postID int64) error {
	post, err := s.posts.FindByID(postID)
	if err != nil || post == nil || post.IsDeleted {
		return ErrPostNotFound
	}
	if post.AuthorID != userID {
		return ErrPinNotOwner
	}

	exists, err := s.pins.UserPinExists(userID, postID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyPinned
	}

	count, err := s.pins.CountUserPins(userID)
	if err != nil {
		return err
	}
	return s.pins.AddUserPin(userID, postID, int(count)+1)
}

func (s *PinService) UserUnpinPost(userID int64, postID int64) error {
	exists, err := s.pins.UserPinExists(userID, postID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPinNotFound
	}
	return s.pins.RemoveUserPin(userID, postID)
}

func (s *PinService) GrantSitePinPermission(adminID int64, targetUserID int64) error {
	admin, err := s.users.FindByID(adminID)
	if err != nil || admin == nil {
		return ErrUserNotFound
	}
	if admin.Role != domain.UserRoleAdmin {
		return ErrPinForbidden
	}
	target, err := s.users.FindByID(targetUserID)
	if err != nil || target == nil {
		return ErrUserNotFound
	}
	target.CanSitePin = true
	return s.users.Update(target)
}

func (s *PinService) RevokeSitePinPermission(adminID int64, targetUserID int64) error {
	admin, err := s.users.FindByID(adminID)
	if err != nil || admin == nil {
		return ErrUserNotFound
	}
	if admin.Role != domain.UserRoleAdmin {
		return ErrPinForbidden
	}
	target, err := s.users.FindByID(targetUserID)
	if err != nil || target == nil {
		return ErrUserNotFound
	}
	target.CanSitePin = false
	return s.users.Update(target)
}

func sitePinToSummary(pin *domain.SitePin) PinSummary {
	ps := PinSummary{
		PostID:       pin.PostID,
		DisplayOrder: pin.DisplayOrder,
	}
	if pin.Post != nil {
		ps.Post = postToSummaryForPin(pin.Post)
	}
	return ps
}

func userPinToSummary(pin *domain.UserPin) PinSummary {
	ps := PinSummary{
		PostID:       pin.PostID,
		DisplayOrder: pin.DisplayOrder,
	}
	if pin.Post != nil {
		ps.Post = postToSummaryForPin(pin.Post)
	}
	return ps
}

func postToSummaryForPin(post *domain.Post) dto.PostSummary {
	summary := dto.PostSummary{
		ID:            post.ID,
		Title:         post.Title,
		Slug:          post.Slug,
		CoverImageURL: post.CoverImageURL,
		Summary:       markdown.Summary(post.ContentHTML, 200),
		Status:        string(post.Status),
		Visibility:    string(post.Visibility),
		ViewCount:     post.ViewCount,
		CreatedAt:     util.FormatTime(post.CreatedAt),
	}
	if post.PublishedAt != nil {
		t := util.FormatTime(*post.PublishedAt)
		summary.PublishedAt = &t
	}
	if post.Author != nil {
		summary.Author = toUserInfo(post.Author, false)
	}
	return summary
}
