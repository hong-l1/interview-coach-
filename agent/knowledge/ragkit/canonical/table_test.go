package canonical

import "testing"

func TestNormalizeMarkdownPipeTablesDetectsAndRenders(t *testing.T) {
	raw := "intro\n\n| 名称 | 值 |\n| --- | --- |\n| a | 1 |\n| b | 2 |\n\ntail"
	doc := NewNormalizedDocument(raw, Source{FileName: "t.md"})
	rules := NormalizeMarkdownPipeTables(doc)
	if len(doc.Tables) != 1 {
		t.Fatalf("want 1 table, got %d", len(doc.Tables))
	}
	tbl := doc.Tables[0]
	if len(tbl.Rows) != 3 {
		t.Fatalf("want 3 rows (header+2), got %d", len(tbl.Rows))
	}
	if tbl.Rows[0][0] != "名称" || tbl.Rows[2][1] != "2" {
		t.Fatalf("unexpected rows: %v", tbl.Rows)
	}
	if tbl.MarkdownStart <= 0 || tbl.MarkdownEnd <= tbl.MarkdownStart {
		t.Fatalf("bad span: %d..%d", tbl.MarkdownStart, tbl.MarkdownEnd)
	}
	if len(rules) == 0 {
		t.Fatalf("want applied rule recorded")
	}
	// 规范渲染仍包含表格且边界落在 raw 内
	rendered := doc.ContentMarkdown
	if !contains(rendered, "| 名称 | 值 |") {
		t.Fatalf("rendered missing header: %q", rendered)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && indexOf(s, sub) >= 0 }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
