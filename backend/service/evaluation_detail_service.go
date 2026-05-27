package service

import (
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"context"
)

type EvaluationDetailService struct {
	evaluationDetailRepository *repository.EvaluationDetailRepository
}

func NewEvaluationDetailService(repo *repository.EvaluationDetailRepository) *EvaluationDetailService {
	return &EvaluationDetailService{
		evaluationDetailRepository: repo,
	}
}

func (s *EvaluationDetailService) Create(ctx context.Context, detail *dao.EvaluationDetail) error {
	return s.evaluationDetailRepository.Create(ctx, detail)
}

func (s *EvaluationDetailService) GetByID(ctx context.Context, id uint64) (*dao.EvaluationDetail, error) {
	return s.evaluationDetailRepository.GetByID(ctx, id)
}

func (s *EvaluationDetailService) GetByRecordID(ctx context.Context, recordID uint64) (*dao.EvaluationDetail, error) {
	return s.evaluationDetailRepository.GetByRecordID(ctx, recordID)
}

func (s *EvaluationDetailService) ListByUserID(ctx context.Context, userID uint) ([]dao.EvaluationDetail, error) {
	return s.evaluationDetailRepository.ListByUserID(ctx, userID)
}

func (s *EvaluationDetailService) Update(ctx context.Context, detail *dao.EvaluationDetail) error {
	return s.evaluationDetailRepository.Update(ctx, detail)
}

func (s *EvaluationDetailService) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return s.evaluationDetailRepository.UpdateStatus(ctx, id, status)
}
