package governance

import (
	"context"
	"testing"

	"awesomeProject4/agent/knowledge/ragkit/ragkitdb"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := ragkitdb.Migrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestProfileStoreActivateSwitchesActive(t *testing.T) {
	db := newTestDB(t)
	store := NewGormProfileStore(db)
	ctx := context.Background()

	// 建一个 baseline + 一个 candidate
	base := &StrategyProfile{Name: "base", Status: "active", TopKConfig: TopKConfigJSON{BaseCandidateTopK: 10, BaseFinalTopK: 5}}
	cand := &StrategyProfile{Name: "cand", Status: "candidate", TopKConfig: TopKConfigJSON{BaseCandidateTopK: 15, BaseFinalTopK: 5}}
	if err := store.Create(ctx, base); err != nil {
		t.Fatal(err)
	}
	if err := store.Create(ctx, cand); err != nil {
		t.Fatal(err)
	}

	activated, err := store.Activate(ctx, cand.ID)
	if err != nil {
		t.Fatal(err)
	}
	if activated.Status != "active" {
		t.Fatalf("cand should be active, got %s", activated.Status)
	}
	cur, _ := store.GetActive(ctx)
	if cur.Name != "cand" {
		t.Fatalf("active should be cand, got %s", cur.Name)
	}
}

func TestProfileStoreRollbackToBaseline(t *testing.T) {
	db := newTestDB(t)
	store := NewGormProfileStore(db)
	ctx := context.Background()
	base := &StrategyProfile{Name: "base", Status: "active"}
	cand := &StrategyProfile{Name: "cand", Status: "candidate"}
	store.Create(ctx, base)
	store.Create(ctx, cand)
	store.Activate(ctx, cand.ID) // 现在 cand active, base baseline
	rolled, err := store.Rollback(ctx, cand.ID)
	if err != nil {
		t.Fatal(err)
	}
	if rolled.Name != "base" {
		t.Fatalf("rollback should restore base, got %s", rolled.Name)
	}
}
