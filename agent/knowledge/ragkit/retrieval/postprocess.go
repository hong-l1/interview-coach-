package retrieval

import "context"

// RetrieveProfile 聚合检索后处理开关与阈值，来自 governance.active profile。
type RetrieveProfile struct {
	TopK            TopKConfig
	RerankEnabled   bool
	EvidenceGate    EvidenceGateThresholds
	CitationEnabled bool
}

// DefaultRetrieveProfile 返回默认的检索配置。
func DefaultRetrieveProfile() RetrieveProfile {
	return RetrieveProfile{
		TopK:            DefaultTopKConfig(),
		RerankEnabled:   true,
		EvidenceGate:    DefaultEvidenceGateThresholds(),
		CitationEnabled: false,
	}
}

// PostProcess 对搜索结果执行后处理流水线。
func PostProcess(ctx context.Context, query string, items []Item, profile RetrieveProfile) (Result, error) {
	m := Metrics{CandidateTopK: len(items)}

	items = Dedupe(items)
	m.DedupedCount = len(items)

	if profile.RerankEnabled {
		items = (&JaccardReranker{}).Rerank(query, items)
		m.RerankApplied = true
	}
	if profile.CitationEnabled {
		items = (&CitationConsistencyChecker{Enabled: true}).Check(query, items)
		m.CitationChecked = true
	}

	final := DecideStrategicTopK(items, profile.TopK)
	m.FinalTopK = final
	items = ApplyTokenBudgetGuard(items, profile.TopK.TokenBudget)
	if final > len(items) {
		final = len(items)
	}
	if final > 0 {
		items = items[:final]
	} else {
		items = nil
	}

	gate := EvaluateEvidenceGate(items, query, profile.EvidenceGate, true)
	m.EvidenceGate = gate
	if gate.Outcome == GateRefused {
		return Result{Items: nil, Metrics: m, EvidenceGate: gate}, nil
	}
	return Result{Items: items, Metrics: m, EvidenceGate: gate}, nil
}
