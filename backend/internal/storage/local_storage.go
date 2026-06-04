package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LocalStorage struct {
	baseDir    string
	publicBase string
}

type SaveResult struct {
	Path string
	URL  string
}

var (
	contentHashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)
	variantPattern     = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,49}$`)
)

func NewLocalStorage(baseDir, publicBase string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir, publicBase: publicBase}
}

func (s *LocalStorage) Save(userID int64, filename string, r io.Reader) (*SaveResult, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".bin"
	}

	now := time.Now()
	relDir := filepath.Join(
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
		fmt.Sprintf("%02d", now.Day()),
	)
	fullDir := filepath.Join(s.baseDir, relDir)
	if err := os.MkdirAll(fullDir, 0o755); err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%d_%s%s", userID, uuid.NewString(), ext)
	relPath := filepath.ToSlash(filepath.Join(relDir, name))
	fullPath := filepath.Join(s.baseDir, relDir, name)

	f, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return nil, err
	}

	url := strings.TrimRight(s.publicBase, "/") + "/" + relPath
	return &SaveResult{Path: relPath, URL: url}, nil
}

func (s *LocalStorage) ContentAddressedLocation(contentHash, variant, filename string) (*SaveResult, error) {
	if !contentHashPattern.MatchString(contentHash) || !variantPattern.MatchString(variant) {
		return nil, errors.New("invalid content-addressed storage key")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" || len(ext) > 10 || strings.ContainsAny(ext, `/\`) {
		ext = ".bin"
	}

	relPath := filepath.ToSlash(filepath.Join("objects", variant, contentHash[:2], contentHash+ext))
	url := strings.TrimRight(s.publicBase, "/") + "/" + relPath
	return &SaveResult{Path: relPath, URL: url}, nil
}

// Ensure atomically materializes a content-addressed file. Existing files are reused.
func (s *LocalStorage) Ensure(path string, r io.Reader) error {
	fullPath, err := s.resolve(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(fullPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(fullPath), ".upload-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, fullPath); err != nil {
		if _, statErr := os.Stat(fullPath); statErr == nil {
			return nil
		}
		return err
	}
	return nil
}

func (s *LocalStorage) Delete(path string) error {
	fullPath, err := s.resolve(path)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) resolve(path string) (string, error) {
	base, err := filepath.Abs(s.baseDir)
	if err != nil {
		return "", err
	}
	fullPath, err := filepath.Abs(filepath.Join(base, filepath.FromSlash(path)))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(base, fullPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", errors.New("storage path escapes base directory")
	}
	return fullPath, nil
}
