package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"io"
	"net/http"
	"net/url"
	"os"
)

var apiKey = os.Getenv("weather_key")

type weatherTool struct{}

func (w *weatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "获取指定城市和日期的天气信息。例如：get_weather(city='北京', extensions='base')",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"city": {
				Type:     schema.String,
				Required: true,
				Desc:     "城市名称",
			},
			"extensions": {
				Desc: "气象类型: base(实况天气) / all(预报天气)",
				Type: schema.String,
				Enum: []string{"base", "all"},
			},
		}),
	}, nil
}
func (w *weatherTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var params map[string]any
	if err := json.Unmarshal([]byte(argumentsInJSON), &params); err != nil {
		return "", fmt.Errorf("failed to parse input: %w", err)
	}
	city, ok := params["city"].(string)
	if !ok || city == "" {
		return "", fmt.Errorf("city is required")
	}
	baseURL := "https://restapi.amap.com/v3/weather/weatherInfo"
	queryParams := url.Values{}
	queryParams.Set("city", city)
	queryParams.Set("key", apiKey)
	if extensions, ok := params["extensions"].(string); ok {
		queryParams.Set("extensions", extensions)
	} else {
		queryParams.Set("extensions", "base")
	}
	queryParams.Set("output", "JSON")
	fullURL := fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}
func main() {
	ctx := context.Background()
	var weatherTool *weatherTool
	toolNode, err := compose.NewToolNode(ctx, &compose.ToolsNodeConfig{
		Tools: []tool.BaseTool{weatherTool},
	})
	if err != nil {
		panic(err)
	}
	model, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
		APIKey: os.Getenv("api_key"),
		Model:  os.Getenv("model_id"),
	})
	if err != nil {
		panic(err)
	}
	weatherInfo, err := weatherTool.Info(ctx)
	if err != nil {
		panic(err)
	}
	toolCallingChatModel, err := model.WithTools([]*schema.ToolInfo{
		weatherInfo,
	})
	if err != nil {
		panic(err)
	}
	messages := prompt.FromMessages(schema.GoTemplate,
		schema.SystemMessage("你是一个AI助手，你必须调用工具来获取天气信息"),
		schema.UserMessage("我需要查询重庆今天的天气"),
	)
	vars := map[string]any{}
	result, err := messages.Format(ctx, vars)
	input, err := toolCallingChatModel.Generate(ctx, result)
	if err != nil {
		panic(err)
	}
	println(input.Content)
	toolMessage, err := toolNode.Invoke(ctx, input)
	if err != nil {
		panic(err)
	}
	for _, v := range toolMessage {
		println(v.Content)
	}
}
