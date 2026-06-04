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

	// 主页 Hero / Banner 等全屏背景：短边最高 1080（如 1920×1080）
	MaxThemeBackgroundShortSidePx = 1080
	MaxThemeBackgroundImageBytes  = 2 * 1024 * 1024 // 2MB

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
		return encodeJPEG(img, 88)
	}

	return encodeJPEGAdaptive(img, MaxThemeBackgroundImageBytes)
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

func encodeJPEG(img image.Image, quality int) ([]byte, string, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), ".jpg", nil
}
