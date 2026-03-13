package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	ri "github.com/cloudwego/eino-ext/components/indexer/redis"
	ark2 "github.com/cloudwego/eino-ext/components/model/ark"
	rr "github.com/cloudwego/eino-ext/components/retriever/redis"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strings"
)

const (
	redisIndexName    = "test_index"
	redisKeyPrefix    = "test_doc:"
	knowledgeFilePath = "./knowledge_base.txt"
)

func prepareKnowledge(ctx context.Context, rdb *redis.Client) error {
	//1. 检查索引是否存在，不存在则创建
	if err := createRedisIndexIfNotExists(ctx, rdb); err != nil {
		return err
	}
	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
	})
	if err != nil {
		return err
	}
	docs, err := loader.Load(ctx, document.Source{
		URI: knowledgeFilePath,
	})
	if err != nil {
		return err
	}
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   200, // 每个块的目标大小
		OverlapSize: 20,  // 块之间的重叠大小，以保持上下文连续性
	})
	if err != nil {
		return err
	}
	splitDocs, err := splitter.Transform(ctx, docs)
	if err != nil {
		return err
	}
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("api_key"),
		Model:  os.Getenv("embedding_id"),
	})
	if err != nil {
		return err
	}
	indexer, err := ri.NewIndexer(ctx, &ri.IndexerConfig{
		Client:    rdb,
		KeyPrefix: redisKeyPrefix,
		Embedding: embedder,
	})
	if err != nil {
		return err
	}
	ids, err := indexer.Store(ctx, splitDocs)
	if err != nil {
		return err
	}
	fmt.Printf("%v", ids)
	return nil
}

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:          "43.142.57.35:6379",
		UnstableResp3: true,
		Protocol:      2,
	})
	if err := prepareKnowledge(ctx, rdb); err != nil {
		panic(err)
	}
	//阶段2
	userQuestion := "Eino 框架是什么？它的吉祥物是什么？"
	if err := answerQuestion(ctx, rdb, userQuestion); err != nil {
		panic(err)
	}
}

func answerQuestion(ctx context.Context, rdb *redis.Client, question string) error {
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("api_key"),
		Model:  os.Getenv("embedding_id"),
	})
	if err != nil {
		return err
	}
	retriever, err := rr.NewRetriever(ctx, &rr.RetrieverConfig{
		Client:    rdb,
		Index:     redisIndexName,
		Embedding: embedder,
		TopK:      3,
	})
	if err != nil {
		return err
	}
	retrievedDocs, err := retriever.Retrieve(ctx, question)
	if err != nil {
		return err
	}
	if len(retrievedDocs) == 0 {
		log.Println("未能从知识库中找到相关信息。")
		return nil
	}
	var contextBuilder strings.Builder
	for i, doc := range retrievedDocs {
		log.Printf("  - 相关片段 %d: %s\n", i+1, strings.ReplaceAll(doc.Content, "\n", " "))
		contextBuilder.WriteString(doc.Content)
		contextBuilder.WriteString("\n\n")
	}
	ragTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(`
			你是一个智能助手。请根据下面提供的上下文信息来回答用户的问题。请确保你的回答完全基于所给的上下文，不要使用任何外部知识。如果上下文中没有足够信息来回答问题，请直接说“根据所提供的信息，我无法回答该问题。
			--- 上下文 ---
			{context}
		`),
		schema.UserMessage("{question}"),
	)
	var vars = map[string]any{
		"question": question,
		"context":  contextBuilder.String(),
	}
	messages, err := ragTemplate.Format(ctx, vars)
	if err != nil {
		return err
	}
	log.Println("--- 最终发送给 LLM 的提示词 ---")
	for _, msg := range messages {
		log.Printf("[%s]: %s\n", msg.Role, msg.Content)
	}
	log.Println("---------------------------------")
	chatModel, err := ark2.NewChatModel(ctx, &ark2.ChatModelConfig{
		Model:  os.Getenv("model_id"),
		APIKey: os.Getenv("api_key"),
	})
	if err != nil {
		return err
	}
	stream, err := chatModel.Stream(ctx, messages)
	if err != nil {
		return err
	}
	defer stream.Close()
	log.Println("\n--- AI 回答 ---")
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break // 流结束或发生错误
		}
		print(chunk.Content)
	}
	println() // 换行
	return nil
}
func createRedisIndexIfNotExists(ctx context.Context, rdb *redis.Client) error {
	indices, err := rdb.Do(ctx, "FT._LIST").StringSlice()
	if err != nil {
		return err
	}
	for _, v := range indices {
		if v == redisIndexName {
			return nil
		}
	}
	_, err = rdb.FTCreate(ctx, redisIndexName, &redis.FTCreateOptions{
		OnHash: true,
		Prefix: []any{redisKeyPrefix},
	}, &redis.FieldSchema{FieldName: "content", FieldType: redis.SearchFieldTypeText},
		&redis.FieldSchema{
			FieldName: "vector_content",
			FieldType: redis.SearchFieldTypeVector,
			VectorArgs: &redis.FTVectorArgs{
				FlatOptions: &redis.FTFlatOptions{
					Type:           "FLOAT32",
					Dim:            2048,
					DistanceMetric: "L2",
				},
			},
		},
	).Result()
	if err != nil {
		return err
	}
	return nil
}
