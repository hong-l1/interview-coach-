package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"awesomeProject4/Init"
	"awesomeProject4/agent/knowledge/rag"
	"awesomeProject4/agent/knowledge/ragkit"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("skip .env: %v", err)
	}
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "ingest":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: ragkit-cli ingest <path>")
			os.Exit(1)
		}
		path := os.Args[2]
		ctx := context.Background()
		manager := Init.NewMilvusManger()
		defer manager.Client.Close(context.Background())
		embedder := rag.NewEmbedder(ctx)
		indexer := rag.NewIndexer(ctx, embedder, manager.Client)
		// 支持目录或文件
		files, err := expandPaths(path)
		if err != nil {
			log.Fatal(err)
		}
		total := 0
		for _, f := range files {
			n, err := ragkit.Index(ctx, embedder, indexer, f)
			if err != nil {
				log.Printf("skip %s: %v", f, err)
				continue
			}
			total += n
			log.Printf("imported %s: %d chunks", f, n)
		}
		log.Printf("done: %d files, %d chunks", len(files), total)
	case "retrieve":
		fmt.Fprintln(os.Stderr, "retrieve: see Task 15")
		os.Exit(1)
	default:
		usage()
		os.Exit(1)
	}
}

func expandPaths(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	return filepath.Glob(filepath.Join(path, "*"))
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ragkit-cli <ingest|retrieve|health|activate|rollback|reindex> ...")
}
