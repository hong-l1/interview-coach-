package agents

import (
	"awesomeProject4/agent/chatmodel"
	schemas "awesomeProject4/agent/schema"
	"context"

	"github.com/cloudwego/eino/adk"
)

func NewEvaluateAllAgent(ctx context.Context) (adk.Agent, error) {
	question := schemas.NewResponseFormat(schemas.EvaluationAll{})
	model, err := chatmodel.NewChatModel(ctx, question)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "RecordEvaluationAgent",
		Description: "Evaluate interview records and generate a structured report",
		Model:       model,
		Instruction: `你是一名专业的面试评估专家。请根据面试记录输出结构化评估结果。

要求：
- 只返回 JSON，不要输出 markdown，不要输出解释。
- 顶层必须包含 comment 和 dimensions。
- dimensions 是数组。
- 每个 dimension 必须包含并且只包含以下字段：name、content、score。
- score 必须是 0 到 100 之间的整数。
- name 是中文维度名。
- content 是该维度的详细评估意见。
- comment 是整体评价与改进建议。
- 输出字段名必须与 schema 完全一致。`,
		MaxIterations: 7,
	})
}
