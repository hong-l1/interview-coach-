package chunking

import (
	"context"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"github.com/cloudwego/eino/schema"
)

const (
	StrategyMarkdown  = "markdown"
	StrategyTableAware = "table_aware"
	StrategyStructure = "structure_aware" // 留接口，本轮不实现
	StrategyOCRAware  = "ocr_aware"      // 留接口，本轮不实现
)

// Request 是切块请求。
type Request struct {
	Document       *canonical.NormalizedDocument
	BaseMeta       map[string]any
	NormalizedPath string
}

// Strategy 是切块策略接口。
type Strategy interface {
	Split(ctx context.Context, req Request) ([]*schema.Document, error)
	Name() string
}
