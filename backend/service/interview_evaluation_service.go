package service

import (
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"context"
)

type InterviewEvaluationService struct {
	interviewEvaluationRepository *repository.InterviewEvaluationRepository
}

func NewInterviewEvaluationService(repo *repository.InterviewEvaluationRepository) *InterviewEvaluationService {
	return &InterviewEvaluationService{
		interviewEvaluationRepository: repo,
	}
}

func (s *InterviewEvaluationService) Create(ctx context.Context, evaluation *dao.InterviewEvaluation) error {
	return s.interviewEvaluationRepository.Create(ctx, evaluation)
}

func (s *InterviewEvaluationService) GetByID(ctx context.Context, id uint64) (*dao.InterviewEvaluation, error) {
	return s.interviewEvaluationRepository.GetByID(ctx, id)
}

func (s *InterviewEvaluationService) GetByRecordID(ctx context.Context, recordID uint64) (*dao.InterviewEvaluation, error) {
	return s.interviewEvaluationRepository.GetByRecordID(ctx, recordID)
}

func (s *InterviewEvaluationService) ListByUserID(ctx context.Context, userID uint) ([]dao.InterviewEvaluation, error) {
	return s.interviewEvaluationRepository.ListByUserID(ctx, userID)
}

func (s *InterviewEvaluationService) Update(ctx context.Context, evaluation *dao.InterviewEvaluation) error {
	return s.interviewEvaluationRepository.Update(ctx, evaluation)
}

func (s *InterviewEvaluationService) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return s.interviewEvaluationRepository.UpdateStatus(ctx, id, status)
}
