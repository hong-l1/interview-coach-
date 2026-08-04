package chunking

import (
	"context"
	"testing"

	"awesomeProject4/agent/knowledge/ragkit/canonical"
	"github.com/cloudwego/eino/schema"
)

type fakeStrategy struct{ name string }

func (f fakeStrategy) Name() string { return f.name }
func (f fakeStrategy) Split(ctx context.Context, req Request) ([]*schema.Document, error) {
	return []*schema.Document{{ID: f.name, Content: "x", MetaData: map[string]any{}}}, nil
}

func TestRouterPicksFirstMatchAndTagsRoute(t *testing.T) {
	def := fakeStrategy{"default"}
	tbl := fakeStrategy{StrategyTableAware}
	md := fakeStrategy{StrategyMarkdown}
	router := NewStrategyRouter(md,
		RoutedStrategy{Name: StrategyTableAware, Match: func(req Request) bool { return len(req.Document.Tables) > 0 }, Strategy: tbl},
	)
	_ = def // 未使用
	doc := canonical.NewNormalizedDocument("raw", canonical.Source{})
	doc.Tables = []canonical.NormalizedTable{{Rows: []canonical.TableRow{{"a"}}}}
	chunks, err := router.Split(context.Background(), Request{Document: doc})
	if err != nil {
		t.Fatal(err)
	}
	if chunks[0].MetaData["chunking_route"] != StrategyTableAware {
		t.Fatalf("route tag = %v", chunks[0].MetaData["chunking_route"])
	}
}

func TestRouterFallsBackToDefault(t *testing.T) {
	md := fakeStrategy{StrategyMarkdown}
	router := NewStrategyRouter(md,
		RoutedStrategy{Name: StrategyTableAware, Match: func(req Request) bool { return len(req.Document.Tables) > 0 }, Strategy: fakeStrategy{StrategyTableAware}},
	)
	doc := canonical.NewNormalizedDocument("raw", canonical.Source{})
	chunks, err := router.Split(context.Background(), Request{Document: doc})
	if err != nil {
		t.Fatal(err)
	}
	if chunks[0].MetaData["chunking_route"] != StrategyMarkdown {
		t.Fatalf("route tag = %v", chunks[0].MetaData["chunking_route"])
	}
}
