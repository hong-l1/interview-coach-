package retrieval

import "strings"

// EvidenceGateThresholds 控制门控判定。
type EvidenceGateThresholds struct {
	Enabled              bool
	MinRerankScore       float64
	MinCitationCoverage float64
}

// DefaultEvidenceGateThresholds 返回阶段二默认门控阈值。
func DefaultEvidenceGateThresholds() EvidenceGateThresholds {
	return EvidenceGateThresholds{
		Enabled:              true,
		MinRerankScore:       0.3,
		MinCitationCoverage: 0.5,
	}
}

// EvaluateEvidenceGate 对已重排结果做门控判定。
func EvaluateEvidenceGate(hits []Item, query string, th EvidenceGateThresholds, enabled bool) EvidenceGate {
	if !enabled || !th.Enabled {
		return EvidenceGate{Outcome: GateDisabled}
	}
	if len(hits) == 0 {
		return EvidenceGate{Outcome: GateRefused, Reason: "No-Retrieval-Hit"}
	}
	top := hits[0]
	if top.Score < th.MinRerankScore {
		return EvidenceGate{Outcome: GateRefused, Reason: "Low-Rerank-Confidence"}
	}
	// 矛盾证据：top 命中含强否定且 query 不含否定
	if contradictionDetected(hits) && !strings.Contains(query, "不") {
		return EvidenceGate{Outcome: GateRefused, Reason: "Contradictory-Evidence"}
	}
	return EvidenceGate{Outcome: GatePass}
}

var negationCues = []string{"不支持", "不能", "并非", "不是"}

// contradictionDetected 扫描 top 3 命中是否含强否定线索。
func contradictionDetected(hits []Item) bool {
	if len(hits) < 2 {
		return false
	}
	for _, h := range hits[:min(3, len(hits))] {
		for _, cue := range negationCues {
			if strings.Contains(h.Content, cue) {
				return true
			}
		}
	}
	return false
}

// min 返回两整数中较小者。
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
