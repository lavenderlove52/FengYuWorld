package service

import (
	"errors"
	"io"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"

	"github.com/google/uuid"
)

const (
	homeThemeImageLimit = int64(10)
)

var (
	ErrHomeThemeImageLimit    = errors.New("home theme image limit reached")
	ErrHomeThemeImageNotFound = errors.New("home theme image not found")
)

type homeThemeUploadManager interface {
	SaveThemeImage(userID int64, ownerKey, filename string, size int64, r io.Reader) (*dto.UploadResponse, error)
	Release(userID, referenceID int64) error
}

type homeThemeImageStore interface {
	CountByUserID(userID int64) (int64, error)
	ListByUserID(userID int64) ([]domain.HomeThemeImage, error)
	CreateWithLimit(image *domain.HomeThemeImage, limit int64) error
	FindByIDAndUserID(id, userID int64) (*domain.HomeThemeImage, error)
	DeleteAndRemoveThemeReference(image *domain.HomeThemeImage) error
}

type HomeThemeImageService struct {
	images  homeThemeImageStore
	uploads homeThemeUploadManager
}

func NewHomeThemeImageService(images homeThemeImageStore, uploads homeThemeUploadManager) *HomeThemeImageService {
	return &HomeThemeImageService{images: images, uploads: uploads}
}

func (s *HomeThemeImageService) List(userID int64) ([]domain.HomeThemeImage, error) {
	return s.images.ListByUserID(userID)
}

func (s *HomeThemeImageService) Upload(userID int64, filename string, size int64, r io.Reader) (*domain.HomeThemeImage, error) {
	count, err := s.images.CountByUserID(userID)
	if err != nil {
		return nil, err
	}
	if count >= homeThemeImageLimit {
		return nil, ErrHomeThemeImageLimit
	}

	ref, err := s.uploads.SaveThemeImage(userID, uuid.NewString(), filename, size, r)
	if err != nil {
		return nil, err
	}

	image := &domain.HomeThemeImage{
		UserID:   userID,
		UploadID: ref.ReferenceID,
		URL:      ref.URL,
	}
	if err := s.images.CreateWithLimit(image, homeThemeImageLimit); err != nil {
		releaseErr := s.uploads.Release(userID, ref.ReferenceID)
		if errors.Is(err, repository.ErrHomeThemeImageLimit) {
			return nil, ErrHomeThemeImageLimit
		}
		if releaseErr != nil {
			return nil, errors.Join(err, releaseErr)
		}
		return nil, err
	}
	return image, nil
}

func (s *HomeThemeImageService) Delete(userID, imageID int64) error {
	image, err := s.images.FindByIDAndUserID(imageID, userID)
	if err != nil {
		return err
	}
	if image == nil {
		return ErrHomeThemeImageNotFound
	}
	if err := s.images.DeleteAndRemoveThemeReference(image); err != nil {
		return err
	}
	return s.uploads.Release(userID, image.UploadID)
}
