package canonical

import (
	"crypto/sha1"
	"encoding/hex"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestNormalizeRunsFullPipelineAndRecordsRules(t *testing.T) {
	raw := "#title\r\n\r\n| a | b |\n| --- | --- |\n| 1 | 2 |\n\n## **bold**\nbody"
	doc := NewNormalizedDocument(raw, Source{FileName: "x.md"})
	rules := Normalize(doc)
	if len(rules) == 0 {
		t.Fatal("no rules recorded")
	}
	if doc.ContentMarkdown == "" {
		t.Fatal("ContentMarkdown not set")
	}
	if doc.ContentMarkdown == raw {
		t.Fatal("ContentMarkdown unchanged")
	}
	if doc.Canonicalization.Version != Version {
		t.Fatalf("version mismatch: %q", doc.Canonicalization.Version)
	}
	if doc.Canonicalization.RawSHA1 == "" || doc.Canonicalization.CanonicalSHA1 == "" {
		t.Fatal("sha1 missing")
	}
	rawSum := sha1.Sum([]byte(raw))
	if doc.Canonicalization.RawSHA1 != hex.EncodeToString(rawSum[:]) {
		t.Fatalf("raw sha1 mismatch")
	}
	if len(doc.Tables) != 1 {
		t.Fatalf("want 1 table, got %d", len(doc.Tables))
	}
}

func TestAnnotateChunksWithProvenance(t *testing.T) {
	raw := "# Title\n\nintro paragraph here\n\n## Section\n\nsecond paragraph text"
	doc := NewNormalizedDocument(raw, Source{FileName: "x.md"})
	Normalize(doc)
	chunks := []*schema.Document{
		{ID: "c1", Content: "intro paragraph here"},
		{ID: "c2", Content: "second paragraph text"},
	}
	out := AnnotateChunksWithProvenance(chunks, doc)
	if out[0].MetaData["provenance_text"] != "intro paragraph here" {
		t.Fatalf("provenance not set: %v", out[0].MetaData)
	}
}
