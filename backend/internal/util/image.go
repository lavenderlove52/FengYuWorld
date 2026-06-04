package util

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	_ "image/gif"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	MaxImageBytes  = 1 * 1024 * 1024 // 1MB，正文/封面等通用图
	MaxShortSidePx = 1000

	MaxPostImageShortSidePx = 1080
	MaxPostImageBytes       = 2 * 1024 * 1024 // 2MB

	// 主页 Hero / Banner 等全屏背景：短边最高 1080（如 1920×1080）
	MaxThemeBackgroundShortSidePx = 1080
	MaxThemeBackgroundImageBytes  = 2 * 1024 * 1024 // 2MB
	MaxThemeMobileLongSidePx      = 1280
	MaxThemeMobileImageBytes      = 1 * 1024 * 1024 // 1MB

	AvatarSizePx = 200
)

// ProcessImage 对上传的图片进行缩放和压缩，返回处理后的字节和推荐扩展名（".jpg" 或 ".png"）。
// 若解码失败则原样返回，由上层决定是否拒绝。
func ProcessImage(r io.Reader, originalExt string) ([]byte, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		// 无法解码（如 GIF 动图等），直接返回原始数据
		return data, originalExt, nil
	}

	img = resizeIfNeeded(img, MaxShortSidePx)

	ext := strings.ToLower(originalExt)
	if ext == ".png" {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		if buf.Len() <= MaxImageBytes {
			return buf.Bytes(), ".png", nil
		}
		// PNG 无损压缩后仍超限，转 JPEG
		return encodeJPEG(img, 85)
	}

	// JPEG、WebP 或其他格式 → JPEG，逐步降质
	return encodeJPEGAdaptive(img, MaxImageBytes)
}

// ProcessPostImage generates the high-quality desktop variant used by article content and covers.
func ProcessPostImage(r io.Reader) ([]byte, string, error) {
	return processResponsivePostImage(r, MaxPostImageShortSidePx, 0, MaxPostImageBytes, 60, 92)
}

// ProcessPostMobileImage generates the mobile article variant with a 1280px longest edge.
func ProcessPostMobileImage(r io.Reader) ([]byte, string, error) {
	return processResponsivePostImage(r, 0, MaxThemeMobileLongSidePx, MaxThemeMobileImageBytes, 60, 90)
}

// ProcessThemeBackgroundImage 用于主页 Hero 轮播与 Banner，保留更高分辨率（2K 级）。
func ProcessThemeBackgroundImage(r io.Reader, originalExt string) ([]byte, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, originalExt, nil
	}

	img = resizeIfNeeded(img, MaxThemeBackgroundShortSidePx)

	ext := strings.ToLower(originalExt)
	if ext == ".png" {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		if buf.Len() <= MaxThemeBackgroundImageBytes {
			return buf.Bytes(), ".png", nil
		}
		return encodeJPEGAdaptiveQuality(img, MaxThemeBackgroundImageBytes, 60, 92)
	}

	return encodeJPEGAdaptiveQuality(img, MaxThemeBackgroundImageBytes, 60, 92)
}

// ProcessThemeMobileImage generates a smaller hero image for mobile viewports.
func ProcessThemeMobileImage(r io.Reader, originalExt string) ([]byte, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, originalExt, nil
	}

	img = resizeLongestSideIfNeeded(img, MaxThemeMobileLongSidePx)

	ext := strings.ToLower(originalExt)
	if ext == ".png" {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		if buf.Len() <= MaxThemeMobileImageBytes {
			return buf.Bytes(), ".png", nil
		}
	}

	return encodeJPEGAdaptiveQuality(img, MaxThemeMobileImageBytes, 60, 90)
}

func processResponsivePostImage(r io.Reader, maxShortSide, maxLongSide, maxBytes, minQuality, maxQuality int) ([]byte, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}
	if maxShortSide > 0 {
		img = resizeIfNeeded(img, maxShortSide)
	}
	if maxLongSide > 0 {
		img = resizeLongestSideIfNeeded(img, maxLongSide)
	}
	if !imageIsOpaque(img) {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), ".png", nil
	}
	return encodeJPEGAdaptiveQuality(img, maxBytes, minQuality, maxQuality)
}

func imageIsOpaque(img image.Image) bool {
	if opaque, ok := img.(interface{ Opaque() bool }); ok {
		return opaque.Opaque()
	}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha != 0xffff {
				return false
			}
		}
	}
	return true
}

// ProcessAvatar 将头像处理为固定 200x200 的 JPEG。
func ProcessAvatar(r io.Reader) ([]byte, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, "", err
	}

	img = cropToSquare(img)

	dst := image.NewRGBA(image.Rect(0, 0, AvatarSizePx, AvatarSizePx))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return encodeJPEG(dst, 90)
}

// resizeIfNeeded 若图片短边超过 maxShortSidePx，则等比缩放。
func resizeIfNeeded(src image.Image, maxShortSidePx int) image.Image {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	short := w
	if h < w {
		short = h
	}
	if short <= maxShortSidePx {
		return src
	}

	ratio := float64(maxShortSidePx) / float64(short)
	newW := int(float64(w) * ratio)
	newH := int(float64(h) * ratio)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func resizeLongestSideIfNeeded(src image.Image, maxLongSidePx int) image.Image {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	longSide := w
	if h > w {
		longSide = h
	}
	if longSide <= maxLongSidePx {
		return src
	}

	ratio := float64(maxLongSidePx) / float64(longSide)
	newW := max(1, int(float64(w)*ratio))
	newH := max(1, int(float64(h)*ratio))
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func cropToSquare(src image.Image) image.Image {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	side := w
	if h < w {
		side = h
	}
	x0 := bounds.Min.X + (w-side)/2
	y0 := bounds.Min.Y + (h-side)/2
	return cropImage(src, image.Rect(x0, y0, x0+side, y0+side))
}

func cropImage(src image.Image, rect image.Rectangle) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(dst, dst.Bounds(), src, rect.Min, draw.Src)
	return dst
}

// encodeJPEGAdaptive 从质量 90 开始编码，每次降 10，直到 ≤ maxBytes 或质量降至 40 强制接受。
func encodeJPEGAdaptive(img image.Image, maxBytes int) ([]byte, string, error) {
	for q := 90; q > 40; q -= 10 {
		b, ext, err := encodeJPEG(img, q)
		if err != nil {
			return nil, "", err
		}
		if len(b) <= maxBytes {
			return b, ext, nil
		}
	}
	// 质量 40，强制接受
	return encodeJPEG(img, 40)
}

func encodeJPEGAdaptiveQuality(img image.Image, maxBytes, minQuality, maxQuality int) ([]byte, string, error) {
	var best []byte
	low, high := minQuality, maxQuality
	for low <= high {
		quality := low + (high-low)/2
		encoded, _, err := encodeJPEG(img, quality)
		if err != nil {
			return nil, "", err
		}
		if len(encoded) <= maxBytes {
			best = encoded
			low = quality + 1
		} else {
			high = quality - 1
		}
	}
	if best != nil {
		return best, ".jpg", nil
	}
	return encodeJPEG(img, minQuality)
}

func encodeJPEG(img image.Image, quality int) ([]byte, string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), ".jpg", nil
}
