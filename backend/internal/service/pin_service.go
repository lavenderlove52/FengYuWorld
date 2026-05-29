package service

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/markdown"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/util"
)

const (
	maxSitePins = 20
	maxUserPins = 30
)

var (
	ErrPinNotFound          = errors.New("pin not found")
	ErrPinForbidden         = errors.New("no permission to pin")
	ErrPinLimitExceeded     = errors.New("site pin limit exceeded")
	ErrUserPinLimitExceeded = errors.New("user pin limit exceeded")
	ErrAlreadyPinned        = errors.New("already pinned")
	ErrPinNotOwner          = errors.New("can only pin your own posts")
)

type PinSummary struct {
	PostID       int64           `json:"post_id"`
	DisplayOrder int             `json:"display_order"`
	Post         dto.PostSummary `json:"post"`
}

type PinService struct {
	pins        *repository.PinRepository
	posts       *repository.PostRepository
	users       *repository.UserRepository
	friendships *repository.FriendshipRepository
	policy      *PostPolicy
}

func NewPinService(
	pins *repository.PinRepository,
	posts *repository.PostRepository,
	users *repository.UserRepository,
	friendships *repository.FriendshipRepository,
	policy *PostPolicy,
) *PinService {
	return &PinService{pins: pins, posts: posts, users: users, friendships: friendships, policy: policy}
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

// GetPublicUserPins 返回指定用户的置顶文章，按 viewerID 的权限过滤不可见文章。
func (s *PinService) GetPublicUserPins(username string, viewerID *int64) ([]PinSummary, error) {
	user, err := s.users.FindByUsername(username)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	// M4: 锁定账号的置顶对外隐藏，返回空列表而非错误
	if user.Status == domain.UserStatusLocked {
		return []PinSummary{}, nil
	}
	pins, err := s.pins.GetUserPins(user.ID)
	if err != nil {
		return nil, err
	}

	// M3: 预查一次好友关系，避免对每条 post 重复查库
	var isFriend bool
	if viewerID != nil && *viewerID != user.ID {
		isFriend, err = s.friendships.AreFriends(user.ID, *viewerID)
		if err != nil {
			return nil, err
		}
	}

	result := make([]PinSummary, 0, len(pins))
	for i := range pins {
		if pins[i].Post == nil {
			continue
		}
		post := pins[i].Post

		// 作者本人可以看到自己所有置顶
		if viewerID != nil && *viewerID == post.AuthorID {
			result = append(result, userPinToSummary(&pins[i]))
			continue
		}

		// 非作者：只能看已发布且未删除的文章
		if post.IsDeleted || post.Status != domain.PostStatusPublished {
			continue
		}

		switch post.Visibility {
		case domain.PostVisibilityPublic:
			result = append(result, userPinToSummary(&pins[i]))
		case domain.PostVisibilityFriends:
			if viewerID != nil && isFriend {
				result = append(result, userPinToSummary(&pins[i]))
			}
		case domain.PostVisibilitySpecified:
			if viewerID != nil {
				visible, verr := s.posts.IsVisibleToUser(post.ID, *viewerID)
				if verr == nil && visible {
					result = append(result, userPinToSummary(&pins[i]))
				}
			}
		}
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
	if count >= maxUserPins {
		return ErrUserPinLimitExceeded
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
		summary.Author = toPublicUserInfo(post.Author)
	}
	return summary
}
