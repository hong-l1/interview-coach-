package main

import (
	"awesomeProject4/Init"
	"awesomeProject4/agent/knowledge/rag"
	"context"
	"log"
	"path/filepath"
	"sort"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	manager := Init.NewMilvusManger()
	defer manager.Client.Close(context.Background())

	files, err := filepath.Glob(filepath.Join("doc", "*.docx"))
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(files)

	embedder := rag.NewEmbedder(ctx)
	indexer := rag.NewIndexer(ctx, embedder, manager.Client)
	// === ragkit 接线点（标准化 + 路由切块），默认关闭，手动切换时启用 ===
	// 启用方式：设置环境变量 RAGKIT_ENABLED=1，并替换下方 splitter.Transform 为 ragkit.Split
	// _ = ragkit.NormalizeDocs(docs)
	// _ = ragkit.DefaultRouter()
	splitter := rag.NewRecursiveSplit(ctx, embedder)

	totalChunks := 0
	for _, path := range files {
		name := filepath.Base(path)
		log.Printf("processing %s", name)

		docs, err := rag.NewDocumentsLoader(ctx, path)
		if err != nil {
			log.Printf("skip %s: %v", name, err)
			continue
		}

		chunks, err := splitter.Transform(ctx, docs)
		if err != nil {
			log.Printf("skip %s: %v", name, err)
			continue
		}

		meta := rag.NewDocumentMetaData(name, "cv_paper_eval")
		chunks = rag.EnrichDocumentsWithMetadata(ctx, chunks, meta)

		ids, err := indexer.Store(ctx, chunks)
		if err != nil {
			log.Printf("skip %s: %v", name, err)
			continue
		}

		totalChunks += len(ids)
		log.Printf("imported %s: %d chunks", name, len(ids))
	}
	log.Printf("done: %d files, %d chunks", len(files), totalChunks)
	log.Printf("flush completed")
}
