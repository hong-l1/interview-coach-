package rag

import (
	"context"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/components/document"
	"unicode/utf8"
)

func NewSemanticSplit(ctx context.Context, embedder *ark.Embedder) document.Transformer {
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    embedder,
		BufferSize:   1,
		MinChunkSize: 600,
		Percentile:   0.8,
		Separators: []string{
			"\n\n",
			"\n",
			"。",
			"！",
			"？",
			"；",
			"：",
		},
		LenFunc: func(s string) int {
			return utf8.RuneCountInString(s)
		},
	})
	if err != nil {
		panic(err)
	}
	return splitter
}
func NewRecursiveSplit(ctx context.Context, embedder *ark.Embedder) document.Transformer {
	recursive, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   384,
		OverlapSize: 64,
		Separators:  []string{"\n\n", "\n", "。", "？", "！", "；", "："},
		LenFunc: func(s string) int {
			return utf8.RuneCountInString(s)
		},
	})
	if err != nil {
		panic(err)
	}
	return recursive
}
