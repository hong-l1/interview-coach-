package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	rr "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/redis/go-redis/v9"
	"os"
)

func main() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:          "43.142.57.35:6379",
		Protocol:      2,
		UnstableResp3: true,
	})
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("api_key"),
		Model:  os.Getenv("embedding_id"),
	})
	if err != nil {
		panic(err)
	}
	r, err := rr.NewRetriever(ctx, &rr.RetrieverConfig{
		Client:    client,
		Index:     "doc_index",
		Embedding: embedder,
	})
	if err != nil {
		panic(err)
	}
	docs, err := r.Retrieve(ctx, "dog")
	if err != nil {
		panic(err)
	}
	for _, v := range docs {
		fmt.Printf("ID:%s, CONTENT:%v \n", v.ID, v.Content)
	}
}
