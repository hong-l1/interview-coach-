package rag

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"github.com/google/uuid"
	"path/filepath"
	"time"

	"github.com/cloudwego/eino/schema"
)

type DocumentMetaData struct {
	FileName    string `json:"file_name"`
	Source      string `json:"source"`
	ChunkIndex  int    `json:"chunk_index,omitempty"`
	TotalChunks int    `json:"total_chunks,omitempty"`
	ContentHash string `json:"content_hash,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func NewDocumentMetaData(filename, source string) *DocumentMetaData {
	now := time.Now()
	baseName := filepath.Base(filename)
	return &DocumentMetaData{
		FileName:  baseName,
		Source:    source,
		CreatedAt: now.Format(time.RFC3339),
	}
}

func (m *DocumentMetaData) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"file_name":    m.FileName,
		"created_at":   m.CreatedAt,
		"source":       m.Source,
		"chunk_index":  m.ChunkIndex,
		"total_chunks": m.TotalChunks,
		"content_hash": m.ContentHash,
	}
	return result
}

func EnrichDocumentsWithMetadata(ctx context.Context, docs []*schema.Document, metadata *DocumentMetaData) []*schema.Document {
	if metadata == nil {
		return docs
	}
	totalChunks := len(docs)
	for i, chunk := range docs {
		chunkMetadata := *metadata
		chunkMetadata.ChunkIndex = i
		chunkMetadata.TotalChunks = totalChunks
		chunkMetadata.ContentHash = hashContent(chunk.Content)
		if chunk.MetaData == nil {
			chunk.MetaData = make(map[string]interface{})
		} else {
			removeFrameworkMetadata(chunk.MetaData)
		}
		docs[i].MetaData = chunkMetadata.ToMap()
		docs[i].ID = uuid.New().String()
	}
	return docs
}
func hashContent(content string) string {
	sum := sha1.Sum([]byte(content))
	return hex.EncodeToString(sum[:])
}
func removeFrameworkMetadata(meta map[string]interface{}) {
	delete(meta, "_source")
	delete(meta, "_file_name")
	delete(meta, "_extension")
}
