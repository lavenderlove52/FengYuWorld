package service

import (
	"errors"
	"fmt"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/util"
)

var (
	ErrFriendNotFound   = errors.New("friend not found")
	ErrFriendConflict   = errors.New("friend conflict")
	ErrFriendForbidden  = errors.New("friend forbidden")
	ErrCannotFriendSelf = errors.New("cannot friend self")
	// 下面几个是 ErrFriendConflict 的具体情况，前端要显示不同提示
	ErrAlreadyFriends   = fmt.Errorf("already friends: %w", ErrFriendConflict)
	ErrRequestPending   = fmt.Errorf("request pending: %w", ErrFriendConflict)
	ErrRequestIncoming  = fmt.Errorf("request incoming: %w", ErrFriendConflict)
)

type FriendshipService struct {
	friendships *repository.FriendshipRepository
	users       *repository.UserRepository
	// plans 可选：取消好友时顺手清掉「指定可见」授权
	plans *repository.GridPlanRepository
}

func NewFriendshipService(friendships *repository.FriendshipRepository, users *repository.UserRepository) *FriendshipService {
	return &FriendshipService{friendships: friendships, users: users}
}

// SetGridPlans 注入计划仓储，让取消好友同时回收「指定可见」授权。
func (s *FriendshipService) SetGridPlans(plans *repository.GridPlanRepository) {
	s.plans = plans
}

func (s *FriendshipService) SendRequest(userID int64, req dto.FriendRequestCreate) (*dto.FriendshipItem, error) {
	target, err := s.users.FindByUsername(req.FriendUsername)
	if err != nil || target == nil {
		return nil, ErrFriendNotFound
	}
	if target.ID == userID {
		return nil, ErrCannotFriendSelf
	}

	existing, err := s.friendships.FindPair(userID, target.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		switch existing.Status {
		case domain.FriendshipStatusPending:
			if existing.FriendID == userID {
				// 对方先发的申请，我没处理又去点添加
				return nil, ErrRequestIncoming
			}
			return nil, ErrRequestPending
		case domain.FriendshipStatusAccepted:
			return nil, ErrAlreadyFriends
		case domain.FriendshipStatusRejected:
			// 之前被拒绝过：复用同一行改回 pending。
			// 直接 Create 会撞 (user_id, friend_id) 唯一键，用户看到的是「系统错误」，永远加不上。
			if err := s.friendships.Reopen(userID, target.ID, existing.ID); err != nil {
				return nil, err
			}
			existing.UserID = userID
			existing.FriendID = target.ID
			existing.Status = domain.FriendshipStatusPending
			return s.toItem(existing, target), nil
		}
	}

	f := &domain.Friendship{
		UserID:   userID,
		FriendID: target.ID,
		Status:   domain.FriendshipStatusPending,
	}
	if err := s.friendships.Create(f); err != nil {
		return nil, err
	}
	return s.toItem(f, target), nil
}

func (s *FriendshipService) ListReceived(userID int64) ([]dto.FriendshipItem, error) {
	list, err := s.friendships.ListReceived(userID)
	if err != nil {
		return nil, err
	}
	return s.mapItems(list, true)
}

func (s *FriendshipService) ListSent(userID int64) ([]dto.FriendshipItem, error) {
	list, err := s.friendships.ListSent(userID)
	if err != nil {
		return nil, err
	}
	return s.mapItems(list, false)
}

func (s *FriendshipService) Accept(userID, requestID int64) error {
	f, err := s.friendships.FindByID(requestID)
	if err != nil || f == nil {
		return ErrFriendNotFound
	}
	if f.FriendID != userID || f.Status != domain.FriendshipStatusPending {
		return ErrFriendForbidden
	}
	f.Status = domain.FriendshipStatusAccepted
	return s.friendships.Update(f)
}

func (s *FriendshipService) Reject(userID, requestID int64) error {
	f, err := s.friendships.FindByID(requestID)
	if err != nil || f == nil {
		return ErrFriendNotFound
	}
	if f.FriendID != userID || f.Status != domain.FriendshipStatusPending {
		return ErrFriendForbidden
	}
	f.Status = domain.FriendshipStatusRejected
	return s.friendships.Update(f)
}

func (s *FriendshipService) ListFriends(userID int64) ([]dto.FriendListItem, error) {
	list, err := s.friendships.ListFriends(userID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.FriendListItem, 0, len(list))
	for _, f := range list {
		otherID := f.FriendID
		if otherID == userID {
			otherID = f.UserID
		}
		user, err := s.users.FindByID(otherID)
		if err != nil || user == nil {
			continue
		}
		result = append(result, dto.FriendListItem{
			UserInfo:     toUserInfo(user, false),
			FriendsSince: util.FormatTime(f.UpdatedAt),
		})
	}
	return result, nil
}

func (s *FriendshipService) DeleteFriend(userID, friendUserID int64) error {
	ok, err := s.friendships.AreFriends(userID, friendUserID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrFriendNotFound
	}
	if err := s.friendships.DeletePair(userID, friendUserID); err != nil {
		return err
	}
	// 取消好友后，之前「指定可见」给对方的计划要立刻失效，不能等到下次编辑计划才清
	if s.plans != nil {
		if err := s.plans.PurgeVisibleUsers(userID, friendUserID); err != nil {
			return err
		}
	}
	return nil
}

func (s *FriendshipService) mapItems(list []domain.Friendship, received bool) ([]dto.FriendshipItem, error) {
	items := make([]dto.FriendshipItem, 0, len(list))
	for _, f := range list {
		otherID := f.UserID
		if received {
			otherID = f.UserID
		} else {
			otherID = f.FriendID
		}
		user, err := s.users.FindByID(otherID)
		if err != nil || user == nil {
			continue
		}
		items = append(items, *s.toItem(&f, user))
	}
	return items, nil
}

func (s *FriendshipService) toItem(f *domain.Friendship, user *domain.User) *dto.FriendshipItem {
	return &dto.FriendshipItem{
		ID:        f.ID,
		Status:    string(f.Status),
		CreatedAt: util.FormatTime(f.CreatedAt),
		User:      toUserInfo(user, false),
	}
}
