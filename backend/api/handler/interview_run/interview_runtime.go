package interview_run

import (
	"awesomeProject4/agent/agents"
	"awesomeProject4/backend/api/constant"
	"awesomeProject4/backend/api/utils"
	"awesomeProject4/backend/event"
	"awesomeProject4/backend/repository/dao"
	"awesomeProject4/backend/service"
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/adk"
)

var runtimeStore sync.Once
var Store interviewRuntimeStore

func init() {
	runtimeStore.Do(func() {
		Store = interviewRuntimeStore{
			items: make(map[string]*InterviewRuntime),
		}
	})
}

type InterviewRuntime struct {
	runner   *adk.Runner
	client   redis.Cmdable
	answerCh chan string
}

type EvaluationEventPublisher interface {
	PublishInterviewEvaluation(ctx context.Context, event event.InterviewEvaluationRequested) error
}
type interviewRuntimeStore struct {
	sync.RWMutex
	items map[string]*InterviewRuntime
}

func getAgent(ctx context.Context, interviewType string, resumeService *service.ResumeService) (adk.Agent, error) {
	switch interviewType {
	case "comprehensive":
		return agents.NewComprehensiveAgent(ctx, resumeService)
	case "specialized":
		return agents.NewSpecializedAgent(ctx)
	default:
		return nil, fmt.Errorf("interview type %s not recognized", interviewType)
	}
}
func CreateInterviewRuntime(ctx context.Context, interviewType string, resumeService *service.ResumeService, client redis.Cmdable) (*InterviewRuntime, error) {
	agent, err := getAgent(ctx, interviewType, resumeService)
	if err != nil {
		return nil, err
	}
	return &InterviewRuntime{
		runner: adk.NewRunner(ctx, adk.RunnerConfig{
			Agent: agent,
		}),
		client:   client,
		answerCh: make(chan string, 1),
	}, nil
}
func (r *InterviewRuntime) RunInterviewLoop(ctx *gin.Context, loopCtx context.Context, sessionID string, cur int,
	dialogueDAO *service.InterviewDialogueService, eventPublisher EvaluationEventPublisher, userID uint, reportID uint64, interviewType string, domain string, resumeID uint64, company string, position string, loadHistoryFromRedis bool) {
	pending := make([]*dao.InterviewDialogue, 0, 20) //数据库
	history := make([]*schema.Message, 0, 10)        // agent
	if loadHistoryFromRedis {
		records, err := r.loadHistory(ctx, sessionID)
		if err != nil {
			utils.SendSSEvent(ctx, "error", map[string]interface{}{
				"session_id": sessionID,
				"message":    err.Error(),
			})
			return
		}
		history = historyMessagesToSchemaMessages(records)
	}
	contextPrompt := buildInterviewPrompt(interviewType, domain, resumeID, company, position)
	lastFlushAt := time.Now()
	flush := func(flushCtx context.Context, force bool) error {
		if !force && len(pending) < constant.DialogueBatchSize && time.Since(lastFlushAt) < constant.DialogueFlushInterval {
			return nil
		}
		if err := dialogueDAO.BatchCreate(flushCtx, pending); err != nil {
			return err
		}
		pending = pending[:0]
		lastFlushAt = time.Now()
		return nil
	}
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := flush(cleanupCtx, true); err != nil {
			return
		}
		if r.isSessionEnded(cleanupCtx, sessionID) && eventPublisher != nil {
			_ = PublishEvaluationOnce(cleanupCtx, r.client, eventPublisher, event.InterviewEvaluationRequested{
				SessionID: sessionID,
				UserID:    userID,
				RecordID:  reportID,
				ReportID:  reportID,
			})
		}
		Store.DeleteInterviewRuntime(sessionID)
	}()
	for questionIndex := cur; questionIndex <= constant.MaxInterviewQuestions; questionIndex++ {
		if len(history) > 10 {
			history = history[len(history)-10:]
		}
		runHistory := history
		if contextPrompt != "" {
			runHistory = append([]*schema.Message{schema.UserMessage(contextPrompt)}, history...)
		}
		itor := r.runner.Run(loopCtx, runHistory)
		event, ok := itor.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return
		}
		if event.Output.MessageOutput != nil && event.Output.MessageOutput.Message.Content != "" {
			last := event.Output.MessageOutput.Message.Content
			if err := r.SaveHistory(ctx, sessionID, historyMessage{Role: "assistant", Index: questionIndex, Type: "question", Content: last}); err != nil {
				utils.SendSSEvent(ctx, "error", map[string]interface{}{
					"session_id": sessionID,
					"message":    err.Error(),
				})
				return
			}
			utils.SendSSEvent(ctx, "question", map[string]interface{}{
				"session_id":    sessionID,
				"questionIndex": questionIndex,
				"question":      last,
			})
			history = append(history, schema.AssistantMessage(last, nil))
			select {
			case <-loopCtx.Done():
				return
			case curAnswer := <-r.answerCh:
				if err := r.SaveHistory(ctx, sessionID, historyMessage{Role: "user", Type: "answer", Index: questionIndex, Content: curAnswer}); err != nil {
					utils.SendSSEvent(ctx, "error", map[string]interface{}{
						"session_id": sessionID,
						"message":    err.Error(),
					})
					return
				}
				history = append(history, schema.UserMessage(curAnswer))
				pending = append(pending, &dao.InterviewDialogue{
					UserID:   userID,
					ReportID: reportID,
					Question: last,
					Answer:   curAnswer,
				})
				if err := flush(ctx, false); err != nil {
					utils.SendSSEvent(ctx, "error", map[string]interface{}{
						"session_id": sessionID,
						"message":    err.Error(),
					})
					return
				}
			}
		}
	}
	if err := flush(ctx, true); err != nil {
		utils.SendSSEvent(ctx, "error", map[string]interface{}{
			"session_id": sessionID,
			"message":    err.Error(),
		})
		return
	}
	now := time.Now().Unix()
	if err := r.markSessionEnded(ctx, sessionID, now); err != nil {
		utils.SendSSEvent(ctx, "error", map[string]interface{}{
			"session_id": sessionID,
			"message":    err.Error(),
		})
		return
	}
	utils.SendSSEvent(ctx, "interview_end", map[string]interface{}{
		"session_id": sessionID,
		"status":     "ended",
		"ended_at":   now,
	})
}

