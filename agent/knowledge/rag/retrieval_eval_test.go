package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"awesomeProject4/Init"
	"github.com/cloudwego/eino-ext/components/retriever/milvus2"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

type retrievalAnswerSample struct {
	Question        string `json:"question"`
	Answer          string `json:"answer,omitempty"`
	GroundTruth     string `json:"ground_truth,omitempty"`
	ReferenceAnswer string `json:"reference_answer,omitempty"`
	Filter          string `json:"filter,omitempty"`
}

type retrievalAnswerResult struct {
	Question    string   `json:"question"`
	Contexts    []string `json:"contexts"`
	Answer      string   `json:"answer"`
	GroundTruth string   `json:"ground_truth"`
}

func TestEvaluateRetrievalMetrics(t *testing.T) {
	t.Parallel()

	samples := []RetrievalEvalSample{
		{Query: "q1", RelevantID: "a"},
		{Query: "q2", RelevantIDs: []string{"c", "d"}},
	}

	runner := func(_ context.Context, query string, _ int, _ string) ([]*schema.Document, error) {
		switch query {
		case "q1":
			return []*schema.Document{
				{ID: "a"},
				{ID: "x"},
			}, nil
		case "q2":
			return []*schema.Document{
				{ID: "x"},
				{ID: "c"},
				{ID: "y"},
			}, nil
		default:
			return nil, nil
		}
	}

	metrics, err := EvaluateRetrievalMetrics(context.Background(), samples, 5, runner)
	if err != nil {
		t.Fatal(err)
	}

	if metrics.Total != 2 {
		t.Fatalf("unexpected total: %d", metrics.Total)
	}
	if metrics.HitRateAt1 != 0.5 {
		t.Fatalf("unexpected hit@1: %.4f", metrics.HitRateAt1)
	}
	if metrics.HitRateAt5 != 1.0 {
		t.Fatalf("unexpected hit@5: %.4f", metrics.HitRateAt5)
	}
	if metrics.MRRAt1 != 0.5 {
		t.Fatalf("unexpected mrr@1: %.4f", metrics.MRRAt1)
	}
}

func TestExportRetrievalAnswers(t *testing.T) {
	if strings.TrimSpace(os.Getenv("RAG_EXPORT")) == "" {
		t.Skip("set RAG_EXPORT=1 to export doc/ans.json")
	}

	rootDir, err := findProjectRoot()
	if err != nil {
		t.Fatal(err)
	}
	restoreWorkingDir(t, rootDir)

	inputPath := filepath.Join(rootDir, "doc", "val.json")
	if rawPath := strings.TrimSpace(os.Getenv("RAG_EVAL_FILE")); rawPath != "" {
		inputPath = rawPath
	}

	outputPath := filepath.Join(rootDir, "doc", "ans.json")
	if rawPath := strings.TrimSpace(os.Getenv("RAG_EVAL_OUTPUT")); rawPath != "" {
		outputPath = rawPath
	}

	samples, err := loadRetrievalAnswerSamples(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) == 0 {
		t.Fatalf("evaluation file %q contains no samples", inputPath)
	}

	topK := 5
	if rawTopK := strings.TrimSpace(os.Getenv("RAG_EVAL_TOPK")); rawTopK != "" {
		parsed, err := strconv.Atoi(rawTopK)
		if err != nil {
			t.Fatalf("invalid RAG_EVAL_TOPK: %v", err)
		}
		if parsed > 0 {
			topK = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	manager := Init.NewMilvusManger()
	hybridRetriever := NewHybridRetriever(ctx, manager.Client)

	runner := func(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error) {
		opts := make([]retriever.Option, 0, 2)
		if topK > 0 {
			opts = append(opts, retriever.WithTopK(topK))
		}
		if strings.TrimSpace(filter) != "" {
			opts = append(opts, milvus2.WithFilter(filter))
		}
		return hybridRetriever.Retrieve(ctx, query, opts...)
	}

	results := make([]retrievalAnswerResult, 0, len(samples))
	for i, sample := range samples {
		question := strings.TrimSpace(sample.Question)
		if question == "" {
			t.Fatalf("sample %d: question is required", i)
		}

		docs, err := runner(ctx, question, topK, sample.Filter)
		if err != nil {
			t.Fatalf("sample %d query %q: %v", i, question, err)
		}

		answer := firstNonEmpty(sample.Answer, sample.GroundTruth, sample.ReferenceAnswer)
		groundTruth := firstNonEmpty(sample.GroundTruth, sample.Answer, sample.ReferenceAnswer)

		results = append(results, retrievalAnswerResult{
			Question:    question,
			Contexts:    extractContexts(docs),
			Answer:      answer,
			GroundTruth: groundTruth,
		})
	}

	if err := writeRetrievalAnswerResults(outputPath, results); err != nil {
		t.Fatal(err)
	}

	t.Logf("exported %d samples to %s", len(results), outputPath)
}

func loadRetrievalAnswerSamples(path string) ([]retrievalAnswerSample, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var samples []retrievalAnswerSample
	if err := json.Unmarshal(content, &samples); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return samples, nil
}

func writeRetrievalAnswerResults(path string, results []retrievalAnswerResult) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	payload, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0o644)
}

func extractContexts(docs []*schema.Document) []string {
	contexts := make([]string, 0, len(docs))
	seen := make(map[string]struct{}, len(docs))

	for _, doc := range docs {
		text := extractDocumentText(doc)
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		contexts = append(contexts, text)
	}

	return contexts
}

func extractDocumentText(doc *schema.Document) string {
	if doc == nil {
		return ""
	}

	if text := strings.TrimSpace(doc.Content); text != "" {
		return text
	}

	if doc.MetaData == nil {
		return ""
	}

	for _, key := range []string{"content", "text", "page_content", "chunk_text"} {
		value, ok := doc.MetaData[key]
		if !ok {
			continue
		}
		if text := strings.TrimSpace(fmt.Sprint(value)); text != "" {
			return text
		}
	}

	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root from %s", dir)
		}
		dir = parent
	}
}

func restoreWorkingDir(t *testing.T, dir string) {
	t.Helper()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(currentDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
}
