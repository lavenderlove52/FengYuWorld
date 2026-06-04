package service

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
)

type fakeHomeThemeImageStore struct {
	count       int64
	images      []domain.HomeThemeImage
	found       *domain.HomeThemeImage
	createErr   error
	deleteErr   error
	created     *domain.HomeThemeImage
	deleted     *domain.HomeThemeImage
	createLimit int64
}

func (f *fakeHomeThemeImageStore) CountByUserID(int64) (int64, error) {
	return f.count, nil
}

func (f *fakeHomeThemeImageStore) ListByUserID(int64) ([]domain.HomeThemeImage, error) {
	return f.images, nil
}

func (f *fakeHomeThemeImageStore) CreateWithLimit(image *domain.HomeThemeImage, limit int64) error {
	f.created = image
	f.createLimit = limit
	if f.createErr == nil {
		image.ID = 7
	}
	return f.createErr
}

func (f *fakeHomeThemeImageStore) FindByIDAndUserID(int64, int64) (*domain.HomeThemeImage, error) {
	return f.found, nil
}

func (f *fakeHomeThemeImageStore) DeleteAndRemoveThemeReference(image *domain.HomeThemeImage) error {
	f.deleted = image
	return f.deleteErr
}

type fakeHomeThemeUploadManager struct {
	saved       bool
	releasedIDs []int64
	releaseErr  error
}

func (f *fakeHomeThemeUploadManager) SaveThemeImageVariants(int64, string, string, int64, io.Reader) (*ThemeImageUploadVariants, error) {
	f.saved = true
	return &ThemeImageUploadVariants{
		Desktop: dto.UploadResponse{ReferenceID: 99, URL: "/uploads/theme.jpg"},
		Mobile:  dto.UploadResponse{ReferenceID: 100, URL: "/uploads/theme-mobile.jpg"},
	}, nil
}

func (f *fakeHomeThemeUploadManager) Release(_ int64, referenceID int64) error {
	f.releasedIDs = append(f.releasedIDs, referenceID)
	return f.releaseErr
}

func TestHomeThemeImageUploadEnforcesLimitBeforeSaving(t *testing.T) {
	store := &fakeHomeThemeImageStore{count: homeThemeImageLimit}
	uploads := &fakeHomeThemeUploadManager{}
	svc := NewHomeThemeImageService(store, uploads)

	_, err := svc.Upload(3, "hero.jpg", 100, bytes.NewReader([]byte("image")))
	if !errors.Is(err, ErrHomeThemeImageLimit) {
		t.Fatalf("expected ErrHomeThemeImageLimit, got %v", err)
	}
	if uploads.saved {
		t.Fatal("upload was saved after gallery reached its limit")
	}
}

func TestHomeThemeImageUploadCreatesDedicatedRecord(t *testing.T) {
	store := &fakeHomeThemeImageStore{}
	uploads := &fakeHomeThemeUploadManager{}
	svc := NewHomeThemeImageService(store, uploads)

	image, err := svc.Upload(3, "hero.jpg", 100, bytes.NewReader([]byte("image")))
	if err != nil {
		t.Fatal(err)
	}
	if store.created == nil || store.created.UserID != 3 || store.created.UploadID != 99 ||
		store.created.MobileUploadID == nil || *store.created.MobileUploadID != 100 {
		t.Fatalf("unexpected created image: %#v", store.created)
	}
	if store.createLimit != homeThemeImageLimit {
		t.Fatalf("unexpected create limit: %d", store.createLimit)
	}
	if image.ID != 7 || image.URL != "/uploads/theme.jpg" || image.MobileURL != "/uploads/theme-mobile.jpg" {
		t.Fatalf("unexpected response image: %#v", image)
	}
}

func TestHomeThemeImageUploadReleasesBothVariantsWhenRecordCreationFails(t *testing.T) {
	createErr := errors.New("create failed")
	store := &fakeHomeThemeImageStore{createErr: createErr}
	uploads := &fakeHomeThemeUploadManager{}
	svc := NewHomeThemeImageService(store, uploads)

	_, err := svc.Upload(3, "hero.jpg", 100, bytes.NewReader([]byte("image")))
	if !errors.Is(err, createErr) {
		t.Fatalf("expected create error, got %v", err)
	}
	if len(uploads.releasedIDs) != 2 || uploads.releasedIDs[0] != 99 || uploads.releasedIDs[1] != 100 {
		t.Fatalf("expected both variants to be released, got %v", uploads.releasedIDs)
	}
}

func TestHomeThemeImageDeleteChecksOwnershipAndReleasesReference(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		store := &fakeHomeThemeImageStore{}
		uploads := &fakeHomeThemeUploadManager{}
		svc := NewHomeThemeImageService(store, uploads)

		err := svc.Delete(3, 5)
		if !errors.Is(err, ErrHomeThemeImageNotFound) {
			t.Fatalf("expected not found, got %v", err)
		}
		if len(uploads.releasedIDs) != 0 {
			t.Fatal("released an upload not owned by the user")
		}
	})

	t.Run("release after record removal", func(t *testing.T) {
		mobileUploadID := int64(100)
		image := &domain.HomeThemeImage{ID: 5, UserID: 3, UploadID: 99, MobileUploadID: &mobileUploadID}
		store := &fakeHomeThemeImageStore{found: image}
		uploads := &fakeHomeThemeUploadManager{}
		svc := NewHomeThemeImageService(store, uploads)

		if err := svc.Delete(3, 5); err != nil {
			t.Fatal(err)
		}
		if store.deleted != image || len(uploads.releasedIDs) != 2 ||
			uploads.releasedIDs[0] != 99 || uploads.releasedIDs[1] != 100 {
			t.Fatalf("delete/release mismatch: deleted=%#v released=%v", store.deleted, uploads.releasedIDs)
		}
	})
}
