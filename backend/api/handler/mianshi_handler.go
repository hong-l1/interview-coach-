package handler

import (
	"awesomeProject4/backend/api/constant"
	"awesomeProject4/backend/api/handler/interview_run"
	"awesomeProject4/backend/api/utils"
	"awesomeProject4/backend/pkg/zapx"
	"net/http"

	"awesomeProject4/backend/api/validate"
	"awesomeProject4/backend/event"
	"context"
	_ "embed"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed lua/subanswer.lua
var subAnswerScript string

//go:embed lua/endmianshi.lua
var endInterviewScript string

func (h *Handler) startInterview(c *gin.Context) {
	req := validate.InterviewQuestionValidate(c)
	if req == nil {
		return
	}
	userID, ok := validate.ParseUserIDHeader(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Minute)
	entry, err := h.prepareInterviewEntry(ctx, req, userID)
	if err != nil {
		cancel()
		h.l.Error("prepare interview entry failed", zapx.Error(err))
		utils.InternalServerError(c, err.Error())
		return
	}
	sessionID := entry.session.SessionID
	h.cancelRegistry.Set(sessionID, cancel)
	utils.SetupSSEResponse(c)
	utils.SendSSEvent(c, "interview_begin", map[string]any{
		"session_id": sessionID,
		"message":    entry.beginMessage,
	})
	c.Status(http.StatusOK)
	runntime, err := interview_run.CreateInterviewRuntime(ctx, req.InterviewType, h.resumeService, h.redisClient)
	if err != nil {
		h.cancelRegistry.Delete(sessionID)
	}
	interview_run.Store.SetInterviewRuntime(sessionID, runntime)
	runntime.RunInterviewLoop(c, ctx, sessionID, 0, h.InterviewDialogueService, h.evaluationPublisher, uint(userID), entry.session.RecordID, req.InterviewType, req.Domain, req.ResumeID, req.Company, req.Position, false)
}

func (h *Handler) EndInterview(c *gin.Context) {
	sessionID, ok := validate.ParseSessionID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	now := time.Now().Unix()
	session, err := h.redisClient.HGetAll(ctx, interviewSessionKey(sessionID)).Result()
	if err != nil {
		h.l.Error("get interview session before end failed", zapx.Error(err), zapx.String("session_id", sessionID))
		utils.InternalServerError(c, err.Error())
		return
	}
	c.Status(http.StatusOK)
	answerListKey := interview_run.InterviewMsgListKey(sessionID)
	result, err := h.redisClient.Eval(ctx, endInterviewScript,
		[]string{
			interviewSessionKey(sessionID),
			answerListKey,
		},
		interviewStatusActive, interviewStatusEnded, now, int64(constant.EndedInterviewSessionTTL/time.Second), "__END__").Int()
	if err != nil {
		h.l.Error("end interview script failed", zapx.Error(err), zapx.String("session_id", sessionID))
		utils.InternalServerError(c, err.Error())
		return
	}
	switch result {
	case 0:
	case 1:
		utils.NotFound(c, "session not found")
		return
	case 2:
		utils.BadRequest(c, "interview session status is invalid")
		return
	}
	runtime := interview_run.Store.GetInterviewRuntime(sessionID)
	h.cancelRegistry.Cancel(sessionID)
	if runtime == nil {
		interview_run.Store.DeleteInterviewRuntime(sessionID)
		if err := h.publishEvaluationRequested(ctx, sessionID, session); err != nil {
			h.l.Error("publish interview evaluation event failed", zapx.Error(err), zapx.String("session_id", sessionID))
			utils.InternalServerError(c, err.Error())
			return
		}
	}
	utils.SuccessWithMessage(c, "interview ended", gin.H{
		"session_id": sessionID,
		"status":     interviewStatusEnded,
		"ended_at":   now,
	})
}

func (h *Handler) SubmitAnswer(c *gin.Context) {
	req := validate.SubmitAnswerValidate(c)
	if req == nil {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	now := time.Now().Unix()
	result, err := h.redisClient.Eval(
		ctx,
		subAnswerScript,
		[]string{
			interviewSessionKey(req.SessionID),
		},
		interviewStatusActive,
		now,
		int64(interviewSessionTTL/time.Second),
	).Int()
	if err != nil {
		h.l.Error("submit interview answer script failed", zapx.Error(err), zapx.String("session_id", req.SessionID))
		utils.InternalServerError(c, err.Error())
		return
	}
	switch result {
	case 0:
		if err := interview_run.Store.SendAnswer(ctx, req.SessionID, req.Answer); err != nil {
			h.l.Error("notify interview runtime failed", zapx.Error(err), zapx.String("session_id", req.SessionID))
			utils.InternalServerError(c, err.Error())
			return
		}
		utils.SuccessWithMessage(c, "answer submitted", map[string]any{
			"session_id": req.SessionID,
			"status":     interviewStatusActive,
			"updated_at": now,
		})
	case 1:
		utils.NotFound(c, "session not found")
	case 2:
		utils.BadRequest(c, "interview session status is invalid")
	default:
		utils.InternalServerError(c, "unknown submit answer result")
	}
}

func (h *Handler) ReconnectInterview(c *gin.Context) {
	sessionID, ok := validate.ParseSessionID(c)
	if !ok {
		return
	}
	queryCtx, queryCancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer queryCancel()
	result, err := h.redisClient.HGetAll(queryCtx, interviewSessionKey(sessionID)).Result()
	if err != nil {
		h.l.Error("reconnect interview failed", zapx.Error(err), zapx.String("session_id", sessionID))
		utils.InternalServerError(c, err.Error())
		return
	}
	if len(result) == 0 || result["status"] == interviewStatusEnded {
		utils.NotFound(c, "session not found or timeout")
		return
	}

	userID, err := strconv.ParseUint(result["user_id"], 10, 64)
	if err != nil {
		h.l.Error("parse reconnect user id failed", zapx.Error(err), zapx.String("session_id", sessionID))
		utils.InternalServerError(c, err.Error())
		return
	}
	reportID, err := strconv.ParseUint(result["record_id"], 10, 64)
	if err != nil {
		h.l.Error("parse reconnect record id failed", zapx.Error(err), zapx.String("session_id", sessionID))
		utils.InternalServerError(c, err.Error())
		return
	}
	index, err := strconv.ParseUint(result["current_index"], 10, 64)
	if err != nil {
		h.l.Error("parse current index  failed", zapx.Error(err), zapx.String("session_id", sessionID))
		utils.InternalServerError(c, err.Error())
		return
	}
	runtime := interview_run.Store.GetInterviewRuntime(sessionID)
	if runtime == nil {
		utils.InternalServerError(c, "interview runtime not found")
		return
	}
	loopCtx, loopCancel := context.WithTimeout(c.Request.Context(), 60*time.Minute)
	h.cancelRegistry.Set(sessionID, loopCancel)
	runtime, err = interview_run.CreateInterviewRuntime(loopCtx, result["interview_type"], h.resumeService, h.redisClient)
	if err != nil {
		h.cancelRegistry.Delete(sessionID)
		loopCancel()
		return
	}
	utils.SetupSSEResponse(c)
	utils.SendSSEvent(c, "interview_reconnected", map[string]any{
		"session_id": sessionID,
		"status":     result["status"],
	})
	interview_run.Store.SetInterviewRuntime(sessionID, runtime)
	resumeID, _ := strconv.ParseUint(result["resume_id"], 10, 64)
	runtime.RunInterviewLoop(c, loopCtx, sessionID, int(index+1), h.InterviewDialogueService, h.evaluationPublisher, uint(userID), reportID, result["interview_type"], result["domain"], resumeID, result["company"], result["position"], true)
}

func (h *Handler) GetSession(c *gin.Context) {
	sessionID, ok := validate.ParseSessionID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	result, err := h.redisClient.HGetAll(ctx, interviewSessionKey(sessionID)).Result()
	if err != nil {
		h.l.Error("get session failed", zapx.Error(err), zapx.String("session_id", sessionID))
		utils.InternalServerError(c, err.Error())
		return
	}
	if len(result) == 0 {
		utils.NotFound(c, "session not found")
		return
	}

	utils.SuccessWithMessage(c, "session found", result)
}

func (h *Handler) publishEvaluationRequested(ctx context.Context, sessionID string, session map[string]string) error {
	userID, err := strconv.ParseUint(session["user_id"], 10, 64)
	if err != nil {
		return err
	}
	recordID, err := strconv.ParseUint(session["record_id"], 10, 64)
	if err != nil {
		return err
	}
	return interview_run.PublishEvaluationOnce(ctx, h.redisClient, h.evaluationPublisher, event.InterviewEvaluationRequested{
		SessionID: sessionID,
		UserID:    uint(userID),
		RecordID:  recordID,
		ReportID:  recordID,
	})
}

type item struct {
	id  int
	cnt int
}
