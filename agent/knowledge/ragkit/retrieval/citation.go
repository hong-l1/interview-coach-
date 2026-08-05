package retrieval

// CitationConsistencyChecker 检查每条命中是否支撑 query 中的 claim。
// 本轮默认关闭（面试 query 多为短问题，claim 抽取噪声大），留接口。
type CitationConsistencyChecker struct {
	Enabled bool
}

// Check 在启用时对每条命中做 claim 支撑度评分；默认关闭直接返回原 hits。
func (c *CitationConsistencyChecker) Check(query string, hits []Item) []Item {
	if !c.Enabled {
		return hits
	}
	// TODO(enabled=true): extract claims → score per doc → mark citation_supported
	return hits
}
