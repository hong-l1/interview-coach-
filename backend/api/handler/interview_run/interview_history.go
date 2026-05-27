package interview_run

import (
	"awesomeProject4/backend/api/constant"
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

type historyMessage struct {
	Role    string `json:"role"`
	Type    string `json:"type"`
	Index   int    `json:"index"`
	Content string `json:"content"`
}

func (r *InterviewRuntime) SaveHistory(ctx context.Context, sessionID string, msg historyMessage) error {
	redisClient := r.client
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	key := InterviewMsgListKey(sessionID)
	if err := redisClient.RPush(ctx, key, payload).Err(); err != nil {
		return err
	}
	return redisClient.Expire(ctx, key, constant.InterviewMsgListTTl).Err()
}

func (r *InterviewRuntime) loadHistory(ctx context.Context, sessionID string) ([]historyMessage, error) {
	redisClient := r.client
	val, err := redisClient.LRange(ctx, InterviewMsgListKey(sessionID), constant.Start, constant.End).Result()
	if err != nil {
		return nil, err
	}
	history := make([]historyMessage, 0, len(val))
	for _, val := range val {
		var msg historyMessage
		if err := json.Unmarshal([]byte(val), &msg); err != nil {
			continue
		}
		history = append(history, msg)
	}
	return history, nil
}

func historyMessagesToSchemaMessages(records []historyMessage) []*schema.Message {
	messages := make([]*schema.Message, 0, 5*2)
	for _, record := range records {
		if record.Role == "assistant" {
			messages = append(messages, schema.AssistantMessage(record.Content, nil))
		}
		if record.Role == "user" {
			messages = append(messages, schema.UserMessage(record.Content))
		}
	}
	return messages
}

func InterviewMsgListKey(sessionID string) string {
	return fmt.Sprintf("mianshi:session:%s:msgs", sessionID)
}
