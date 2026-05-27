package repository

import (
	"awesomeProject4/backend/repository/dao"
	"context"
)

type InterviewDialogueRepository struct {
	interviewDialogueDAO *dao.InterviewDialogueDAO
}

func NewInterviewDialogueRepository(dialogueDAO *dao.InterviewDialogueDAO) *InterviewDialogueRepository {
	return &InterviewDialogueRepository{
		interviewDialogueDAO: dialogueDAO,
	}
}

func (r *InterviewDialogueRepository) Create(ctx context.Context, dialogue *dao.InterviewDialogue) error {
	return r.interviewDialogueDAO.Create(ctx, dialogue)
}

func (r *InterviewDialogueRepository) BatchCreate(ctx context.Context, dialogues []*dao.InterviewDialogue) error {
	return r.interviewDialogueDAO.BatchCreate(ctx, dialogues)
}

func (r *InterviewDialogueRepository) GetByID(ctx context.Context, id uint64) (*dao.InterviewDialogue, error) {
	return r.interviewDialogueDAO.GetByID(ctx, id)
}

func (r *InterviewDialogueRepository) ListByReportID(ctx context.Context, reportID uint64) ([]dao.InterviewDialogue, error) {
	return r.interviewDialogueDAO.ListByReportID(ctx, reportID)
}

func (r *InterviewDialogueRepository) ListByUserID(ctx context.Context, userID uint) ([]dao.InterviewDialogue, error) {
	return r.interviewDialogueDAO.ListByUserID(ctx, userID)
}
