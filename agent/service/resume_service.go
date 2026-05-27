package service

import (
	"awesomeProject4/agent/agents"
	schemas "awesomeProject4/agent/schema"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func ParseResumeService(ctx context.Context, filepath string) (*schemas.Resume, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	agent, err := agents.NewResumeAgent(ctx)
	if err != nil {
		return nil, errors.New("Failed to create resume agent: " + err.Error())
	}
	runner := adk.NewRunner(timeoutCtx, adk.RunnerConfig{
		Agent: agent,
	})
	quotedPath, err := json.Marshal(filepath)
	if err != nil {
		return nil, errors.New("Failed to marshal file path: " + err.Error())
	}
	prompt := fmt.Sprintf("Parse this PDF resume file. File path: %s", quotedPath)
	iter := runner.Run(timeoutCtx, []adk.Message{schema.UserMessage(prompt)})
	var msg string
	for {
		t, ok := iter.Next()
		if !ok {
			break
		}
		if t.Err != nil {
			return nil, fmt.Errorf("agent execution failed: %w", t.Err)
		}
		if t.Err == nil && t.Output.MessageOutput.Message != nil {
			msg = t.Output.MessageOutput.Message.Content
		}
	}
	if msg == "" {
		return nil, errors.New("agent execution failed")
	}
	var temp *schemas.Resume
	if err := json.Unmarshal([]byte(msg), &temp); err != nil {
		return nil, errors.New("Failed to parse resume message: " + err.Error())
	}
	return temp, nil
}
