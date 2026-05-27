package rag

import (
	"context"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus2"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"os"
)

func NewIndexer(ctx context.Context, embedder *ark.Embedder, client *milvusclient.Client) *milvus2.Indexer {
	indexer, err := milvus2.NewIndexer(ctx, &milvus2.IndexerConfig{
		Client:     client,
		Embedding:  embedder,
		Collection: os.Getenv("collection"),
		Sparse: &milvus2.SparseVectorConfig{
			VectorField:  "sparse_vector",
			MetricType:   milvus2.BM25,
			Method:       milvus2.SparseMethodAuto,
			IndexBuilder: milvus2.NewSparseInvertedIndexBuilder().WithDropRatioBuild(0.2),
		},
		Vector: &milvus2.VectorConfig{
			Dimension:    2048,
			MetricType:   milvus2.COSINE,
			IndexBuilder: milvus2.NewHNSWIndexBuilder().WithM(16).WithEfConstruction(200),
			VectorField:  "vector",
		},
		FieldParams: map[string]map[string]string{
			"content": {
				"enable_analyzer": "true",
				"analyzer_params": `{"type":"chinese"}`,
			},
		},
		EnableDynamicSchema: true,
	})
	if err != nil {
		panic(err)
	}
	return indexer
}
