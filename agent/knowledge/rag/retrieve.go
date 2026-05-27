package rag

import (
	"context"
	"strings"

	"github.com/cloudwego/eino-ext/components/retriever/milvus2"
	"github.com/cloudwego/eino-ext/components/retriever/milvus2/search_mode"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"os"
)

func NewHybridRetriever(ctx context.Context, client *milvusclient.Client) *milvus2.Retriever {
	embedder := NewEmbedder(ctx)
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

func HybridRetrieve(ctx context.Context, client *milvusclient.Client, query string, topK int, filter string) ([]*schema.Document, error) {
	r := NewHybridRetriever(ctx, client)
	opts := make([]retriever.Option, 0, 2)
	if topK > 0 {
		opts = append(opts, retriever.WithTopK(topK))
	}
	if strings.TrimSpace(filter) != "" {
		opts = append(opts, milvus2.WithFilter(filter))
	}
	return r.Retrieve(ctx, query, opts...)
}

func FloatVectorConverter(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
	ans := make([]entity.Vector, 0, len(vectors))
	for _, vector := range vectors {
		float32Vec := make([]float32, len(vector))
		for k, v := range vector {
			float32Vec[k] = float32(v)
		}
		ans = append(ans, entity.FloatVector(float32Vec))
	}
	return ans, nil
}
