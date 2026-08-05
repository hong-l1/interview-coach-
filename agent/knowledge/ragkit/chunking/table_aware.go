package chunking

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// TableAwareStrategy renders each table in the document as a single chunk in
// "key: value" row format (header row as keys, data rows as values).
// When the document has no tables, Split returns nil, nil so that a router
// can fall back to the default strategy.
type TableAwareStrategy struct {
	maxChunkBytes int
}

// NewTableAwareStrategy creates a TableAwareStrategy. If maxChunkBytes <= 0,
// it defaults to 1000.
func NewTableAwareStrategy(maxChunkBytes int) *TableAwareStrategy {
	if maxChunkBytes <= 0 {
		maxChunkBytes = 1000
	}
	return &TableAwareStrategy{maxChunkBytes: maxChunkBytes}
}

// Name returns the strategy identifier.
func (t *TableAwareStrategy) Name() string { return StrategyTableAware }

// Split renders each table in req.Document.Tables as an independent chunk.
// Header row cells are used as keys, and each data row is rendered as
// "key: value" lines. Metadata records chunking_unit="table", table_index,
// and table_page. Returns nil, nil when the document has no tables.
func (t *TableAwareStrategy) Split(ctx context.Context, req Request) ([]*schema.Document, error) {
	if len(req.Document.Tables) == 0 {
		return nil, nil
	}
	var chunks []*schema.Document
	for i, tbl := range req.Document.Tables {
		if len(tbl.Rows) == 0 {
			continue
		}
		header := tbl.Rows[0]
		var lines []string
		for _, r := range tbl.Rows[1:] {
			for j, cell := range r {
				key := ""
				if j < len(header) {
					key = header[j]
				}
				lines = append(lines, key+": "+cell)
			}
		}
		chunks = append(chunks, &schema.Document{
			Content: strings.Join(lines, "\n"),
			MetaData: map[string]any{
				"chunking_unit": "table",
				"table_index":   i,
				"table_page":    tbl.Page,
			},
		})
	}
	return chunks, nil
}
