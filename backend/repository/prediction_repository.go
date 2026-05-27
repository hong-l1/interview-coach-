package repository

import (
	"awesomeProject4/backend/repository/dao"
	"context"
)

type PredictionRepository struct {
	predictionDAO *dao.PredictionDAO
}

func NewPredictionRepository(predictionDAO *dao.PredictionDAO) *PredictionRepository {
	return &PredictionRepository{
		predictionDAO: predictionDAO,
	}
}

func (r *PredictionRepository) Create(ctx context.Context, prediction *dao.Prediction) error {
	return r.predictionDAO.Create(ctx, prediction)
}

func (r *PredictionRepository) GetByID(ctx context.Context, id uint64) (*dao.Prediction, error) {
	return r.predictionDAO.GetByID(ctx, id)
}

func (r *PredictionRepository) GetByRecordID(ctx context.Context, recordID uint64) (*dao.Prediction, error) {
	return r.predictionDAO.GetByRecordID(ctx, recordID)
}

func (r *PredictionRepository) ListByUserID(ctx context.Context, userID uint) ([]dao.Prediction, error) {
	return r.predictionDAO.ListByUserID(ctx, userID)
}

func (r *PredictionRepository) CreateQuestion(ctx context.Context, question *dao.PredictionQuestion) error {
	return r.predictionDAO.CreateQuestion(ctx, question)
}

func (r *PredictionRepository) BatchCreateQuestions(ctx context.Context, questions []*dao.PredictionQuestion) error {
	return r.predictionDAO.BatchCreateQuestions(ctx, questions)
}

func (r *PredictionRepository) ListQuestions(ctx context.Context, predictionID uint64) ([]dao.PredictionQuestion, error) {
	return r.predictionDAO.ListQuestions(ctx, predictionID)
}
