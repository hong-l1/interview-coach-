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

func NewResumeAgent(ctx context.Context) (adk.Agent, error) {
	resume := schemas.NewResponseFormat(schemas.Resume{})
	model, err := chatmodel.NewChatModel(ctx, resume)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "ResumeParserAgent",
		Description: "Parse a PDF resume and output structured JSON.",
		Instruction: `你是一个专业的简历解析助手。你的任务是把 PDF 简历解析为结构化 JSON。

工作流程：
1. 先调用 pdf_to_text 工具读取 PDF 全文。
2. 根据简历原文抽取信息。
3. 只输出最终 JSON，不要输出 markdown，不要输出解释。

抽取要求：
- base_info 中尽量提取：name、gender、age、work_years、job_intention、expected_city、email、phone。
- education 必须按时间或简历顺序完整提取每段教育经历，包含 school、degree、major、start、end。
- work_experience 用于提取正式工作经历；如果简历没有明确工作经历，返回空数组。
- projects 必须完整提取项目经历，包含 name、role、duration、description、contribution、tech_stack。
- skills 只保留明确的个人技能、框架、中间件、数据库、工程能力关键词，不要把整句项目描述塞进去。
- certifications 只放证书或正式认证；竞赛奖项、荣誉可放到 other。
- other 放不适合归类但有价值的信息，例如荣誉奖项。

输出时要尽量忠实于原文：
- 原文明确出现的信息尽量不要遗漏。
- 原文没有出现的信息不要编造。
- 缺失字段返回空字符串或空数组。`,
		Model:         model,
		MaxIterations: 7,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{tools.CreatParseResumeTool()},
			},
		},
	})
}
