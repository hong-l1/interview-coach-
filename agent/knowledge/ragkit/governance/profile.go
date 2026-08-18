package governance

import "encoding/json"

// StrategyProfile 是领域模型，与 GORM 行解耦。
type StrategyProfile struct {
	ID                     uint64
	Name                   string
	Status                 string
	FusionStrategy         string
	TopKConfig             TopKConfigJSON
	EvidenceGateThresholds EvidenceGateJSON
}

type TopKConfigJSON struct {
	BaseCandidateTopK  int `json:"base_candidate_top_k"`
	BaseFinalTopK      int `json:"base_final_top_k"`
	BroadCandidateTopK int `json:"broad_candidate_top_k"`
}

type EvidenceGateJSON struct {
	Enabled             bool    `json:"enabled"`
	MinRerankScore      float64 `json:"min_rerank_score"`
	MinCitationCoverage float64 `json:"min_citation_coverage"`
}

func (p *StrategyProfile) ToTopKConfigJSON() ([]byte, error) { return json.Marshal(p.TopKConfig) }
func (p *StrategyProfile) ToEvidenceGateJSON() ([]byte, error) {
	return json.Marshal(p.EvidenceGateThresholds)
}
