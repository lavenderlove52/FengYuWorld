package util

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
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

func TestProcessThemeMobileImageCapsLongestSideAt1280(t *testing.T) {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 3840, 2160))
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	processed, _, err := ProcessThemeMobileImage(bytes.NewReader(buf.Bytes()), ".png")
	if err != nil {
		t.Fatal(err)
	}
	out, _, err := image.Decode(bytes.NewReader(processed))
	if err != nil {
		t.Fatal(err)
	}
	bounds := out.Bounds()
	if bounds.Dx() != 1280 || bounds.Dy() != 720 {
		t.Fatalf("expected 1280x720, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestResizeLongestSideIfNeededPreservesSmallImage(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 800, 1200))
	out := resizeLongestSideIfNeeded(src, MaxThemeMobileLongSidePx)
	if out != src {
		t.Fatal("small mobile image was unnecessarily resized")
	}
}

func TestProcessPostImageVariantsUseExpectedDimensions(t *testing.T) {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 3840, 2160))
	for y := 0; y < 2160; y++ {
		for x := 0; x < 3840; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 30, G: 60, B: 90, A: 255})
		}
	}
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	desktopBytes, desktopExt, err := ProcessPostImage(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	mobileBytes, mobileExt, err := ProcessPostMobileImage(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if desktopExt != ".jpg" || mobileExt != ".jpg" {
		t.Fatalf("expected JPEG variants, got desktop=%s mobile=%s", desktopExt, mobileExt)
	}

	desktop, _, err := image.Decode(bytes.NewReader(desktopBytes))
	if err != nil {
		t.Fatal(err)
	}
	mobile, _, err := image.Decode(bytes.NewReader(mobileBytes))
	if err != nil {
		t.Fatal(err)
	}
	if desktop.Bounds().Dx() != 1920 || desktop.Bounds().Dy() != 1080 {
		t.Fatalf("expected 1920x1080 desktop image, got %dx%d", desktop.Bounds().Dx(), desktop.Bounds().Dy())
	}
	if mobile.Bounds().Dx() != 1280 || mobile.Bounds().Dy() != 720 {
		t.Fatalf("expected 1280x720 mobile image, got %dx%d", mobile.Bounds().Dx(), mobile.Bounds().Dy())
	}
}

func TestProcessPostImagePreservesTransparency(t *testing.T) {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	img.SetRGBA(10, 10, color.RGBA{R: 255, A: 128})
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	desktop, desktopExt, err := ProcessPostImage(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	_, mobileExt, err := ProcessPostMobileImage(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if desktopExt != ".png" || mobileExt != ".png" {
		t.Fatalf("transparent image lost PNG format: desktop=%s mobile=%s", desktopExt, mobileExt)
	}
	decoded, _, err := image.Decode(bytes.NewReader(desktop))
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, alpha := decoded.At(10, 10).RGBA()
	if alpha == 0 || alpha == 0xffff {
		t.Fatalf("transparent alpha was not preserved: %d", alpha)
	}
}
