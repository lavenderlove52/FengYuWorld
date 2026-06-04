package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestContentAddressedLocationIncludesVariant(t *testing.T) {
	store := NewLocalStorage(t.TempDir(), "/uploads")
	hash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	image, err := store.ContentAddressedLocation(hash, "image", "photo.png")
	if err != nil {
		t.Fatal(err)
	}
	avatar, err := store.ContentAddressedLocation(hash, "avatar", "photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if image.Path == avatar.Path {
		t.Fatalf("different variants resolved to the same path: %s", image.Path)
	}
	if image.URL != "/uploads/objects/image/aa/"+hash+".png" {
		t.Fatalf("unexpected URL: %s", image.URL)
	}
}

func TestEnsureReusesAndDeleteRemovesContentAddressedFile(t *testing.T) {
	baseDir := t.TempDir()
	store := NewLocalStorage(baseDir, "/uploads")
	path := "objects/image/aa/file.png"

	if err := store.Ensure(path, bytes.NewBufferString("first")); err != nil {
		t.Fatal(err)
	}
	if err := store.Ensure(path, bytes.NewBufferString("second")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(baseDir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first" {
		t.Fatalf("Ensure replaced existing content: %q", got)
	}
	if err := store.Delete(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(baseDir, filepath.FromSlash(path))); !os.IsNotExist(err) {
		t.Fatalf("file still exists after Delete: %v", err)
	}
}

func TestStorageRejectsEscapingPaths(t *testing.T) {
	store := NewLocalStorage(t.TempDir(), "/uploads")
	if err := store.Ensure("../outside", bytes.NewBufferString("bad")); err == nil {
		t.Fatal("Ensure accepted a path outside the storage root")
	}
	if err := store.Delete("../outside"); err == nil {
		t.Fatal("Delete accepted a path outside the storage root")
	}
}
