package tool

import (
	"awesomeProject4/Init"
	"awesomeProject4/agent/knowledge/ragkit"
	"awesomeProject4/agent/knowledge/ragkit/retrieval"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type RetrieverInput struct {
	Query  string `json:"query"`
	TopK   int    `json:"top_k"`
	Filter string `json:"filter"`
}

type RetrieverOutput struct {
	Query     string         `json:"query"`
	TopK      int            `json:"top_k"`
	Filter    string         `json:"filter,omitempty"`
	Documents []RetrieverDoc `json:"documents"`
	Total     int            `json:"total"`
}

type RetrieverDoc struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Score    float64        `json:"score"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func GetRetrieverWithInput(ctx context.Context, input *RetrieverInput) (*RetrieverOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("get_milvus_retriever: request is nil")
	}
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return nil, fmt.Errorf("get_milvus_retriever: query is required")
	}
	topK := input.TopK
	if topK <= 0 {
		topK = 5
	}
	if topK > 20 {
		topK = 20
	}

	searchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	manager := Init.NewMilvusManger()
	searcher := ragkit.NewMilvusSearcher(searchCtx, manager.Client)

	// ragkit 检索链路：动态 TopK + 去重 + Jaccard 重排 + 证据门控
	profile := retrieval.DefaultRetrieveProfile()
	profile.TopK.BaseFinalTopK = topK // 尊重调用方 top_k
	res, err := ragkit.Retrieve(searchCtx, searcher, query, input.Filter, profile)
	if err != nil {
		return nil, err
	}

	items := make([]RetrieverDoc, 0, len(res.Items))
	for _, it := range res.Items {
		items = append(items, RetrieverDoc{
			ID:       it.ID,
			Content:  it.Content,
			Score:    it.Score,
			Metadata: it.Metadata,
		})
	}
	return &RetrieverOutput{
		Query:     query,
		TopK:      topK,
		Filter:    input.Filter,
		Documents: items,
		Total:     len(items),
	}, nil
}

func CreateRetrieverTool() tool.InvokableTool {
	tl, err := utils.InferTool(
		"get_milvus_retriever",
		"Hybrid search knowledge snippets from Milvus by query. Input requires query; optional top_k and filter.",
		GetRetrieverWithInput,
	)
	if err != nil {
		panic(err)
	}
	return tl
}
