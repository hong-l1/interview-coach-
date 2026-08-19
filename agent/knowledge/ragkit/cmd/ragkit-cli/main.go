package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"awesomeProject4/Init"
	"awesomeProject4/agent/knowledge/rag"
	"awesomeProject4/agent/knowledge/ragkit"
	"awesomeProject4/agent/knowledge/ragkit/governance"
	"awesomeProject4/agent/knowledge/ragkit/ragkitdb"
	"awesomeProject4/agent/knowledge/ragkit/retrieval"

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
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: ragkit-cli retrieve <query>")
			os.Exit(1)
		}
		query := os.Args[2]
		ctx := context.Background()
		manager := Init.NewMilvusManger()
		defer manager.Client.Close(context.Background())
		searcher := retrieval.NewMilvusSearcher(manager.Client)
		// 读 active profile（若有 DB 则用，否则默认）
		res, err := ragkit.Retrieve(ctx, searcher, query, "", retrieval.DefaultRetrieveProfile())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("gate=%s reason=%s items=%d\n", res.EvidenceGate.Outcome, res.EvidenceGate.Reason, len(res.Items))
		for _, it := range res.Items {
			fmt.Printf("- [%.3f] %s\n", it.Score, truncate(it.Content, 200))
		}
	case "health":
		ctx := context.Background()
		manager := Init.NewMilvusManger()
		defer manager.Client.Close(context.Background())
		collection := os.Getenv("collection")
		h, err := governance.HealthCheck(ctx, manager.Client, collection)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ok=%v checks=%v gaps=%v\n", h.OK, h.Checks, h.Gaps)
	case "activate":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: ragkit-cli activate <profile_id>")
			os.Exit(1)
		}
		// 需要 DB：复用 InitMysql 或独立连接。简化：用 Init.InitMysql
		db, err := Init.InitMysql()
		if err != nil {
			log.Fatal(err)
		}
		ragkitdb.Migrate(db)
		store := governance.NewGormProfileStore(db)
		id := atoi(os.Args[2])
		p, err := store.Activate(context.Background(), uint64(id))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("activated profile: %s (id=%d)", p.Name, p.ID)
	case "rollback":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: ragkit-cli rollback <profile_id>")
			os.Exit(1)
		}
		// 同 activate，调 store.Rollback
		db, err := Init.InitMysql()
		if err != nil {
			log.Fatal(err)
		}
		ragkitdb.Migrate(db)
		store := governance.NewGormProfileStore(db)
		id := atoi(os.Args[2])
		p, err := store.Rollback(context.Background(), uint64(id))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("rolled back to: %s (id=%d)", p.Name, p.ID)
	case "reindex":
		fmt.Fprintln(os.Stderr, "reindex not implemented in this phase")
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

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("invalid number %q: %v", s, err)
	}
	return n
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ragkit-cli <ingest|retrieve|health|activate|rollback|reindex> ...")
}
