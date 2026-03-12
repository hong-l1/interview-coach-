package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/schema"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	am, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		Model:  os.Getenv("model_id"),
		APIKey: os.Getenv("ark_key"),
	})
	if err != nil {
		log.Fatalf("failed to create agentic model, err: %v", err)
	}
	input := []*schema.Message{
		schema.SystemMessage("你是一个AI助手，中文回答问题"),
		schema.UserMessage("今天是2026.03.12，今天重庆天气"),
	}
	msg, err := am.Generate(ctx, input)
	if err != nil {
		log.Fatalf("failed to generate message, err: %v", err)
	}
	fmt.Println(msg.Content)
}
