package retrieval

import (
	"context"

	"awesomeProject4/agent/knowledge/rag"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// Searcher 抽象向量检索，便于注入 mock。
type Searcher interface {
	Search(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error)
}

// milvusSearcher 适配现有 rag.HybridRetrieve。
type milvusSearcher struct {
	client *milvusclient.Client
}

// NewMilvusSearcher 构造一个基于 milvusclient 的 Searcher。
func NewMilvusSearcher(client *milvusclient.Client) Searcher {
	return &milvusSearcher{client: client}
}

// Search 委托给 rag.HybridRetrieve。
func (m *milvusSearcher) Search(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error) {
	return rag.HybridRetrieve(ctx, m.client, query, topK, filter)
}
