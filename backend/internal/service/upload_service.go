package service

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/storage"
	"my-blog/backend/internal/util"

	"github.com/google/uuid"
)

var (
	ErrInvalidUpload           = errors.New("invalid upload")
	ErrUploadReferenceNotFound = errors.New("upload reference not found")
)

type ImagePurpose string

const (
	ImagePurposePost   ImagePurpose = "post-image"
	ImagePurposeAvatar ImagePurpose = "avatar"
	ImagePurposeTheme  ImagePurpose = "theme-image"

	ImageVariantDefault         = "image-v1"
	ImageVariantAvatar          = "avatar-v1"
	ImageVariantThemeBackground = "theme-bg-v1"
	ImageVariantThemeDesktop    = "theme-bg-desktop-v2"
	ImageVariantThemeMobile     = "theme-bg-mobile-v1"
	ImageVariantPostDesktop     = "post-image-desktop-v2"
	ImageVariantPostMobile      = "post-image-mobile-v1"
)

var allowedMimes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/gif":  {},
}

type SaveImageOptions struct {
	Purpose   ImagePurpose
	OwnerType string
	OwnerKey  string
	Variant   string
	MaxSize   int64
}

type UploadReferenceStore interface {
	Acquire(asset *domain.UploadAsset, ref *domain.Upload, materialize func() error) error
	Release(userID, referenceID int64, remove func(path string) error) error
}

type UploadService struct {
	storage  *storage.LocalStorage
	uploads  UploadReferenceStore
	maxBytes int64
}

type ThemeImageUploadVariants struct {
	Desktop dto.UploadResponse
	Mobile  dto.UploadResponse
}

func NewUploadService(storage *storage.LocalStorage, uploads UploadReferenceStore, maxMB int64) *UploadService {
	return &UploadService{
		storage:  storage,
		uploads:  uploads,
		maxBytes: maxMB * 1024 * 1024,
	}
}

func (s *UploadService) SaveImage(userID int64, filename string, size int64, r io.Reader, maxSize int64) (*dto.UploadResponse, error) {
	_, _, validatedMaxSize, err := s.validateImage(filename, size, maxSize)
	if err != nil {
		return nil, err
	}
	original, err := io.ReadAll(io.LimitReader(r, validatedMaxSize+1))
	if err != nil || int64(len(original)) > validatedMaxSize || len(original) == 0 {
		return nil, ErrInvalidUpload
	}
	ownerKey := uuid.NewString()
	desktop, err := s.SaveImageWithPurpose(userID, filename, int64(len(original)), bytes.NewReader(original), SaveImageOptions{
		Purpose:   ImagePurposePost,
		OwnerType: "upload",
		OwnerKey:  ownerKey,
		Variant:   ImageVariantPostDesktop,
		MaxSize:   validatedMaxSize,
	})
	if err != nil {
		return nil, err
	}
	mobile, err := s.SaveImageWithPurpose(userID, filename, int64(len(original)), bytes.NewReader(original), SaveImageOptions{
		Purpose:   ImagePurposePost,
		OwnerType: "upload",
		OwnerKey:  ownerKey,
		Variant:   ImageVariantPostMobile,
		MaxSize:   validatedMaxSize,
	})
	if err != nil {
		return nil, errors.Join(err, s.Release(userID, desktop.ReferenceID))
	}
	desktop.MobileURL = mobile.URL
	return desktop, nil
}

func (s *UploadService) SaveAvatar(userID int64, filename string, size int64, r io.Reader) (*dto.UploadResponse, error) {
	return s.SaveImageWithPurpose(userID, filename, size, r, SaveImageOptions{
		Purpose:   ImagePurposeAvatar,
		OwnerType: "user",
		OwnerKey:  fmt.Sprintf("%d:%s", userID, uuid.NewString()),
		Variant:   ImageVariantAvatar,
		MaxSize:   2 * 1024 * 1024,
	})
}

