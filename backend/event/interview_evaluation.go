package event

import (
	agentservice "awesomeProject4/agent/service"
	backendservice "awesomeProject4/backend/service"
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
)

const InterviewEvaluationTopic = "interview.evaluation.requested"

type InterviewEvaluationRequested struct {
	SessionID string `json:"session_id"`
	UserID    uint   `json:"user_id"`
	RecordID  uint64 `json:"record_id"`
	ReportID  uint64 `json:"report_id"`
}

type InterviewEvaluationProducer struct {
	producer sarama.SyncProducer
}

func NewInterviewEvaluationProducer(producer sarama.SyncProducer) *InterviewEvaluationProducer {
	return &InterviewEvaluationProducer{
		producer: producer,
	}
}

func (p *InterviewEvaluationProducer) PublishInterviewEvaluation(ctx context.Context, event InterviewEvaluationRequested) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: InterviewEvaluationTopic,
		Key:   sarama.StringEncoder(event.SessionID),
		Value: sarama.ByteEncoder(payload),
	}
	_, _, err = p.producer.SendMessage(msg)
	return err
}

type InterviewEvaluationConsumer struct {
	dialogueService         *backendservice.InterviewDialogueService
	evaluationDetailService *backendservice.EvaluationDetailService
	predictionService       *backendservice.PredictionService
}

func NewInterviewEvaluationConsumer(
	dialogueService *backendservice.InterviewDialogueService,
	evaluationDetailService *backendservice.EvaluationDetailService,
	predictionService *backendservice.PredictionService,
) *InterviewEvaluationConsumer {
	return &InterviewEvaluationConsumer{
		dialogueService:         dialogueService,
		evaluationDetailService: evaluationDetailService,
		predictionService:       predictionService,
	}
}

func (c *InterviewEvaluationConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *InterviewEvaluationConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *InterviewEvaluationConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event InterviewEvaluationRequested
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			return fmt.Errorf("decode interview evaluation event: %w", err)
		}
		if err := agentservice.EvaluationAndPredictService(
			session.Context(),
			event.UserID,
			event.RecordID,
			event.ReportID,
			c.dialogueService,
			c.evaluationDetailService,
			c.predictionService,
		); err != nil {
			return err
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
