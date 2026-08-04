package canonical

import (
	"strings"

	"github.com/cloudwego/eino/schema"
)

// AnnotateChunksWithProvenance 把 chunk 内容定位回 canonical markdown，
// 写入 provenance_text / canonical_offset 元数据。
func AnnotateChunksWithProvenance(chunks []*schema.Document, doc *NormalizedDocument) []*schema.Document {
	canonical := doc.ContentMarkdown
	for _, c := range chunks {
		if c.MetaData == nil {
			c.MetaData = map[string]any{}
		}
		c.MetaData["provenance_text"] = c.Content
		if off := strings.Index(canonical, c.Content); off >= 0 {
			c.MetaData["canonical_offset"] = off
		}
		// 区块映射：找包含该 offset 的 block
		if off, ok := c.MetaData["canonical_offset"].(int); ok {
			for _, b := range doc.Blocks {
				if off >= b.MarkdownStart && off < b.MarkdownEnd {
					c.MetaData["block_ids"] = []string{b.ID}
					c.MetaData["page"] = b.Page
					break
				}
			}
		}
	}
	return chunks
}