// SaveThemeImage is the public entry point for the dedicated theme image gallery.
// ownerKey should be the gallery image ID so each gallery item owns an independent reference.
func (s *UploadService) SaveThemeImage(userID int64, ownerKey, filename string, size int64, r io.Reader) (*dto.UploadResponse, error) {
	return s.SaveImageWithPurpose(userID, filename, size, r, SaveImageOptions{
		Purpose:   ImagePurposeTheme,
		OwnerType: "theme-gallery",
		OwnerKey:  ownerKey,
		Variant:   ImageVariantThemeDesktop,
	})
}

func (s *UploadService) SaveThemeImageVariants(userID int64, ownerKey, filename string, size int64, r io.Reader) (*ThemeImageUploadVariants, error) {
	_, _, maxSize, err := s.validateImage(filename, size, 0)
	if err != nil {
		return nil, err
	}
	original, err := io.ReadAll(io.LimitReader(r, maxSize+1))
	if err != nil || int64(len(original)) > maxSize || len(original) == 0 {
		return nil, ErrInvalidUpload
	}

	desktop, err := s.SaveImageWithPurpose(userID, filename, int64(len(original)), bytes.NewReader(original), SaveImageOptions{
		Purpose:   ImagePurposeTheme,
		OwnerType: "theme-gallery",
		OwnerKey:  ownerKey,
		Variant:   ImageVariantThemeDesktop,
	})
	if err != nil {
		return nil, err
	}
	mobile, err := s.SaveImageWithPurpose(userID, filename, int64(len(original)), bytes.NewReader(original), SaveImageOptions{
		Purpose:   ImagePurposeTheme,
		OwnerType: "theme-gallery",
		OwnerKey:  ownerKey,
		Variant:   ImageVariantThemeMobile,
	})
	if err != nil {
		return nil, errors.Join(err, s.Release(userID, desktop.ReferenceID))
	}
	return &ThemeImageUploadVariants{Desktop: *desktop, Mobile: *mobile}, nil
}

// SaveBannerImage 主页 Banner 条图，与 Hero 背景共用 2K 级压缩策略。
func (s *UploadService) SaveBannerImage(userID int64, filename string, size int64, r io.Reader) (*dto.UploadResponse, error) {
	return s.SaveImageWithPurpose(userID, filename, size, r, SaveImageOptions{
		Purpose:   ImagePurposeTheme,
		OwnerType: "theme-banner",
		OwnerKey:  uuid.NewString(),
		Variant:   ImageVariantThemeBackground,
	})
}

