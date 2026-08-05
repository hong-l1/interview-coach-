package retrieval

import (
	"sort"
	"strings"
)

// JaccardReranker 用词法 Jaccard 重叠度调整排序，原分和 Jaccard 各占 0.5。
type JaccardReranker struct{}

// Rerank 用词法 Jaccard/Coverage 调整排序，原分占 0.5。
func (j *JaccardReranker) Rerank(query string, hits []Item) []Item {
	qSet := tokenSet(query)
	out := make([]Item, len(hits))
	for i, h := range hits {
		dSet := tokenSet(h.Content)
		jacc := jaccard(qSet, dSet)
		merged := 0.5*h.Score + 0.5*jacc
		out[i] = h
		out[i].Score = merged
		if out[i].Metadata == nil {
			out[i].Metadata = map[string]any{}
		}
		out[i].Metadata["rerank_score"] = merged
		out[i].Metadata["jaccard"] = jacc
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Score > out[b].Score })
	return out
}

func tokenSet(s string) map[string]struct{} {
	set := map[string]struct{}{}
	// ASCII 词
	for _, w := range strings.Fields(strings.ToLower(s)) {
		if len(w) > 1 {
			set[w] = struct{}{}
		}
	}
	// CJK bigram
	r := []rune(s)
	for i := 0; i+1 < len(r); i++ {
		set[string(r[i:i+2])] = struct{}{}
	}
	return set
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	inter := 0
	uni := 0
	for k := range a {
		if _, ok := b[k]; ok {
			inter++
		} else {
			uni++
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			uni++
		}
	}
	return float64(inter) / float64(inter+uni)
}
