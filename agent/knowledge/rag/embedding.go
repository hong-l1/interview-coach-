package rag

import (
	"context"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"os"
)

func NewEmbedder(ctx context.Context) *ark.Embedder {
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("api_key"),
		Model:  os.Getenv("embedding_id"),
	})
	if err != nil {
		panic(err)
	}
	return embedder
}
