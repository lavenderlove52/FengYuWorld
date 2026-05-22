package service

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
)

var ErrTagNotFound = errors.New("tag not found")

type TagService struct {
	tags *repository.TagRepository
}

func NewTagService(tags *repository.TagRepository) *TagService {
	return &TagService{tags: tags}
}

func (s *TagService) List(userID int64) ([]dto.TagResponse, error) {
	list, err := s.tags.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.TagResponse, 0, len(list))
	for _, t := range list {
		out = append(out, dto.TagResponse{ID: t.ID, Name: t.Name})
	}
	return out, nil
}

func (s *TagService) Create(userID int64, req dto.TagRequest) (*dto.TagResponse, error) {
	t := &domain.Tag{UserID: userID, Name: req.Name}
	if err := s.tags.Create(t); err != nil {
		return nil, err
	}
	return &dto.TagResponse{ID: t.ID, Name: t.Name}, nil
}

func (s *TagService) Delete(userID, id int64) error {
	t, err := s.tags.FindByIDAndUser(id, userID)
	if err != nil || t == nil {
		return ErrTagNotFound
	}
	return s.tags.Delete(id, userID)
}
