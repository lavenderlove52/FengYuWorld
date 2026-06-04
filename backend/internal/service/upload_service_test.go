package service

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/storage"
)

type memoryUploadStore struct {
	nextAssetID int64
	nextRefID   int64
	assets      map[string]domain.UploadAsset
	refs        map[int64]domain.Upload
}

func newMemoryUploadStore() *memoryUploadStore {
	return &memoryUploadStore{
		assets: make(map[string]domain.UploadAsset),
		refs:   make(map[int64]domain.Upload),
	}
}

func (s *memoryUploadStore) Acquire(asset *domain.UploadAsset, ref *domain.Upload, materialize func() error) error {
	key := asset.ContentHash + ":" + asset.Variant
	if existing, ok := s.assets[key]; ok {
		*asset = existing
	} else {
		s.nextAssetID++
		asset.ID = s.nextAssetID
		s.assets[key] = *asset
	}
	if err := materialize(); err != nil {
		return err
	}
	for _, existing := range s.refs {
		if *existing.AssetID == asset.ID &&
			existing.UserID == ref.UserID &&
			*existing.Purpose == *ref.Purpose &&
			*existing.OwnerType == *ref.OwnerType &&
			*existing.OwnerKey == *ref.OwnerKey {
			*ref = existing
			return nil
		}
	}
	s.nextRefID++
	ref.ID = s.nextRefID
	ref.AssetID = &asset.ID
	ref.Path = asset.Path
	ref.URL = asset.URL
	ref.Mime = asset.Mime
	ref.SizeBytes = asset.SizeBytes
	s.refs[ref.ID] = *ref
	return nil
}

func (s *memoryUploadStore) Release(userID, referenceID int64, remove func(path string) error) error {
	ref, ok := s.refs[referenceID]
	if !ok || ref.UserID != userID {
		return repository.ErrUploadReferenceNotFound
	}
	delete(s.refs, referenceID)
	for _, remaining := range s.refs {
		if *remaining.AssetID == *ref.AssetID {
			return nil
		}
	}
	for key, asset := range s.assets {
		if asset.ID == *ref.AssetID {
			if err := remove(asset.Path); err != nil {
				return err
			}
			delete(s.assets, key)
			return nil
		}
	}
	return nil
}

