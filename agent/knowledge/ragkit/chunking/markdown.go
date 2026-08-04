package chunking

import (
	"bytes"
	"context"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownStrategy splits markdown documents by heading sections using
// github.com/yuin/goldmark AST. Each chunk carries metadata for
// section_title, hierarchy_path, heading_level, and chunking_unit.
type MarkdownStrategy struct {
	maxChunkBytes int
}

// NewMarkdownStrategy creates a MarkdownStrategy. If maxChunkBytes <= 0,
// defaults to 1000.
func NewMarkdownStrategy(maxChunkBytes int) *MarkdownStrategy {
	if maxChunkBytes <= 0 {
		maxChunkBytes = 1000
	}
	return &MarkdownStrategy{maxChunkBytes: maxChunkBytes}
}

// Name returns the strategy identifier.
func (m *MarkdownStrategy) Name() string { return StrategyMarkdown }

// Split parses the document as markdown, walks the AST, and emits chunks
// keyed on heading sections. Paragraph/section content that exceeds
// maxChunkBytes is further subdivided at sentence boundaries.
func (m *MarkdownStrategy) Split(ctx context.Context, req Request) ([]*schema.Document, error) {
	src := []byte(req.Document.ContentMarkdown)
	md := goldmark.New()
	reader := text.NewReader(src)
	root := md.Parser().Parse(reader)

	var chunks []*schema.Document
	pathStack := make([]string, 0)

	err := ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n.Kind() {
		case ast.KindHeading:
			heading := n.(*ast.Heading)
			level := heading.Level
			title := strings.TrimSpace(string(n.Text(src)))

			// Pop pathStack entries at or below the current heading level.
			for len(pathStack) >= level {
				pathStack = pathStack[:len(pathStack)-1]
			}
			pathStack = append(pathStack, title)
			return ast.WalkContinue, nil

		case ast.KindParagraph:
			seg := bytes.TrimSpace(n.Text(src))
			if len(seg) == 0 {
				return ast.WalkContinue, nil
			}
			parts := splitByByteLimit(string(seg), m.maxChunkBytes)
			for _, p := range parts {
				chunks = append(chunks, m.makeChunk(p, pathStack))
			}
			// Skip recursing into inline children (text nodes, etc.).
			return ast.WalkSkipChildren, nil

		default:
			return ast.WalkContinue, nil
		}
	})
	if err != nil {
		return nil, err
	}

	// Fallback: if no paragraphs were detected, emit a single chunk for the
	// whole document.
	if len(chunks) == 0 && len(src) > 0 {
		chunks = append(chunks, m.makeChunk(string(bytes.TrimSpace(src)), pathStack))
	}

	return chunks, nil
}

// makeChunk creates a schema.Document with section metadata derived from the
// current heading path.
func (m *MarkdownStrategy) makeChunk(content string, path []string) *schema.Document {
	title := ""
	if len(path) > 0 {
		title = path[len(path)-1]
	}
	return &schema.Document{
		Content: content,
		MetaData: map[string]any{
			"section_title":  title,
			"hierarchy_path": strings.Join(path, " > "),
			"heading_level":  len(path),
			"chunking_unit":  "heading_section",
		},
	}
}

// splitByByteLimit splits content at a byte limit, preferring boundaries at
// newlines or Chinese sentence-ending punctuation when the split point is
// past the halfway mark.
func splitByByteLimit(s string, limit int) []string {
	if len(s) <= limit {
		return []string{s}
	}
	var parts []string
	for len(s) > limit {
		cut := limit
		if i := strings.LastIndexAny(s[:cut], "\n。；！？"); i > limit/2 {
			cut = i + 1 // include the boundary character
		}
		parts = append(parts, strings.TrimSpace(s[:cut]))
		s = s[cut:]
	}
	if strings.TrimSpace(s) != "" {
		parts = append(parts, strings.TrimSpace(s))
	}
	return parts
}
