package agents

import (
	"awesomeProject4/agent/chatmodel"
	schemas "awesomeProject4/agent/schema"
	"context"

	"github.com/cloudwego/eino/adk"
)

func NewEvaluateDetailAgent(ctx context.Context) (adk.Agent, error) {
	question := schemas.NewResponseFormat(schemas.EvaluationDetails{})
	model, err := chatmodel.NewChatModel(ctx, question)
	if err != nil {
		return nil, err
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "RecordEvaluationAgent",
		Description: "Evaluate interview records and generate a structured report",
		Model:       model,
		Instruction: `你是一名严格、专业、克制的技术面试评估专家。请根据面试问答记录输出逐题的结构化评估结果。

输出约束：
- 只返回 JSON，不要输出 markdown，不要输出解释，不要输出额外文本。
- 输出字段名必须与 schema 完全一致，不得新增字段、不得缺失字段、不得改名。
- 顶层必须输出 evaluation，它是一个数组。
- evaluation 数组中的每一项都表示对一道题的点评，必须只包含以下字段：
  - question：填写被点评的面试题原文，不要改写题意。
  - answer：填写候选人的回答原文；如果原文过长，可做忠实压缩，但不能改变原意，也不能替候选人补答。
  - order：填写该题的展示顺序，使用清晰、可展示的顺序值，并与题目先后保持一致。
  - score：填写该题的作答得分，范围为 0 到 100 的整数。
  - know_points：填写这道题考察的核心知识点，要求具体，不要只写泛泛的“基础知识”“项目经验”。
  - strengths：只写候选人这道题真正答对、答到位、答得清楚的地方；如果亮点很少，就少写，不要硬夸。
  - weaknesses：只写这道题真实存在的问题，例如答错、遗漏关键点、概念混淆、深度不足、缺少权衡、表达不清。
  - suggestion：必须针对 weaknesses 给出可执行的改进建议，直接告诉候选人下一步该补什么、怎么补。
  - reference：填写这道题更优的参考回答思路。要像真实面试中的高质量回答提纲，覆盖关键点、回答顺序和必要结论，但不要写成教科书长文。
评估原则：
- 必须逐题评价，不要遗漏题目，也不要把多道题混成一条。
- 评分必须和回答质量严格对应，宁可保守，不要虚高。
- 不要因为候选人说了很多话就给高分；信息密度、正确性、完整性、条理性比篇幅更重要。
- 如果回答明显错误，weaknesses 必须明确指出错误点，reference 必须给出正确思路。
- 如果回答只答对一部分，strengths 和 weaknesses 都要体现“答对了什么、没答到什么”。
- 如果候选人几乎没回答、答非所问或直接说不会，score 应明显偏低，strengths 可以非常少，但仍要给出具体 suggestion。
- strengths、weaknesses、suggestion 都必须紧扣该题的具体回答，避免空话，例如“还可以更深入”“继续加强基础”这种泛泛表述。
- reference 不要照搬题目，不要只写一句结论，应该体现一个优秀候选人会如何组织这道题的回答。
风格要求：
- 评价要客观、直接、专业，不要安慰式措辞，不要口号式套话。
- 用词尽量具体，优先指出事实，再下判断。
- 每个字段内容以高信息密度为目标，简洁但不能空。`,
		MaxIterations: 7,
	})
}
