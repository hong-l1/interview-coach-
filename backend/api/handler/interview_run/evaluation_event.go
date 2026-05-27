package interview_run

import (
	"awesomeProject4/backend/api/constant"
	"awesomeProject4/backend/event"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func PublishEvaluationOnce(ctx context.Context, client redis.Cmdable, publisher EvaluationEventPublisher, payload event.InterviewEvaluationRequested) error {
	if publisher == nil {
		return nil
	}
	key := fmt.Sprintf("mianshi:session:%s:evaluation_sent", payload.SessionID)
	ok, err := client.SetNX(ctx, key, "1", constant.EndedInterviewSessionTTL).Result()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if err := publisher.PublishInterviewEvaluation(ctx, payload); err != nil {
		_ = client.Del(ctx, key).Err()
		return err
	}
	return nil
}