func TestSaveImageWithPurposeKeepsDifferentVariantsIndependent(t *testing.T) {
	baseDir := t.TempDir()
	repo := newMemoryUploadStore()
	svc := NewUploadService(storage.NewLocalStorage(baseDir, "/uploads"), repo, 5)
	data := testPNG(t, 4, 3)

	post, err := svc.SaveImageWithPurpose(1, "same.png", int64(len(data)), bytes.NewReader(data), SaveImageOptions{
		Purpose:   ImagePurposePost,
		OwnerType: "post",
		OwnerKey:  "post-1",
		Variant:   ImageVariantDefault,
	})
	if err != nil {
		t.Fatal(err)
	}
	theme, err := svc.SaveThemeImage(1, "theme-1", "same.png", int64(len(data)), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	if post.URL == theme.URL {
		t.Fatalf("post and theme variants unexpectedly shared an asset: %q", post.URL)
	}
	if post.ReferenceID == theme.ReferenceID {
		t.Fatal("different purposes shared a logical reference")
	}
	if len(repo.assets) != 2 || len(repo.refs) != 2 {
		t.Fatalf("unexpected assets/references: %d/%d", len(repo.assets), len(repo.refs))
	}
	var postPath, themePath string
	for _, asset := range repo.assets {
		switch asset.Variant {
		case ImageVariantDefault:
			postPath = asset.Path
		case ImageVariantThemeBackground:
			themePath = asset.Path
		}
	}
	if postPath == "" || themePath == "" {
		t.Fatalf("missing variant assets: post=%q theme=%q", postPath, themePath)
	}
	if err := svc.Release(1, theme.ReferenceID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(baseDir, filepath.FromSlash(themePath))); !os.IsNotExist(err) {
		t.Fatalf("theme asset was not deleted after its reference was released: %v", err)
	}
	if _, err := os.Stat(filepath.Join(baseDir, filepath.FromSlash(postPath))); err != nil {
		t.Fatalf("post asset was deleted while its reference remained: %v", err)
	}
	if err := svc.Release(1, post.ReferenceID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(baseDir, filepath.FromSlash(postPath))); !os.IsNotExist(err) {
		t.Fatalf("post asset was not deleted after its reference was released: %v", err)
	}
}

func TestSaveImageAndAvatarUseDifferentVariants(t *testing.T) {
	repo := newMemoryUploadStore()
	svc := NewUploadService(storage.NewLocalStorage(t.TempDir(), "/uploads"), repo, 5)
	data := testPNG(t, 4, 3)

	imageResult, err := svc.SaveImage(1, "same.png", int64(len(data)), bytes.NewReader(data), 0)
	if err != nil {
		t.Fatal(err)
	}
	avatarResult, err := svc.SaveAvatar(1, "same.png", int64(len(data)), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if imageResult.URL == avatarResult.URL {
		t.Fatal("image and avatar variants reused the same physical asset")
	}
	if len(repo.assets) != 2 {
		t.Fatalf("expected two variant assets, got %d", len(repo.assets))
	}
}

func TestSaveImageUsesDecodedFormatForStableDeduplication(t *testing.T) {
	repo := newMemoryUploadStore()
	svc := NewUploadService(storage.NewLocalStorage(t.TempDir(), "/uploads"), repo, 5)
	data := testPNG(t, 4, 3)

	first, err := svc.SaveImage(1, "same.png", int64(len(data)), bytes.NewReader(data), 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.SaveImage(2, "same.jpg", int64(len(data)), bytes.NewReader(data), 0)
	if err != nil {
		t.Fatal(err)
	}
	if first.URL != second.URL {
		t.Fatalf("identical bytes with different filename extensions did not reuse asset: %q != %q", first.URL, second.URL)
	}
	if filepath.Ext(first.URL) != ".png" {
		t.Fatalf("asset extension did not follow decoded format: %q", first.URL)
	}
	if len(repo.assets) != 1 {
		t.Fatalf("expected one asset, got %d", len(repo.assets))
	}
}

func TestSaveImageTreatsDifferentOriginalBytesAsDifferentAssets(t *testing.T) {
	repo := newMemoryUploadStore()
	svc := NewUploadService(storage.NewLocalStorage(t.TempDir(), "/uploads"), repo, 5)
	firstData := testPNG(t, 4, 3)
	secondData := testPNG(t, 5, 3)

	first, err := svc.SaveImage(1, "image.png", int64(len(firstData)), bytes.NewReader(firstData), 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.SaveImage(1, "image.png", int64(len(secondData)), bytes.NewReader(secondData), 0)
	if err != nil {
		t.Fatal(err)
	}
	if first.URL == second.URL {
		t.Fatal("different original bytes reused the same physical asset")
	}
	if len(repo.assets) != 2 {
		t.Fatalf("expected two assets, got %d", len(repo.assets))
	}
}

func TestSaveImageRejectsInvalidImageContent(t *testing.T) {
	repo := newMemoryUploadStore()
	svc := NewUploadService(storage.NewLocalStorage(t.TempDir(), "/uploads"), repo, 5)
	data := []byte("not an image")

	_, err := svc.SaveImage(1, "fake.png", int64(len(data)), bytes.NewReader(data), 0)
	if !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("expected ErrInvalidUpload, got %v", err)
	}
	if len(repo.assets) != 0 {
		t.Fatal("invalid content created an asset")
	}
}

func firstAssetPath(repo *memoryUploadStore) string {
	for _, asset := range repo.assets {
		return asset.Path
	}
	return ""
}

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 30), G: uint8(y * 30), B: 120, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
