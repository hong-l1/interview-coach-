package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"os"
	"unsafe"
)

func main() {
	ctx := context.Background()
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("api_key"),
		Model:  os.Getenv("embedding_id"),
	})
	if err != nil {
		panic(err)
	}

	embeddings, err := embedder.EmbedStrings(ctx, []string{"hello world"})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Embedding value: %v \n", embeddings[0][:5]) // 只打印前5个值
	fmt.Printf("Embedding length: %d\n", len(embeddings[0]))
	fmt.Printf("Embedding type: %T\n", embeddings[0][0]) // 查看元素类型
	fmt.Printf("Embedding byte size: %d bytes per element\n", unsafe.Sizeof(embeddings[0][0]))
}
