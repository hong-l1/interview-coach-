package ragkit

import (
	"context"

	"awesomeProject4/agent/knowledge/ragkit/retrieval"
)

// Version 是 ragkit 实现版本号，用于审计与元数据留痕。
const Version = "ragkit-v0"

// Retrieve 用 searcher 搜索 + 后处理（facade 入口）。
func Retrieve(ctx context.Context, searcher retrieval.Searcher, query, filter string, profile retrieval.RetrieveProfile) (retrieval.Result, error) {
	cand, _ := retrieval.DecideDynamicTopK(query, profile.TopK)
	docs, err := searcher.Search(ctx, query, cand, filter)
	if err != nil {
		return retrieval.Result{}, err
	}
	items := retrieval.ToItems(docs)
	return retrieval.PostProcess(ctx, query, items, profile)
}
