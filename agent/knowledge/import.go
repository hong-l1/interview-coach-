package main

import (
	"awesomeProject4/Init"
	"awesomeProject4/agent/knowledge/ragkit"
	"context"
	"github.com/joho/godotenv"
	"log"
	"path/filepath"
	"sort"
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

	embedder := ragkit.NewEmbedder(ctx)
	indexer := ragkit.NewIndexer(ctx, embedder, manager.Client)

	totalChunks := 0
	for _, path := range files {
		name := filepath.Base(path)
		log.Printf("processing %s", name)

		n, err := ragkit.Index(ctx, embedder, indexer, path)
		if err != nil {
			log.Printf("skip %s: %v", name, err)
			continue
		}
		log.Printf("imported %s: %d chunks", name, n)
		totalChunks += n
	}
	log.Printf("done: %d files, %d chunks", len(files), totalChunks)
	log.Printf("flush completed")
}
