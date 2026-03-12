package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

func main() {
	//ctx := context.Background()
	//am, err := ark.NewChatModel(ctx, &ark.ChatModelConfig{
	//	Model:  os.Getenv("model_id"),
	//	APIKey: os.Getenv("ark_key"),
	//})
	//if err != nil {
	//	log.Fatalf("failed to create agentic model, err: %v", err)
	//}
	//template := prompt.FromMessages(
	//	schema.FString,
	//	schema.SystemMessage("你是一个{role}, 请用{tone}的语气回答问题"),
	//	schema.UserMessage("{question}"),
	//)
	//vars := map[string]any{
	//	"role":     "技术专家",
	//	"tone":     "专业严谨",
	//	"question": "如何优化数据库性能",
	//}
	//result, err := template.Format(ctx, vars)
	//if err != nil {
	//	log.Fatalf("failed to format template, err: %v", err)
	//}
	//resder, err := am.Stream(ctx, result)
	//if err != nil {
	//	log.Fatalf("failed to generate message, err: %v", err)
	//}
	//for {
	//	recv, err := resder.Recv()
	//	if err != nil {
	//		log.Fatalf("failed to generate message, err: %v", err)
	//	}
	//	if err == io.EOF {
	//		break
	//	}
	//	fmt.Printf(recv.Content)
	//}
	template := prompt.FromMessages(
		schema.GoTemplate,
		schema.SystemMessage("{{if .isExpert}}你是一个专家级{{.domain}}顾问。{{else}}你是一个初级{{.domain}}助手。{{end}}\n{{if .isFormal}}请使用正式的语言风格。{{else}}请使用友好的语言风格。{{end}}\n你的任务是{{.task}}。"),
		schema.UserMessage("{{.question}}"),
		schema.MessagesPlaceholder("history", false),
	)
	vars := map[string]interface{}{
		"isExpert": false,
		"domain":   "编程",
		"isFormal": false,
		"task":     "帮助初学者理解编程概念",
		"question": "什么是变量？",
		"history":  "<UNK>",
	}
	fmt.Println(template.Format(context.TODO(), vars))
}
