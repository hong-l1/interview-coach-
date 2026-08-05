package retrieval

import "strings"

const topkRuleVersion = "phase2-rule-v1"

// TopKConfig 控制动态/战略 TopK。
type TopKConfig struct {
	BaseCandidateTopK  int
	BaseFinalTopK      int
	BroadCandidateTopK int
	ShortQueryRunes    int     // 低于此长度视为宽泛
	ShortQueryWords    int     // 词数低于此视为宽泛
	ScoreCliffRatio    float64 // 相邻分差超过前一个的此比例视为悬崖
	TokenBudget        int
}

// DefaultTopKConfig 返回阶段二默认 TopK 配置。
func DefaultTopKConfig() TopKConfig {
	return TopKConfig{
		BaseCandidateTopK:  10,
		BaseFinalTopK:      5,
		BroadCandidateTopK: 15,
		ShortQueryRunes:    8,
		ShortQueryWords:    3,
		ScoreCliffRatio:    0.3,
		TokenBudget:        1200,
	}
}

// DecideDynamicTopK 由查询宽度决定搜索召回上限与最终返回数。
// 短查询（runes<ShortQueryRunes 或 words<ShortQueryWords）扩候选到 BroadCandidateTopK；
// final 始终保持 BaseFinalTopK，作为后处理后返回条数的信号。
func DecideDynamicTopK(query string, cfg TopKConfig) (candidate, final int) {
	candidate = cfg.BaseCandidateTopK
	final = cfg.BaseFinalTopK
	runes := len([]rune(query))
	words := len(strings.Fields(query))
	if runes < cfg.ShortQueryRunes || words < cfg.ShortQueryWords {
		candidate = cfg.BroadCandidateTopK
	}
	return candidate, final
}

// DecideStrategicTopK 按 rerank 分数分布收敛到最终保留条数。
// 起点 BaseFinalTopK；若分数悬崖守护给出更小值则采用；再 clamp 到 len(hits) 与 >=0。
func DecideStrategicTopK(hits []Item, cfg TopKConfig) int {
	final := cfg.BaseFinalTopK
	if cliff := ApplyScoreCliffGuard(hits); cliff < final {
		final = cliff
	}
	if final > len(hits) {
		final = len(hits)
	}
	if final < 0 {
		final = 0
	}
	return final
}

// ApplyScoreCliffGuard 返回在首个分数悬崖前应保留的条数。
// 扫描相邻对，若 (prev-curr)/prev > 0.3 视为悬崖，返回该 cliff 索引（即此前条数）。
// 跳过 Score<=0 的项以防除零。无悬崖时返回 len(hits)。
func ApplyScoreCliffGuard(hits []Item) int {
	for i := 1; i < len(hits); i++ {
		prev := hits[i-1].Score
		if prev <= 0 {
			continue
		}
		drop := (prev - hits[i].Score) / prev
		if drop > 0.3 {
			return i
		}
	}
	return len(hits)
}

// ApplyTokenBudgetGuard 按 token 预算截断。
// 累加 approxTokensItem；当 used+next > budget 时停止。
func ApplyTokenBudgetGuard(hits []Item, budget int) []Item {
	used := 0
	var out []Item
	for _, h := range hits {
		tok := approxTokensItem(h.Content)
		if used+tok > budget {
			break
		}
		used += tok
		out = append(out, h)
	}
	return out
}

// approxTokensItem 估算单条 Content 的 token 数：
// 优先取词数（strings.Fields），无词时回退到 (runes+3)/4。
func approxTokensItem(s string) int {
	words := strings.Fields(s)
	if len(words) > 0 {
		return len(words)
	}
	return (len([]rune(s)) + 3) / 4
}
