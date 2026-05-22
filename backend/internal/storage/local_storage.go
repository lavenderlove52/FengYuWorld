package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
