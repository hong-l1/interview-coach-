package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	ri "github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
	"os"
)

const (
	FieldContent       = "content"
	FieldContentVector = "content_vector"
)

func main() {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:          "43.142.57.35:6379",
		UnstableResp3: true,
		Protocol:      2,
	})
	keyPrefix := "eino_doc:"
	indexName := "doc_index"
	slice, err2 := client.Do(ctx, "FT._LIST").StringSlice()
	if err2 != nil {
		panic(err2)
	}
	indexExists := false
	for _, v := range slice {
		if v == indexName {
			indexExists = true
			break
		}
	}
	if indexExists {
		result, err2 := client.FTDropIndex(ctx, indexName).Result()
		if err2 != nil {
			panic(err2)
		}
		fmt.Println(result)
	}
	if !indexExists {
		//创建索引
		result, err := client.FTCreate(ctx, indexName, &redis.FTCreateOptions{
			OnHash: true,
			Prefix: []any{keyPrefix},
		}, &redis.FieldSchema{
			FieldName: "content",
			FieldType: redis.SearchFieldTypeText,
			Weight:    1,
		}, &redis.FieldSchema{
			FieldName: "vector_content",
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				FlatOptions: &redis.FTFlatOptions{
					Type:           "FLOAT32",
					Dim:            2048,
					DistanceMetric: "COSINE",
				},
			},
		}).Result()
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	}
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("api_key"),
		Model:  os.Getenv("embedding_id"),
	})
	if err != nil {
		panic(err)
	}
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   10,                            // 必需：目标片段大小
		OverlapSize: 2,                             // 可选：片段重叠大小
		Separators:  []string{"\n", ".", "?", "！"}, // 可选：分隔符列表
		LenFunc:     nil,                           // 可选：自定义长度计算函数
		KeepType:    recursive.KeepTypeNone,        // 可选：分隔符保留策略
		IDGenerator: func(ctx context.Context, originalID string, splitIndex int) string {
			return fmt.Sprintf("%s_%d", originalID, splitIndex)
		},
	})
	if err != nil {
		panic(err)
	}
	docs, err := splitter.Transform(ctx, []*schema.Document{
		{
			ID: "testDoc",
			Content: `
			That is a very happy person。
            That is a happy dog。
            Today is a sunny day。`,
		},
	})
	if err != nil {
		panic(err)
	}
	indexer, err := ri.NewIndexer(ctx, &ri.IndexerConfig{
		Client:    client,
		KeyPrefix: keyPrefix,
		BatchSize: len(docs),
		Embedding: embedder,
		//DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*ri.Hashes, error) {
		//	fields := make(map[string]ri.FieldValue)
		//	fields[FieldContent] = ri.FieldValue{
		//		Value:    doc.Content,
		//		EmbedKey: FieldContentVector,
		//	}
		//	return &ri.Hashes{
		//		Key:         uuid.New().String(),
		//		Field2Value: fields,
		//	}, nil
		//},
	})
	if err != nil {
		panic(err)
	}
	ids, err := indexer.Store(ctx, docs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", ids)
}
