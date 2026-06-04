package util

import (
	"bytes"
	"image"
	"image/jpeg"
	"testing"
)

func TestResizeIfNeededThemeBackgroundPreserves2KLandscape(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 3840, 2160))
	out := resizeIfNeeded(src, MaxThemeBackgroundShortSidePx)
	b := out.Bounds()
	if b.Dx() != 1920 || b.Dy() != 1080 {
		t.Fatalf("expected 1920x1080, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestResizeIfNeededDefaultStillCapsAt1000(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2000, 1500))
	out := resizeIfNeeded(src, MaxShortSidePx)
	b := out.Bounds()
	if b.Dx() != 1333 || b.Dy() != 1000 {
		t.Fatalf("expected 1333x1000, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestProcessThemeBackgroundImageAllowsLargerOutput(t *testing.T) {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
	processed, ext, err := ProcessThemeBackgroundImage(bytes.NewReader(buf.Bytes()), ".jpg")
	if err != nil {
		t.Fatal(err)
	}
	if ext != ".jpg" {
		t.Fatalf("expected .jpg, got %s", ext)
	}
	if len(processed) == 0 {
		t.Fatal("expected non-empty output")
	}
}
