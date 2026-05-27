package service

import (
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"context"
)

type InterviewDialogueService struct {
	interviewDialogueRepository *repository.InterviewDialogueRepository
}

func NewInterviewDialogueService(repo *repository.InterviewDialogueRepository) *InterviewDialogueService {
	return &InterviewDialogueService{
		interviewDialogueRepository: repo,
	}
}

func (s *InterviewDialogueService) Create(ctx context.Context, dialogue *dao.InterviewDialogue) error {
	return s.interviewDialogueRepository.Create(ctx, dialogue)
}

func (s *InterviewDialogueService) BatchCreate(ctx context.Context, dialogues []*dao.InterviewDialogue) error {
	return s.interviewDialogueRepository.BatchCreate(ctx, dialogues)
}

func (s *InterviewDialogueService) GetByID(ctx context.Context, id uint64) (*dao.InterviewDialogue, error) {
	return s.interviewDialogueRepository.GetByID(ctx, id)
}

func (s *InterviewDialogueService) ListByReportID(ctx context.Context, reportID uint64) ([]dao.InterviewDialogue, error) {
	return s.interviewDialogueRepository.ListByReportID(ctx, reportID)
}

func (s *InterviewDialogueService) ListByUserID(ctx context.Context, userID uint) ([]dao.InterviewDialogue, error) {
	return s.interviewDialogueRepository.ListByUserID(ctx, userID)
}
