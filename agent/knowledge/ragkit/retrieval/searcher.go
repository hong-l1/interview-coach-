package retrieval

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// Searcher 抽象向量检索，便于注入 mock。
type Searcher interface {
	Search(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error)
}

// milvusSearcher 适配 RRF 混合检索（HybridRetrieve）。
type milvusSearcher struct {
	client   *milvusclient.Client
	embedder embedding.Embedder
}

// NewMilvusSearcher 构造一个基于 milvusclient 的 Searcher，embedder 由调用方注入。
func NewMilvusSearcher(client *milvusclient.Client, embedder embedding.Embedder) Searcher {
	return &milvusSearcher{client: client, embedder: embedder}
}

// Search 委托给 HybridRetrieve。
func (m *milvusSearcher) Search(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error) {
	return HybridRetrieve(ctx, m.client, m.embedder, query, topK, filter)
}
