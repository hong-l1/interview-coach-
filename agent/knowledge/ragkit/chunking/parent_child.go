package chunking

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"github.com/cloudwego/eino/schema"
)

// FinalizeChunks enriches each chunk with chunk_id / parent_id / offset /
// hierarchy metadata derived from the canonical markdown and heading structure.
func FinalizeChunks(chunks []*schema.Document, doc *canonical.NormalizedDocument, documentID string) []*schema.Document {
	canonicalMd := doc.ContentMarkdown
	parents := buildHeadingParents(canonicalMd)
	for _, c := range chunks {
		if c.MetaData == nil {
			c.MetaData = map[string]any{}
		}
		off := strings.Index(canonicalMd, c.Content)
		start, end := off, off+len(c.Content)
		c.MetaData["child_start_offset"] = start
		c.MetaData["child_end_offset"] = end
		chunkID := shortHash(documentID + ":" + c.Content)
		c.MetaData["chunk_id"] = chunkID
		c.MetaData["child_id"] = chunkID
		c.MetaData["document_id"] = documentID
		// Resolve the narrowest heading section that contains this chunk.
		parent := resolveParent(parents, start, end)
		if parent != nil {
			c.MetaData["parent_id"] = parent.id
			c.MetaData["parent_start_offset"] = parent.start
			c.MetaData["parent_end_offset"] = parent.end
			c.MetaData["parent_build_strategy"] = parent.buildStrategy
		} else if c.MetaData["chunking_unit"] == "table" {
			c.MetaData["parent_build_strategy"] = "table"
		} else {
			c.MetaData["parent_build_strategy"] = "paragraph_window"
		}
		c.MetaData["parent_build_version"] = "phase3-parent-child-v1"
		c.MetaData["parent_token_count"] = approxTokens(c.Content)
	}
	return chunks
}

// headingParent records the byte-offset span and identity of a heading section.
type headingParent struct {
	id, buildStrategy string
	start, end        int
}

// buildHeadingParents scans canonical markdown for "# " / "## " heading lines
// and builds a slice of headingParent with byte offsets. Each section ends at
// the next heading (or EOF).
func buildHeadingParents(md string) []headingParent {
	lines := strings.Split(md, "\n")
	var parents []headingParent
	byteOff := 0
	curStart := -1
	curTitle := ""
	flush := func(end int) {
		if curStart >= 0 {
			parents = append(parents, headingParent{
				id:            shortHash("section:" + curTitle),
				buildStrategy: "heading_section",
				start:         curStart,
				end:           end,
			})
		}
	}
	for _, line := range lines {
		trim := strings.TrimLeft(line, " ")
		if strings.HasPrefix(trim, "# ") || strings.HasPrefix(trim, "## ") {
			flush(byteOff)
			curStart = byteOff
			curTitle = strings.TrimSpace(strings.TrimLeft(trim, "# "))
		}
		byteOff += len(line) + 1
	}
	flush(byteOff)
	return parents
}

// resolveParent returns the narrowest headingParent whose span fully contains
// [start, end), or nil if none match.
func resolveParent(parents []headingParent, start, end int) *headingParent {
	var best *headingParent
	for i := range parents {
		p := parents[i]
		if p.start <= start && end <= p.end {
			if best == nil || (p.end-p.start) < (best.end-best.start) {
				best = &p
			}
		}
	}
	return best
}

// shortHash returns the first 16 hex characters of the SHA-1 hash of s.
func shortHash(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:8])
}

// approxTokens estimates the token count of s using word count, falling back
// to rune count divided by 4.
func approxTokens(s string) int {
	words := strings.Fields(s)
	if len(words) > 0 {
		return len(words)
	}
	return (len([]rune(s)) + 3) / 4
}
