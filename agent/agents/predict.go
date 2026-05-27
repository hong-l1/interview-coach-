package agents

import (
	"awesomeProject4/agent/chatmodel"
	schemas "awesomeProject4/agent/schema"
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

func NewPredictAgent(ctx context.Context) (adk.Agent, error) {
	question := schemas.NewResponseFormat(schemas.PredictionResult{})
	model, err := chatmodel.NewChatModel(ctx, question)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "PredictionAgent",
		Description: "Generate targeted practice questions from interview weaknesses",
		Model:       model,
		Instruction: `你是一名资深技术面试官。请根据候选人在面试中暴露出的薄弱点，生成 5 道针对性的强化练习题。

要求：
- 必须严格生成 5 道题。
- 只返回 JSON，不要输出 markdown，不要输出解释。
- 顶层字段必须是 predict_questions。
- 每一题必须包含并且只包含以下字段：question、focus、reference_answer、follow_up。
- question 是题目本身。
- focus 是该题考察的知识点或能力点。
- reference_answer 是简洁但可靠的参考答案。
- follow_up 是基于该题继续深挖的追问。
- 题目难度应逐步提升。`,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{},
			},
		},
		MaxIterations: 7,
	})
}
