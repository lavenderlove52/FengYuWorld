package service

import (
	"errors"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
)

var ErrCategoryNotFound = errors.New("category not found")

type CategoryService struct {
	categories *repository.CategoryRepository
}

func NewCategoryService(categories *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categories: categories}
}

func (s *CategoryService) List(userID int64) ([]dto.CategoryResponse, error) {
	list, err := s.categories.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CategoryResponse, 0, len(list))
	for _, c := range list {
		out = append(out, dto.CategoryResponse{
			ID: c.ID, Name: c.Name, Icon: c.Icon, SortOrder: c.SortOrder,
		})
	}
	return out, nil
}

func (s *CategoryService) Create(userID int64, req dto.CategoryRequest) (*dto.CategoryResponse, error) {
	c := &domain.Category{
		UserID: userID, Name: req.Name, Icon: req.Icon, SortOrder: req.SortOrder,
	}
	if err := s.categories.Create(c); err != nil {
		return nil, err
	}
	return &dto.CategoryResponse{ID: c.ID, Name: c.Name, Icon: c.Icon, SortOrder: c.SortOrder}, nil
}

func (s *CategoryService) Update(userID, id int64, req dto.CategoryRequest) (*dto.CategoryResponse, error) {
	c, err := s.categories.FindByIDAndUser(id, userID)
	if err != nil || c == nil {
		return nil, ErrCategoryNotFound
	}
	c.Name = req.Name
	c.Icon = req.Icon
	c.SortOrder = req.SortOrder
	if err := s.categories.Update(c); err != nil {
		return nil, err
	}
	return &dto.CategoryResponse{ID: c.ID, Name: c.Name, Icon: c.Icon, SortOrder: c.SortOrder}, nil
}

func (s *CategoryService) Delete(userID, id int64) error {
	c, err := s.categories.FindByIDAndUser(id, userID)
	if err != nil || c == nil {
		return ErrCategoryNotFound
	}
	return s.categories.Delete(id, userID)
}
