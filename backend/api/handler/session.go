package handler

import (
	"context"
	"fmt"
	"time"
)

const (
	interviewSessionTTL   = 1 * time.Hour
	interviewStatusActive = "active"
	interviewStatusEnded  = "ended"
)

type interviewSession struct {
	SessionID     string `json:"session_id"`
	UserID        int64  `json:"user_id"`
	RecordID      uint64 `json:"record_id"`
	ResumeID      uint64 `json:"resume_id"`
	InterviewType string `json:"interview_type"`
	Domain        string `json:"domain"`
	Company       string `json:"company"`
	Position      string `json:"position"`
	Status        string `json:"status"`
	Difficulty    string `json:"difficulty"`
	CurrentIndex  int64  `json:"current_index"`
	CreatedAt     int64  `json:"created_at"`
	UpdatedAt     int64  `json:"updated_at"`
	EndedAt       int64  `json:"ended_at,omitempty"`
}

func (h *Handler) saveInterviewSession(ctx context.Context, session *interviewSession, ttl time.Duration) error {
	key := interviewSessionKey(session.SessionID)
	values := map[string]any{
		"session_id":     session.SessionID,
		"user_id":        session.UserID,
		"record_id":      session.RecordID,
		"resume_id":      session.ResumeID,
		"interview_type": session.InterviewType,
		"domain":         session.Domain,
		"company":        session.Company,
		"position":       session.Position,
		"status":         session.Status,
		"difficulty":     session.Difficulty,
		"current_index":  session.CurrentIndex,
		"created_at":     session.CreatedAt,
		"updated_at":     session.UpdatedAt,
		"ended_at":       session.EndedAt,
	}
	if err := h.redisClient.HSet(ctx, key, values).Err(); err != nil {
		return err
	}
	return h.redisClient.Expire(ctx, key, ttl).Err()
}
func interviewSessionKey(sessionID string) string {
	return fmt.Sprintf("mianshi:session:%s", sessionID)
}
