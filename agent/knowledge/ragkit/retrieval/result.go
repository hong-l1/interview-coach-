package retrieval

import "github.com/cloudwego/eino/schema"

// GateOutcome 是证据门控结果。
type GateOutcome string

const (
	GatePass     GateOutcome = "pass"
	GateRefused  GateOutcome = "refused"
	GateDegraded GateOutcome = "degraded_pass"
	GateDisabled GateOutcome = "disabled"
)

// EvidenceGate 记录门控判定。
type EvidenceGate struct {
	Outcome GateOutcome
	Reason  string // No-Retrieval-Hit / Low-Rerank-Confidence / Insufficient-Citation-Coverage / Contradictory-Evidence / Out-Of-KB-Scope
}

// Item 是检索返回的单条结果。
type Item struct {
	ID       string
	Content  string
	Score    float64
	Metadata map[string]any
}

// Metrics 记录各阶段计数（观测用，不写库）。
type Metrics struct {
	CandidateTopK   int
	FinalTopK       int
	DedupedCount    int
	RerankApplied   bool
	CitationChecked bool
	EvidenceGate    EvidenceGate
}

// Result 是检索后处理最终结果。
type Result struct {
	Items        []Item
	Metrics      Metrics
	EvidenceGate EvidenceGate
}

// ToItems 把 eino Document 转成 Item，分数通过 doc.Score() 读取。
// 该函数被 Task 13 的 facade 复用，故导出。
func ToItems(docs []*schema.Document) []Item {
	items := make([]Item, 0, len(docs))
	for _, d := range docs {
		items = append(items, Item{
			ID:       d.ID,
			Content:  d.Content,
			Score:    d.Score(),
			Metadata: d.MetaData,
		})
	}
	return items
}
