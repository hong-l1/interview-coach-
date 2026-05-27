package repository

import (
	"awesomeProject4/backend/repository/dao"
	"context"
)

type InterviewEvaluationRepository struct {
	interviewEvaluationDAO *dao.InterviewEvaluationDAO
}

func NewInterviewEvaluationRepository(evaluationDAO *dao.InterviewEvaluationDAO) *InterviewEvaluationRepository {
	return &InterviewEvaluationRepository{
		interviewEvaluationDAO: evaluationDAO,
	}
}

func (r *InterviewEvaluationRepository) Create(ctx context.Context, evaluation *dao.InterviewEvaluation) error {
	return r.interviewEvaluationDAO.Create(ctx, evaluation)
}

func (r *InterviewEvaluationRepository) GetByID(ctx context.Context, id uint64) (*dao.InterviewEvaluation, error) {
	return r.interviewEvaluationDAO.GetByID(ctx, id)
}

func (r *InterviewEvaluationRepository) GetByRecordID(ctx context.Context, recordID uint64) (*dao.InterviewEvaluation, error) {
	return r.interviewEvaluationDAO.GetByRecordID(ctx, recordID)
}

func (r *InterviewEvaluationRepository) ListByUserID(ctx context.Context, userID uint) ([]dao.InterviewEvaluation, error) {
	return r.interviewEvaluationDAO.ListByUserID(ctx, userID)
}

func (r *InterviewEvaluationRepository) Update(ctx context.Context, evaluation *dao.InterviewEvaluation) error {
	return r.interviewEvaluationDAO.Update(ctx, evaluation)
}

func (r *InterviewEvaluationRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.interviewEvaluationDAO.UpdateStatus(ctx, id, status)
}
