package rag

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type RetrievalEvalSample struct {
	Query       string   `json:"query,omitempty"`
	Question    string   `json:"question,omitempty"`
	Filter      string   `json:"filter,omitempty"`
	RelevantIDs []string `json:"relevant_ids,omitempty"`
	RelevantID  string   `json:"relevant_id,omitempty"`
	GoldIDs     []string `json:"gold_ids,omitempty"`
	GoldID      string   `json:"gold_id,omitempty"`
	AnswerIDs   []string `json:"answer_ids,omitempty"`
	AnswerID    string   `json:"answer_id,omitempty"`
	DocumentIDs []string `json:"document_ids,omitempty"`
	DocumentID  string   `json:"document_id,omitempty"`
}

type RetrievalRunner func(ctx context.Context, query string, topK int, filter string) ([]*schema.Document, error)

type RetrievalMetrics struct {
	Total      int     `json:"total"`
	HitRateAt1 float64 `json:"hit_rate_at_1"`
	HitRateAt5 float64 `json:"hit_rate_at_5"`
	MRRAt1     float64 `json:"mrr_at_1"`
}

func EvaluateRetrievalMetrics(ctx context.Context, samples []RetrievalEvalSample, topK int, runner RetrievalRunner) (RetrievalMetrics, error) {
	if runner == nil {
		return RetrievalMetrics{}, fmt.Errorf("retrieval runner is required")
	}
	if topK <= 0 {
		topK = 5
	}

	var (
		total  int
		hitAt1 int
		hitAt5 int
		mrrAt1 float64
	)

	for i, rawSample := range samples {
		sample, err := normalizeRetrievalEvalSample(rawSample)
		if err != nil {
			return RetrievalMetrics{}, fmt.Errorf("sample %d: %w", i, err)
		}

		docs, err := runner(ctx, sample.query(), topK, sample.Filter)
		if err != nil {
			return RetrievalMetrics{}, fmt.Errorf("sample %d query %q: %w", i, sample.query(), err)
		}

		rank := firstRelevantRank(docs, sample.relevantSet())
		total++
		if rank == 1 {
			hitAt1++
			hitAt5++
			mrrAt1 += 1
			continue
		}
		if rank > 1 && rank <= 5 {
			hitAt5++
		}
	}

	if total == 0 {
		return RetrievalMetrics{}, fmt.Errorf("no valid samples to evaluate")
	}

	return RetrievalMetrics{
		Total:      total,
		HitRateAt1: float64(hitAt1) / float64(total),
		HitRateAt5: float64(hitAt5) / float64(total),
		MRRAt1:     mrrAt1 / float64(total),
	}, nil
}

func LoadRetrievalEvalSamples(path string) ([]RetrievalEvalSample, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return nil, fmt.Errorf("evaluation file is empty")
	}

	if strings.HasPrefix(trimmed, "[") {
		var samples []RetrievalEvalSample
		if err := json.Unmarshal([]byte(trimmed), &samples); err != nil {
			return nil, fmt.Errorf("decode json array: %w", err)
		}
		return samples, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(trimmed))
	scanner.Buffer(make([]byte, 1024), 1024*1024)

	samples := make([]RetrievalEvalSample, 0, 32)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var sample RetrievalEvalSample
		if err := json.Unmarshal([]byte(line), &sample); err != nil {
			return nil, fmt.Errorf("decode jsonl line %d: %w", lineNo, err)
		}
		samples = append(samples, sample)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return samples, nil
}

func normalizeRetrievalEvalSample(sample RetrievalEvalSample) (retrievalEvalNormalizedSample, error) {
	query := strings.TrimSpace(sample.Query)
	if query == "" {
		query = strings.TrimSpace(sample.Question)
	}
	if query == "" {
		return retrievalEvalNormalizedSample{}, fmt.Errorf("query/question is required")
	}

	relevantIDs := make([]string, 0, len(sample.RelevantIDs)+len(sample.GoldIDs)+len(sample.AnswerIDs)+len(sample.DocumentIDs)+4)
	relevantIDs = append(relevantIDs, sample.RelevantIDs...)
	relevantIDs = append(relevantIDs, sample.GoldIDs...)
	relevantIDs = append(relevantIDs, sample.AnswerIDs...)
	relevantIDs = append(relevantIDs, sample.DocumentIDs...)
	if sample.RelevantID != "" {
		relevantIDs = append(relevantIDs, sample.RelevantID)
	}
	if sample.GoldID != "" {
		relevantIDs = append(relevantIDs, sample.GoldID)
	}
	if sample.AnswerID != "" {
		relevantIDs = append(relevantIDs, sample.AnswerID)
	}
	if sample.DocumentID != "" {
		relevantIDs = append(relevantIDs, sample.DocumentID)
	}

	cleaned := make([]string, 0, len(relevantIDs))
	seen := make(map[string]struct{}, len(relevantIDs))
	for _, id := range relevantIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		cleaned = append(cleaned, id)
	}

	if len(cleaned) == 0 {
		return retrievalEvalNormalizedSample{}, fmt.Errorf("at least one relevant document id is required")
	}

	return retrievalEvalNormalizedSample{
		queryText:   query,
		Filter:      strings.TrimSpace(sample.Filter),
		relevantIDs: cleaned,
	}, nil
}

type retrievalEvalNormalizedSample struct {
	queryText   string
	Filter      string
	relevantIDs []string
}

func (s retrievalEvalNormalizedSample) query() string {
	return s.queryText
}

func (s retrievalEvalNormalizedSample) relevantSet() map[string]struct{} {
	result := make(map[string]struct{}, len(s.relevantIDs))
	for _, id := range s.relevantIDs {
		result[id] = struct{}{}
	}
	return result
}

func firstRelevantRank(docs []*schema.Document, relevant map[string]struct{}) int {
	for i, doc := range docs {
		if doc == nil {
			continue
		}

		if _, ok := relevant[documentID(doc)]; ok {
			return i + 1
		}
	}
	return 0
}

func documentID(doc *schema.Document) string {
	if doc == nil {
		return ""
	}

	id := strings.TrimSpace(doc.ID)
	if id != "" {
		return id
	}

	for _, key := range []string{"id", "doc_id", "docId", "document_id", "documentId"} {
		if doc.MetaData == nil {
			continue
		}
		value, ok := doc.MetaData[key]
		if !ok {
			continue
		}
		text := strings.TrimSpace(fmt.Sprint(value))
		if text != "" {
			return text
		}
	}

	return ""
}
