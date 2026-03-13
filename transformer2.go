package main

import (
	"context"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/schema"
	"os"
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
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    embedder,                      // 必需：用于生成文本向量的嵌入器
		BufferSize:   2,                             // 可选：上下文缓冲区大小
		MinChunkSize: 100,                           // 可选：最小片段大小
		Separators:   []string{"\n", ".", "?", "！"}, // 可选：分隔符列表
		Percentile:   0.9,                           // 可选：分割阈值百分位数
		LenFunc:      nil,                           // 可选：自定义长度计算函数
	})
	if err != nil {
		panic(err)
	}
	docs, err := splitter.Transform(ctx, []*schema.Document{
		{
			ID: "1",
			Content: `这是第一段内容，包含了一些重要信息。
            这是第二段内容，与第一段语义相关。
            这是第三段内容，主题已经改变。
            这是第四段内容，继续讨论新主题。`,
		},
	})
	if err != nil {
		panic(err)
	}
	for _, doc := range docs {
		println(doc.String())
		println("=========================")
	}
}
