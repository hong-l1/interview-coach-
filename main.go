package main

import (
	"awesomeProject4/backend/event"
	"context"
	"github.com/IBM/sarama"
	"log"
)

func main() {
	server, consumerGroup, evaluationConsumer, err := InitializeServer()
	if err != nil {
		log.Fatalf("initialize server failed: %v", err)
	}
	go runInterviewEvaluationConsumer(context.Background(), consumerGroup, evaluationConsumer)
	if err := server.Run(":8080"); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}

func runInterviewEvaluationConsumer(ctx context.Context, consumerGroup sarama.ConsumerGroup, consumer *event.InterviewEvaluationConsumer) {
	for {
		if err := consumerGroup.Consume(ctx, []string{event.InterviewEvaluationTopic}, consumer); err != nil {
			log.Printf("consume interview evaluation event failed: %v", err)
		}
		if ctx.Err() != nil {
			return
		}
	}
}
