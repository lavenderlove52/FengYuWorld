package markdown

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type Renderer struct {
	md        goldmark.Markdown
	sanitizer *bluemonday.Policy
}

func NewRenderer() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()),
	)
	return &Renderer{md: md, sanitizer: bluemonday.UGCPolicy()}
}

func (r *Renderer) Render(source string) (string, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(source), &buf); err != nil {
		return "", err
	}
	return strings.TrimSpace(r.sanitizer.Sanitize(buf.String())), nil
}

func Summary(html string, maxLen int) string {
	text := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(html, "")
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}
