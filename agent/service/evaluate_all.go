package service

import (
	"awesomeProject4/agent/agents"
	agentSchema "awesomeProject4/agent/schema"
	"context"
	"encoding/json"
)

func evaluateAll(ctx context.Context, prompt string) (agentSchema.EvaluationAll, error) {
	agent, err := agents.NewEvaluateAllAgent(ctx)
	if err != nil {
		return agentSchema.EvaluationAll{}, err
	}
	msg, err := runAgentForJSON(ctx, agent, prompt)
	if err != nil {
		return agentSchema.EvaluationAll{}, err
	}
	var result agentSchema.EvaluationAll
	if err := json.Unmarshal([]byte(msg), &result); err != nil {
		return agentSchema.EvaluationAll{}, err
	}
	return result, nil
}
