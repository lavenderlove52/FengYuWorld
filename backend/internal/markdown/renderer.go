package markdown

import (
	"bytes"
	"regexp"
	"strings"

	katex "github.com/FurqanSoftware/goldmark-katex"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	scriptRe  = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	onEventRe = regexp.MustCompile(`(?i)\s(on\w+)\s*=`)
)

type Renderer struct {
	md goldmark.Markdown
}

func NewRenderer() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, &katex.Extender{}),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()),
	)
	return &Renderer{md: md}
}

func (r *Renderer) Render(source string) (string, error) {
	var buf bytes.Buffer
	if err := r.md.Convert([]byte(source), &buf); err != nil {
		return "", err
	}
	return sanitizeHTML(buf.String()), nil
}

func sanitizeHTML(input string) string {
	out := scriptRe.ReplaceAllString(input, "")
	out = onEventRe.ReplaceAllString(out, "")
	return strings.TrimSpace(out)
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
