package tool

import (
	"awesomeProject4/Init"
	"awesomeProject4/agent/knowledge/rag"
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
	// === ragkit 接线点（门控 + 动态 TopK），默认关闭 ===
	// 启用方式：设置环境变量 RAGKIT_ENABLED=1，并把本函数返回替换为 ragkit.Retrieve
	// res, _ := ragkit.Retrieve(searchCtx, retrieval.NewMilvusSearcher(manager.Client), query, input.Filter, retrieval.DefaultRetrieveProfile())
	// return toRetrieverOutput(res), nil
	docs, err := rag.HybridRetrieve(searchCtx, manager.Client, query, topK, input.Filter)
	if err != nil {
		return nil, err
	}

	items := make([]RetrieverDoc, 0, len(docs))
	for _, doc := range docs {
		items = append(items, RetrieverDoc{
			ID:       doc.ID,
			Content:  doc.Content,
			Score:    doc.Score(),
			Metadata: doc.MetaData,
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
