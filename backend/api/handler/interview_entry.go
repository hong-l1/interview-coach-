package handler

import (
	"awesomeProject4/backend/api/validate"
	"awesomeProject4/backend/repository/dao"
	"context"
	"time"

	"github.com/google/uuid"
)

type interviewEntryDeps struct {
	createRecord func(ctx context.Context, record *dao.InterviewRecord) error
	saveSession  func(ctx context.Context, session *interviewSession, ttl time.Duration) error
	newSessionID func() string
	now          func() time.Time
}

type preparedInterviewEntry struct {
	session      *interviewSession
	beginMessage string
}

func buildInterviewEntry(
	ctx context.Context,
	req *validate.InterviewQuestionRequest,
	userID int64,
	deps interviewEntryDeps,
) (*preparedInterviewEntry, error) {
	now := deps.now()
	sessionID := deps.newSessionID()
	record := &dao.InterviewRecord{
		UserID:       uint(userID),
		Type:         req.InterviewType,
		Difficulty:   req.Difficulty,
		Domain:       req.Domain,
		CompanyName:  req.Company,
		PositionName: req.Position,
		Status:       interviewStatusActive,
		Duration:     0,
	}
	if err := deps.createRecord(ctx, record); err != nil {
		return nil, err
	}

	beginMessage := "Interview started successfully."
	if req.InterviewType != "specialized" {
		beginMessage = "Interview started successfully. Please begin with a brief self-introduction."
	}

	session := &interviewSession{
		SessionID:     sessionID,
		UserID:        userID,
		RecordID:      record.ID,
		ResumeID:      req.ResumeID,
		InterviewType: req.InterviewType,
		Domain:        req.Domain,
		Company:       req.Company,
		Position:      req.Position,
		Status:        interviewStatusActive,
		Difficulty:    req.Difficulty,
		CurrentIndex:  0,
		CreatedAt:     now.Unix(),
		UpdatedAt:     now.Unix(),
	}
	if err := deps.saveSession(ctx, session, interviewSessionTTL); err != nil {
		return nil, err
	}

	return &preparedInterviewEntry{
		session:      session,
		beginMessage: beginMessage,
	}, nil
}

func (h *Handler) prepareInterviewEntry(
	ctx context.Context,
	req *validate.InterviewQuestionRequest,
	userID int64,
) (*preparedInterviewEntry, error) {
	return buildInterviewEntry(ctx, req, userID, interviewEntryDeps{
		createRecord: func(ctx context.Context, record *dao.InterviewRecord) error {
			return h.interviewService.Create(ctx, record)
		},
		saveSession: func(ctx context.Context, session *interviewSession, ttl time.Duration) error {
			return h.saveInterviewSession(ctx, session, ttl)
		},
		newSessionID: uuid.NewString,
		now:          time.Now,
	})
}
