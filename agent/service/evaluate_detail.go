package service

import (
	"awesomeProject4/agent/agents"
	agentSchema "awesomeProject4/agent/schema"
	"context"
	"encoding/json"
)

func evaluateDetail(ctx context.Context, prompt string) (agentSchema.EvaluationDetails, error) {
	agent, err := agents.NewEvaluateDetailAgent(ctx)
	if err != nil {
		return agentSchema.EvaluationDetails{}, err
	}
	msg, err := runAgentForJSON(ctx, agent, prompt)
	if err != nil {
		return agentSchema.EvaluationDetails{}, err
	}
	var result agentSchema.EvaluationDetails
	if err := json.Unmarshal([]byte(msg), &result); err != nil {
		return agentSchema.EvaluationDetails{}, err
	}
	return result, nil
}
