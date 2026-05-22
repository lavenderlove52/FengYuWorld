package service

import (
	"errors"

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
)

type FriendshipService struct {
	friendships *repository.FriendshipRepository
	users       *repository.UserRepository
}

func NewFriendshipService(friendships *repository.FriendshipRepository, users *repository.UserRepository) *FriendshipService {
	return &FriendshipService{friendships: friendships, users: users}
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
				return nil, ErrFriendConflict
			}
			return nil, ErrFriendConflict
		case domain.FriendshipStatusAccepted:
			return nil, ErrFriendConflict
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
	return s.friendships.DeletePair(userID, friendUserID)
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
