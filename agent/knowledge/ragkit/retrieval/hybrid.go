package retrieval

import (
	"context"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/retriever/milvus2"
	"github.com/cloudwego/eino-ext/components/retriever/milvus2/search_mode"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// NewHybridRetriever 构造 RRF 混合检索器：dense COSINE + sparse BM25（K=60），
// 集合名来自环境变量 collection。embedder 由调用方注入（避免对 ragkit 根包循环依赖）。
func NewHybridRetriever(ctx context.Context, client *milvusclient.Client, embedder embedding.Embedder) *milvus2.Retriever {
	reranker := milvusclient.NewRRFReranker().WithK(60)
	hybridMode := search_mode.NewHybrid(reranker,
		&search_mode.SubRequest{
			VectorField: "vector",
			VectorType:  milvus2.DenseVector,
			TopK:        10,
			MetricType:  milvus2.COSINE,
		},
		&search_mode.SubRequest{
			VectorField: "sparse_vector",
			VectorType:  milvus2.SparseVector,
			TopK:        10,
			MetricType:  milvus2.BM25,
		},
	)
	retriever, err := milvus2.NewRetriever(ctx, &milvus2.RetrieverConfig{
		Client:            client,
		Embedding:         embedder,
		Collection:        os.Getenv("collection"),
		VectorField:       "vector",
		SparseVectorField: "sparse_vector",
		OutputFields:      []string{"id", "content", "metadata"},
		TopK:              5,
		SearchMode:        hybridMode,
	})
	if err != nil {
		panic(err)
	}
	return retriever
}

// HybridRetrieve 执行 RRF 混合检索，支持 topK 与 filter 覆盖。
func HybridRetrieve(ctx context.Context, client *milvusclient.Client, embedder embedding.Embedder, query string, topK int, filter string) ([]*schema.Document, error) {
	r := NewHybridRetriever(ctx, client, embedder)
	opts := make([]retriever.Option, 0, 2)
	if topK > 0 {
		opts = append(opts, retriever.WithTopK(topK))
	}
	if strings.TrimSpace(filter) != "" {
		opts = append(opts, milvus2.WithFilter(filter))
	}
	return r.Retrieve(ctx, query, opts...)
}
