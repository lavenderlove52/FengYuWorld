package markdown

import (
	"strings"
	"testing"
)

func TestRenderGFMTable(t *testing.T) {
	source := `比较合理的进场方式：

| 情况 | 操作 |
| --- | --- |
| 短线继续冲高 | 不追，容易买在情绪高点 |
| 回踩 5 日线附近企稳 | 可小仓位试 |
`

	got, err := NewRenderer().Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(got, "<table") {
		t.Fatalf("expected HTML table, got: %s", got)
	}
	if strings.Contains(got, "| 情况 |") {
		t.Fatalf("expected parsed table, not raw pipes: %s", got)
	}
}

func TestRenderGFMTableNoBlankLineBefore(t *testing.T) {
	source := `比较合理的进场方式：
| 情况 | 操作 |
| --- | --- |
| 短线继续冲高 | 不追，容易买在情绪高点 |
`

	got, err := NewRenderer().Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(got, "<table") {
		t.Fatalf("expected HTML table without blank line before table, got: %s", got)
	}
}

func TestRenderGFMTableWithoutSeparator(t *testing.T) {
	source := `| 情况 | 操作 |
| 短线继续冲高 | 不追，容易买在情绪高点 |
`

	got, err := NewRenderer().Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	t.Log(got)
	if strings.Contains(got, "<table") {
		t.Log("goldmark parsed as table without separator")
	} else {
		t.Log("goldmark did NOT parse as table without separator")
	}
}

func TestRenderIndentedTable(t *testing.T) {
	source := `比较合理的进场方式：

    | 情况 | 操作 |
    | --- | --- |
    | 短线继续冲高 | 不追 |
`

	got, err := NewRenderer().Render(source)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	t.Log(got)
}
