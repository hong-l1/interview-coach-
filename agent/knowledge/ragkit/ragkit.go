package ragkit

import (
	"context"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"awesomeProject4/agent/knowledge/ragkit/chunking"
	"awesomeProject4/agent/knowledge/ragkit/retrieval"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus2"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

// Version 是 ragkit 实现版本号，用于审计与元数据留痕。
const Version = "ragkit-v0"

// NewMilvusSearcher 便捷构造：用默认 Ark embedder + Milvus client 组装 Searcher，
// 供检索链路（ragkit.Retrieve）使用。
func NewMilvusSearcher(ctx context.Context, client *milvusclient.Client) retrieval.Searcher {
	return retrieval.NewMilvusSearcher(client, NewEmbedder(ctx))
}

// Retrieve 用 searcher 搜索 + 后处理（facade 入口）。
func Retrieve(ctx context.Context, searcher retrieval.Searcher, query, filter string, profile retrieval.RetrieveProfile) (retrieval.Result, error) {
	cand, _ := retrieval.DecideDynamicTopK(query, profile.TopK)
	docs, err := searcher.Search(ctx, query, cand, filter)
	if err != nil {
		return retrieval.Result{}, err
	}
	items := retrieval.ToItems(docs)
	return retrieval.PostProcess(ctx, query, items, profile)
}

// DefaultRouter 返回默认切块路由：table-aware → markdown 兜底。
func DefaultRouter() *chunking.StrategyRouter {
	return chunking.NewStrategyRouter(
		chunking.NewMarkdownStrategy(1000),
		chunking.RoutedStrategy{
			Name: chunking.StrategyTableAware,
			Match: func(req chunking.Request) bool { return len(req.Document.Tables) > 0 },
			Strategy: chunking.NewTableAwareStrategy(1000),
		},
	)
}

// NormalizeDocs 把 eino 原始文档标准化（每个 doc 一个 NormalizedDocument）。
func NormalizeDocs(docs []*schema.Document) []*canonical.NormalizedDocument {
	out := make([]*canonical.NormalizedDocument, 0, len(docs))
	for _, d := range docs {
		src := canonical.Source{FileName: fileNameFromMeta(d)}
		nd := canonical.NewNormalizedDocument(d.Content, src)
		canonical.Normalize(nd)
		out = append(out, nd)
	}
	return out
}

// Split 用给定 router 切块并补父子元数据。
func Split(ctx context.Context, nd *canonical.NormalizedDocument, router *chunking.StrategyRouter, documentID string) ([]*schema.Document, error) {
	chunks, err := router.Split(ctx, chunking.Request{Document: nd})
	if err != nil {
		return nil, err
	}
	chunks = canonical.AnnotateChunksWithProvenance(chunks, nd)
	chunks = chunking.FinalizeChunks(chunks, nd, documentID)
	return chunks, nil
}

// Index 是入库全链路 facade：load → normalize → split → enrich metadata（双SHA1）→ store。
// 返回入库 chunk 数。
func Index(ctx context.Context, embedder *ark.Embedder, indexer *milvus2.Indexer, docPath string) (int, error) {
	docs, err := NewDocumentsLoader(ctx, docPath)
	if err != nil {
		return 0, err
	}
	router := DefaultRouter()
	total := 0
	for _, d := range docs {
		nd := canonical.NewNormalizedDocument(d.Content, canonical.Source{FileName: d.ID})
		canonical.Normalize(nd)
		chunks, err := Split(ctx, nd, router, nd.Source.FileName)
		if err != nil {
			return total, err
		}
		chunks = enrichRagkitMetadata(chunks, nd)
		ids, err := indexer.Store(ctx, chunks)
		if err != nil {
			return total, err
		}
		total += len(ids)
	}
	return total, nil
}

// enrichRagkitMetadata 把双 SHA1 / applied_rules / ragkit 版本写入 chunk metadata。
func enrichRagkitMetadata(chunks []*schema.Document, nd *canonical.NormalizedDocument) []*schema.Document {
	for _, c := range chunks {
		if c.MetaData == nil {
			c.MetaData = map[string]any{}
		}
		c.MetaData["raw_sha1"] = nd.Canonicalization.RawSHA1
		c.MetaData["canonical_sha1"] = nd.Canonicalization.CanonicalSHA1
		c.MetaData["applied_rules"] = nd.Canonicalization.AppliedRules
		c.MetaData["ragkit_version"] = Version
	}
	return chunks
}

func fileNameFromMeta(d *schema.Document) string {
	if d == nil {
		return ""
	}
	if v, ok := d.MetaData["_file_name"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return d.ID
}
