package retrieval

// dedupeKey 构建 document_id:chunk_id 作为去重键，缺少时回退到 item.ID。
func dedupeKey(it Item) string {
	docID, _ := it.Metadata["document_id"].(string)
	chunkID, _ := it.Metadata["chunk_id"].(string)
	if docID == "" {
		docID = it.ID
	}
	if chunkID == "" {
		chunkID = it.ID
	}
	return docID + ":" + chunkID
}

// Dedupe 按 document_id:chunk_id 去重，保留首个（分数最高，已排序）。
func Dedupe(hits []Item) []Item {
	seen := map[string]struct{}{}
	out := make([]Item, 0, len(hits))
	for _, h := range hits {
		k := dedupeKey(h)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, h)
	}
	return out
}
