package chatmodel

import (
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

var apikey = "sk-Eyw5mqfcjERUWmBuEt4rwzRIsKUr7B8OUwAuEYzuaCwgoqi2"
var model = "gpt-5.4"
var baseurl = "https://code.oai1.online/v1"

func NewChatModel(ctx context.Context, responseFormat *openai.ChatCompletionResponseFormat) (*openai.ChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:         apikey,
		Model:          model,
		BaseURL:        baseurl,
		ResponseFormat: responseFormat,
	})
}
