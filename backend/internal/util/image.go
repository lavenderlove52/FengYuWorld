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
	MaxImageBytes  = 1 * 1024 * 1024 // 1MB
	MaxShortSidePx = 1000
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

	img = resizeIfNeeded(img)

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
	return encodeJPEGAdaptive(img)
}

// resizeIfNeeded 若图片短边超过 MaxShortSidePx，则等比缩放。
func resizeIfNeeded(src image.Image) image.Image {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	short := w
	if h < w {
		short = h
	}
	if short <= MaxShortSidePx {
		return src
	}

	ratio := float64(MaxShortSidePx) / float64(short)
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

// encodeJPEGAdaptive 从质量 90 开始编码，每次降 10，直到 ≤ 1MB 或质量降至 40 强制接受。
func encodeJPEGAdaptive(img image.Image) ([]byte, string, error) {
	for q := 90; q > 40; q -= 10 {
		b, ext, err := encodeJPEG(img, q)
		if err != nil {
			return nil, "", err
		}
		if len(b) <= MaxImageBytes {
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
