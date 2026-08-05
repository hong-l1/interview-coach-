package chunking

import (
	"context"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

type SemanticResplit struct {
	embedder             embedding.Embedder
	minBlockRunes        int
	breakpointPercentile int
	enabled              bool
}

func NewSemanticResplit(embedder embedding.Embedder, minBlockRunes, breakpointPercentile int, enabled bool) *SemanticResplit {
	if minBlockRunes <= 0 {
		minBlockRunes = 1200
	}
	if breakpointPercentile <= 0 {
		breakpointPercentile = 20
	}
	return &SemanticResplit{
		embedder:             embedder,
		minBlockRunes:        minBlockRunes,
		breakpointPercentile: breakpointPercentile,
		enabled:              enabled,
	}
}

// Resplit 对超长 chunk 按句 embedding 相似度重切。enabled=false 时原样返回。
func (s *SemanticResplit) Resplit(ctx context.Context, chunks []*schema.Document) ([]*schema.Document, error) {
	if !s.enabled || s.embedder == nil {
		return chunks, nil
	}
	// TODO(enabled=true): 按句切分 → embed → 相邻 cosine → 分位断点重切。
	// 本轮面试文档小，默认关，留接口供 CLI 开启后实现。
	return chunks, nil
}
