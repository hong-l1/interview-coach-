package ragkit

import (
	"context"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
)

// NewEmbedder 构造 Ark（qwen3-embedding）embedder，配置来自环境变量：
// api_key / embedding_id / base_url（对齐现有栈，沿用 rag 包行为）。
func NewEmbedder(ctx context.Context) *ark.Embedder {
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey:  os.Getenv("api_key"),
		Model:   os.Getenv("embedding_id"),
		BaseURL: os.Getenv("base_url"),
	})
	if err != nil {
		panic(err)
	}
	return embedder
}
