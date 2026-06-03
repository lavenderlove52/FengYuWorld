package markdown

import (
	"strings"
	"testing"
)

func TestRenderSanitizesUnsafeHTML(t *testing.T) {
	source := `[safe](https://example.com)

<script>alert(1)</script>
<img src=x onerror=alert(1)>`

	got, err := NewRenderer().Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(got, `<script`) || strings.Contains(got, `onerror`) {
		t.Fatalf("expected unsafe HTML to be removed, got: %s", got)
	}
	if !strings.Contains(got, `href="https://example.com"`) {
		t.Fatalf("expected safe link to remain, got: %s", got)
	}
}
