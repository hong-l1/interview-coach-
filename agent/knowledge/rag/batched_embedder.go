package rag

import (
	"context"
	"time"

	"github.com/cloudwego/eino/components/embedding"
)

const defaultEmbeddingBatchSize = 128

type batchedEmbedder struct {
	inner     embedding.Embedder
	batchSize int
}

func NewBatchedEmbedder(inner embedding.Embedder, batchSize int) embedding.Embedder {
	if batchSize <= 0 {
		batchSize = defaultEmbeddingBatchSize
	}
	return &batchedEmbedder{
		inner:     inner,
		batchSize: batchSize,
	}
}

func (b *batchedEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...embedding.Option) ([][]float64, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	if len(texts) <= b.batchSize {
		return b.inner.EmbedStrings(ctx, texts, opts...)
	}

	vectors := make([][]float64, 0, len(texts))
	for start := 0; start < len(texts); start += b.batchSize {
		end := start + b.batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batchVectors, err := b.inner.EmbedStrings(ctx, texts[start:end], opts...)
		if err != nil {
			return nil, err
		}
		vectors = append(vectors, batchVectors...)
		time.Sleep(time.Second * 3)
	}
	return vectors, nil
}
