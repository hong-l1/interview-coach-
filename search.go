package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/redis/go-redis/v9"
	"math"
	"os"
)

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:          "43.142.57.35:6379", // Redis Stack 服务的地址
		UnstableResp3: true,
		Protocol:      2,
	})
	indexName := "doc_index"
	k := 2
	query := fmt.Sprintf("*=>[KNN %d @vector_content $blob AS score]", k)
	searchContent := "That is a happy person"
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("api_key"),
		Model:  os.Getenv("embedding_id"),
	})
	if err != nil {
		panic(err)
	}
	embeddings, err := embedder.EmbedStrings(ctx, []string{searchContent})
	if err != nil {
		panic(err)
	}
	// 使用 redis.NewSearch 来构建带参数的查询
	searchResult, err := rdb.FTSearchWithArgs(ctx, indexName, query, &redis.FTSearchOptions{
		Params: map[string]interface{}{
			"blob": vector2Bytes(embeddings[0]),
		},
		DialectVersion: 2,
		Return: []redis.FTSearchReturn{
			{
				FieldName: "content",
			},
			{
				FieldName: "score",
			},
		},
	}).Result()
	if err != nil {
		panic(err)
	}
	for _, v := range searchResult.Docs {
		fmt.Printf("%v:%v \n", v.Fields["content"], v.Fields["score"])
	}
}
func vector2Bytes(vector []float64) []byte {
	float32Arr := make([]float32, len(vector))
	for i, v := range vector {
		float32Arr[i] = float32(v)
	}
	bytes := make([]byte, len(float32Arr)*4)
	for i, v := range float32Arr {
		binary.LittleEndian.PutUint32(bytes[i*4:], math.Float32bits(v))
	}
	return bytes
}
