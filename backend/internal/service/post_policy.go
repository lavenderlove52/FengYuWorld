package service

import (
	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"
)

type PostPolicy struct {
	friendships *repository.FriendshipRepository
	posts       *repository.PostRepository
}

func NewPostPolicy(friendships *repository.FriendshipRepository, posts *repository.PostRepository) *PostPolicy {
	return &PostPolicy{friendships: friendships, posts: posts}
}

func (p *PostPolicy) CanReadPost(viewerID *int64, post *domain.Post) (bool, error) {
	if post.IsDeleted {
		if viewerID != nil && *viewerID == post.AuthorID {
			return true, nil
		}
		return false, nil
	}

	if post.Status == domain.PostStatusDraft {
		return viewerID != nil && *viewerID == post.AuthorID, nil
	}

	if viewerID != nil && *viewerID == post.AuthorID {
		return true, nil
	}

	// 非作者只能看已发布且未删除的文章
	if post.Status != domain.PostStatusPublished || post.IsDeleted {
		return false, nil
	}

	switch post.Visibility {
	case domain.PostVisibilityPublic:
		return true, nil
	case domain.PostVisibilityPrivate:
		return false, nil
	case domain.PostVisibilityFriends:
		if viewerID == nil {
			return false, nil
		}
		return p.friendships.AreFriends(post.AuthorID, *viewerID)
	case domain.PostVisibilitySpecified:
		if viewerID == nil {
			return false, nil
		}
		return p.posts.IsVisibleToUser(post.ID, *viewerID)
	default:
		return false, nil
	}
}
