package service

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"path/filepath"
	"strings"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/storage"
	"my-blog/backend/internal/util"
)

var ErrInvalidUpload = errors.New("invalid upload")

var allowedMimes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/gif":  {},
}

type UploadService struct {
	storage  *storage.LocalStorage
	uploads  *repository.UploadRepository
	maxBytes int64
}

func NewUploadService(storage *storage.LocalStorage, uploads *repository.UploadRepository, maxMB int64) *UploadService {
	return &UploadService{
		storage:  storage,
		uploads:  uploads,
		maxBytes: maxMB * 1024 * 1024,
	}
}

func (s *UploadService) SaveImage(userID int64, filename string, size int64, r io.Reader, maxSize int64) (*dto.UploadResponse, error) {
	ext, mimeType, err := s.validateImage(filename, size, maxSize)
	if err != nil {
		return nil, err
	}
	processed, newExt, err := util.ProcessImage(r, ext)
	if err != nil {
		return nil, err
	}

	// 若扩展名发生变化（如 WebP → .jpg），同步更新 filename 和 mimeType
	if newExt != ext {
		base := strings.TrimSuffix(filename, filepath.Ext(filename))
		filename = base + newExt
		if m := mime.TypeByExtension(newExt); m != "" {
			mimeType = m
		}
	}

	return s.saveProcessed(userID, filename, mimeType, processed)
}

func (s *UploadService) SaveAvatar(userID int64, filename string, size int64, r io.Reader) (*dto.UploadResponse, error) {
	const avatarMax = 2 * 1024 * 1024
	if _, _, err := s.validateImage(filename, size, avatarMax); err != nil {
		return nil, err
	}
	processed, newExt, err := util.ProcessAvatar(r)
	if err != nil {
		return nil, ErrInvalidUpload
	}
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	filename = base + newExt
	return s.saveProcessed(userID, filename, "image/jpeg", processed)
}

func (s *UploadService) validateImage(filename string, size int64, maxSize int64) (string, string, error) {
	if maxSize <= 0 {
		maxSize = s.maxBytes
	}
	if size > maxSize {
		return "", "", ErrInvalidUpload
	}

	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if _, ok := allowedMimes[mimeType]; !ok {
		return "", "", ErrInvalidUpload
	}
	return ext, mimeType, nil
}

func (s *UploadService) saveProcessed(userID int64, filename, mimeType string, processed []byte) (*dto.UploadResponse, error) {
	result, err := s.storage.Save(userID, filename, bytes.NewReader(processed))
	if err != nil {
		return nil, err
	}

	upload := &domain.Upload{
		UserID:    userID,
		Path:      result.Path,
		URL:       result.URL,
		Mime:      mimeType,
		SizeBytes: int64(len(processed)),
	}
	if err := s.uploads.Create(upload); err != nil {
		return nil, err
	}

	return &dto.UploadResponse{URL: result.URL}, nil
}
