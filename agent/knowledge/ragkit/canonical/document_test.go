package canonical

import "testing"

func TestNewNormalizedDocumentPreservesRaw(t *testing.T) {
	doc := NewNormalizedDocument("# Title\r\n", Source{FileName: "a.md", FileType: "md"})
	if doc.ContentMarkdownRaw != "# Title\r\n" {
		t.Fatalf("raw not preserved: %q", doc.ContentMarkdownRaw)
	}
	if doc.Source.FileName != "a.md" {
		t.Fatalf("source lost: %+v", doc.Source)
	}
}

func TestVersionConstant(t *testing.T) {
	if Version != "canonical-normalizer-v1" {
		t.Fatalf("unexpected version %q", Version)
	}
}
