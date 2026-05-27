package repository

import (
	"awesomeProject4/backend/repository/dao"
	"context"
)

type EvaluationDetailRepository struct {
	evaluationDetailDAO *dao.EvaluationDetailDAO
}

func NewEvaluationDetailRepository(detailDAO *dao.EvaluationDetailDAO) *EvaluationDetailRepository {
	return &EvaluationDetailRepository{
		evaluationDetailDAO: detailDAO,
	}
}

func (r *EvaluationDetailRepository) Create(ctx context.Context, detail *dao.EvaluationDetail) error {
	return r.evaluationDetailDAO.Create(ctx, detail)
}

func (r *EvaluationDetailRepository) GetByID(ctx context.Context, id uint64) (*dao.EvaluationDetail, error) {
	return r.evaluationDetailDAO.GetByID(ctx, id)
}

func (r *EvaluationDetailRepository) GetByRecordID(ctx context.Context, recordID uint64) (*dao.EvaluationDetail, error) {
	return r.evaluationDetailDAO.GetByRecordID(ctx, recordID)
}

func (r *EvaluationDetailRepository) ListByUserID(ctx context.Context, userID uint) ([]dao.EvaluationDetail, error) {
	return r.evaluationDetailDAO.ListByUserID(ctx, userID)
}

func (r *EvaluationDetailRepository) Update(ctx context.Context, detail *dao.EvaluationDetail) error {
	return r.evaluationDetailDAO.Update(ctx, detail)
}

func (r *EvaluationDetailRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.evaluationDetailDAO.UpdateStatus(ctx, id, status)
}
