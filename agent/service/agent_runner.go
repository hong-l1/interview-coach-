package service

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func runAgentForJSON(ctx context.Context, agent adk.Agent, prompt string) (string, error) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})
	var msg string
	itor := runner.Run(ctx, []*schema.Message{schema.SystemMessage(prompt)})
	for {
		event, ok := itor.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", event.Err
		}
		if event.Output.MessageOutput != nil && event.Output.MessageOutput.Message != nil {
			msg = event.Output.MessageOutput.Message.Content
		}
	}
	return msg, nil
}
