package governance

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// IndexHealth 记录集合健康检查结果。
type IndexHealth struct {
	OK     bool
	Checks map[string]bool
	Gaps   []string
}

// HealthCheck 校验集合存在、维度、度量、load 健康（简化：校验集合存在 + 描述）。
func HealthCheck(ctx context.Context, client *milvusclient.Client, collection string) (IndexHealth, error) {
	h := IndexHealth{Checks: map[string]bool{}}
	// eino-ext milvus client DescribeCollection 简化：调用失败即不健康
	// 实际实现按 milvusclient API：client.DescribeCollection(ctx, NewDescribeCollectionOption(collection))
	// 本轮先校验 client 非空 + collection 名非空，留 Gaps 记录未实现的契约项。
	if client == nil {
		return h, fmt.Errorf("milvus client is nil")
	}
	if collection == "" {
		h.Gaps = append(h.Gaps, "collection name empty")
		return h, nil
	}
	h.Checks["client_ok"] = true
	h.Checks["collection_named"] = true
	h.Gaps = append(h.Gaps, "describe_collection_not_implemented", "dimension_match_not_implemented")
	h.OK = len(h.Gaps) == 0
	return h, nil
}

// Reindex 占位：drop & recreate collection + 重跑入库。本轮留接口，CLI reindex 提示未实现。
func Reindex(ctx context.Context, client *milvusclient.Client, collection string) error {
	return fmt.Errorf("reindex not implemented in this phase")
}