func (r *interviewRuntimeStore) SendAnswer(ctx context.Context, sessionID string, answer string) error {
	runtime := r.GetInterviewRuntime(sessionID)
	if runtime == nil {
		return errors.New("interview runtime not found")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case runtime.answerCh <- answer:
		return nil
	default:
		return errors.New("previous answer is still being processed")
	}
}

func (r *interviewRuntimeStore) GetInterviewRuntime(sessionID string) *InterviewRuntime {
	r.RLock()
	runtime := r.items[sessionID]
	r.RUnlock()
	return runtime
}

func (r *interviewRuntimeStore) SetInterviewRuntime(sessionID string, runtime *InterviewRuntime) {
	r.Lock()
	r.items[sessionID] = runtime
	r.Unlock()
}

func (r *interviewRuntimeStore) DeleteInterviewRuntime(sessionID string) {
	r.Lock()
	delete(r.items, sessionID)
	r.Unlock()
}

func buildInterviewPrompt(interviewType, domain string, resumeID uint64, company string, position string) string {
	roleContext := make([]string, 0, 2)
	if trimmedPosition := strings.TrimSpace(position); trimmedPosition != "" {
		roleContext = append(roleContext, fmt.Sprintf("target position: %s", trimmedPosition))
	}
	if trimmedCompany := strings.TrimSpace(company); trimmedCompany != "" {
		roleContext = append(roleContext, fmt.Sprintf("target company: %s", trimmedCompany))
	}
	roleHint := ""
	if len(roleContext) > 0 {
		roleHint = " Keep the interview questions aligned with the candidate's " + strings.Join(roleContext, " and ") + ". Prioritize responsibilities, project depth, and technical scenarios that are realistic for that target role."
	}
	switch interviewType {
	case "comprehensive":
		return fmt.Sprintf("Start a comprehensive interview. Use resume_id=%d as context.%s The user's first message is their self-introduction. After considering that introduction, ask exactly one formal interview question. On later turns, continue using the resume context, the target role context, and the candidate's latest answer to ask the next question.", resumeID, roleHint)
	case "specialized":
		return fmt.Sprintf("Start a specialized interview for %s.%s Ask exactly one formal interview question each turn, and when possible use scenarios that fit the target role.", domain, roleHint)
	default:
		return "Generate exactly one interview question." + roleHint
	}
}

func (r *InterviewRuntime) isSessionEnded(ctx context.Context, sessionID string) bool {
	status, err := r.client.HGet(ctx, interviewSessionKey(sessionID), "status").Result()
	return err == nil && status == "ended"
}

func (r *InterviewRuntime) markSessionEnded(ctx context.Context, sessionID string, endedAt int64) error {
	key := interviewSessionKey(sessionID)
	if err := r.client.HSet(ctx, key,
		"status", "ended",
		"ended_at", endedAt,
		"updated_at", endedAt,
	).Err(); err != nil {
		return err
	}
	if err := r.client.Expire(ctx, key, constant.EndedInterviewSessionTTL).Err(); err != nil {
		return err
	}
	return r.client.Expire(ctx, InterviewMsgListKey(sessionID), constant.EndedInterviewSessionTTL).Err()
}

func interviewSessionKey(sessionID string) string {
	return fmt.Sprintf("mianshi:session:%s", sessionID)
}