// SaveImageWithPurpose hashes the original bytes and acquires an independently owned reference
// to the processed physical asset. Content with a different byte representation is never reused.
func (s *UploadService) SaveImageWithPurpose(userID int64, filename string, size int64, r io.Reader, opts SaveImageOptions) (*dto.UploadResponse, error) {
	_, _, maxSize, err := s.validateImage(filename, size, opts.MaxSize)
	if err != nil {
		return nil, err
	}
	if err := validateSaveImageOptions(opts); err != nil {
		return nil, err
	}

	original, err := io.ReadAll(io.LimitReader(r, maxSize+1))
	if err != nil || int64(len(original)) > maxSize || len(original) == 0 {
		return nil, ErrInvalidUpload
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(original))
	if err != nil {
		return nil, ErrInvalidUpload
	}
	ext, mimeType, err := canonicalImageType(format)
	if err != nil {
		return nil, err
	}
	filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ext
	contentHash := fmt.Sprintf("%x", sha256.Sum256(original))

	var processed []byte
	var newExt string
	switch opts.Variant {
	case ImageVariantDefault:
		processed, newExt, err = util.ProcessImage(bytes.NewReader(original), ext)
	case ImageVariantAvatar:
		processed, newExt, err = util.ProcessAvatar(bytes.NewReader(original))
	case ImageVariantThemeBackground:
		processed, newExt, err = util.ProcessThemeBackgroundImage(bytes.NewReader(original), ext)
	case ImageVariantThemeDesktop:
		processed, newExt, err = util.ProcessThemeBackgroundImage(bytes.NewReader(original), ext)
	case ImageVariantThemeMobile:
		processed, newExt, err = util.ProcessThemeMobileImage(bytes.NewReader(original), ext)
	case ImageVariantPostDesktop:
		processed, newExt, err = util.ProcessPostImage(bytes.NewReader(original))
	case ImageVariantPostMobile:
		processed, newExt, err = util.ProcessPostMobileImage(bytes.NewReader(original))
	}
	if err != nil || len(processed) == 0 {
		return nil, ErrInvalidUpload
	}
	if newExt != ext {
		filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + newExt
		if m := mime.TypeByExtension(newExt); m != "" {
			mimeType = m
		}
	}

	location, err := s.storage.ContentAddressedLocation(contentHash, opts.Variant, filename)
	if err != nil {
		return nil, err
	}
	purpose, ownerType, ownerKey := string(opts.Purpose), opts.OwnerType, opts.OwnerKey
	asset := &domain.UploadAsset{
		ContentHash:       contentHash,
		Variant:           opts.Variant,
		Path:              location.Path,
		URL:               location.URL,
		Mime:              mimeType,
		SizeBytes:         int64(len(processed)),
		OriginalSizeBytes: int64(len(original)),
	}
	ref := &domain.Upload{
		UserID:    userID,
		Purpose:   &purpose,
		OwnerType: &ownerType,
		OwnerKey:  &ownerKey,
	}
	if err := s.uploads.Acquire(asset, ref, func() error {
		return s.storage.Ensure(asset.Path, bytes.NewReader(processed))
	}); err != nil {
		return nil, err
	}
	return &dto.UploadResponse{ReferenceID: ref.ID, URL: ref.URL}, nil
}

// Release removes one logical reference. It removes the physical asset only after the final
// reference across all users and purposes is gone.
func (s *UploadService) Release(userID, referenceID int64) error {
	err := s.uploads.Release(userID, referenceID, s.storage.Delete)
	if errors.Is(err, repository.ErrUploadReferenceNotFound) {
		return ErrUploadReferenceNotFound
	}
	return err
}

func (s *UploadService) validateImage(filename string, size, maxSize int64) (string, string, int64, error) {
	if maxSize <= 0 {
		maxSize = s.maxBytes
	}
	if size <= 0 || size > maxSize {
		return "", "", 0, ErrInvalidUpload
	}

	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if _, ok := allowedMimes[mimeType]; !ok {
		return "", "", 0, ErrInvalidUpload
	}
	return ext, mimeType, maxSize, nil
}

func validateSaveImageOptions(opts SaveImageOptions) error {
	switch opts.Purpose {
	case ImagePurposePost, ImagePurposeAvatar, ImagePurposeTheme:
	default:
		return ErrInvalidUpload
	}
	if opts.OwnerType == "" || len(opts.OwnerType) > 50 || opts.OwnerKey == "" || len(opts.OwnerKey) > 200 {
		return ErrInvalidUpload
	}
	switch opts.Variant {
	case ImageVariantDefault, ImageVariantAvatar, ImageVariantThemeBackground, ImageVariantThemeDesktop, ImageVariantThemeMobile, ImageVariantPostDesktop, ImageVariantPostMobile:
		return nil
	default:
		return ErrInvalidUpload
	}
}

func canonicalImageType(format string) (string, string, error) {
	switch strings.ToLower(format) {
	case "jpeg":
		return ".jpg", "image/jpeg", nil
	case "png":
		return ".png", "image/png", nil
	case "gif":
		return ".gif", "image/gif", nil
	case "webp":
		return ".webp", "image/webp", nil
	default:
		return "", "", ErrInvalidUpload
	}
}
