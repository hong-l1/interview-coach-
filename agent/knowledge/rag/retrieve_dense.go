package rag

import (
	"context"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/retriever/milvus2"
	"github.com/cloudwego/eino-ext/components/retriever/milvus2/search_mode"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func NewDenseRetriever(ctx context.Context, client *milvusclient.Client) *milvus2.Retriever {
	embedder := NewEmbedder(ctx)
	denseMode := search_mode.NewApproximate(milvus2.COSINE)
	retriever, err := milvus2.NewRetriever(ctx, &milvus2.RetrieverConfig{
		Client:            client,
		Embedding:         embedder,
		Collection:        os.Getenv("collection"),
		VectorField:       "vector",
		SparseVectorField: "sparse_vector",
		OutputFields:      []string{"id", "content", "metadata"},
		TopK:              5,
		SearchMode:        denseMode,
	})
	if err != nil {
		panic(err)
	}
	return retriever
}

func DenseRetrieve(ctx context.Context, client *milvusclient.Client, query string, topK int, filter string) ([]*schema.Document, error) {
	r := NewDenseRetriever(ctx, client)
	opts := make([]retriever.Option, 0, 2)
	if topK > 0 {
		opts = append(opts, retriever.WithTopK(topK))
	}
	if strings.TrimSpace(filter) != "" {
		opts = append(opts, milvus2.WithFilter(filter))
	}
	return r.Retrieve(ctx, query, opts...)
}
