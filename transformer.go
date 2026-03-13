package main

import (
	"context"
	_ "embed"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   10,                                      // 必需：目标片段大小
		OverlapSize: 2,                                       // 可选：片段重叠大小
		Separators:  []string{"\n", "，", "?", "！", "!", "。"}, // 可选：分隔符列表
		LenFunc:     nil,                                     // 可选：自定义长度计算函数
		KeepType:    recursive.KeepTypeNone,                  // 可选：分隔符保留策略
	})
	if err != nil {
		panic(err)
	}
	docs, err := splitter.Transform(ctx, []*schema.Document{
		{
			ID: "1",
			Content: `
			这是第一个段落，包含了一些内容。
            
            这是第二个段落。这个段落有多个句子！这些句子通过标点符号分隔。
            
            这是第三个段落。这里有更多的内容。`,
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
