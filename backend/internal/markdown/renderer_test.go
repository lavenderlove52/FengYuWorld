package markdown

import (
	"strings"
	"testing"
)

func TestRenderBlockMath(t *testing.T) {
	source := `$$
SamplePoint = \frac{1 + TimeSegment1}{Total_Tq}
$$`

	got, err := NewRenderer().Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(got, `class="katex"`) {
		t.Fatalf("expected katex markup, got: %s", got)
	}
	if strings.Contains(got, `\frac`) {
		t.Fatalf("expected rendered math, not raw LaTeX: %s", got)
	}
}
