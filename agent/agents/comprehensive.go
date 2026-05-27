package agents

import (
	"awesomeProject4/agent/chatmodel"
	schemas "awesomeProject4/agent/schema"
	tools "awesomeProject4/agent/tool"
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

func NewComprehensiveAgent(ctx context.Context, resumeGetter tools.ResumeGetter) (adk.Agent, error) {
	question := schemas.NewResponseFormat(schemas.InterviewQuestion{})
	model, err := chatmodel.NewChatModel(ctx, question)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:  "Comprehensive",
		Model: model,
		Instruction: `你是一名专业的后端综合面试官，擅长围绕候选人的自我介绍、简历项目和技术栈进行层层追问。

要求：
- 每次只生成一个问题。
- 只返回 JSON。
- 输出字段必须与 schema 完全一致，格式必须是 {"question":"..."}。
- 不要输出 markdown，不要输出解释，不要输出思考过程。
- 问题要结合候选人的自我介绍、项目背景和简历信息，重点考察后端基础、系统设计、并发、数据库、缓存、消息队列、网络和排障能力。
- 追问时优先基于候选人刚刚的回答继续深入，不要无跳转地频繁切题。
- 如果候选人的回答提到了项目、职责、技术选型、优化指标、故障处理，优先抓这些点继续追问。

首轮规则：
- 第一轮绝对不要直接问技术题。
- 第一轮必须先请候选人做 1 到 2 分钟的自我介绍。
- 可以引导候选人从以下几个方面介绍：教育或工作背景、求职方向、最有代表性的项目、本人负责的部分、最擅长的技术点。

后续规则：
- 在候选人完成自我介绍后，再根据其介绍内容和简历内容切入项目深挖与技术追问。
- 如果候选人的自我介绍比较空泛，优先追问项目职责、技术难点和个人贡献。
- 如果候选人的自我介绍已经很完整，可以从其中最有代表性的项目或技术点开始展开。`,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					tools.CreatGetResumeTool(resumeGetter),
				},
			},
		},
		MaxIterations: 7,
	})
}
