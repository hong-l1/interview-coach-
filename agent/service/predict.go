package service

import (
	"awesomeProject4/agent/agents"
	agentSchema "awesomeProject4/agent/schema"
	"context"
	"encoding/json"
)

func predict(ctx context.Context, prompt string) (agentSchema.PredictionResult, error) {
	agent, err := agents.NewPredictAgent(ctx)
	if err != nil {
		return agentSchema.PredictionResult{}, err
	}
	msg, err := runAgentForJSON(ctx, agent, prompt)
	if err != nil {
		return agentSchema.PredictionResult{}, err
	}
	var result agentSchema.PredictionResult
	if err := json.Unmarshal([]byte(msg), &result); err != nil {
		return agentSchema.PredictionResult{}, err
	}
	return result, nil
}
