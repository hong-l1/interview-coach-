package service

import (
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"context"
)

type PredictionService struct {
	predictionRepository *repository.PredictionRepository
}

func NewPredictionService(repo *repository.PredictionRepository) *PredictionService {
	return &PredictionService{
		predictionRepository: repo,
	}
}

func (s *PredictionService) Create(ctx context.Context, prediction *dao.Prediction) error {
	return s.predictionRepository.Create(ctx, prediction)
}

func (s *PredictionService) GetByID(ctx context.Context, id uint64) (*dao.Prediction, error) {
	return s.predictionRepository.GetByID(ctx, id)
}

func (s *PredictionService) GetByRecordID(ctx context.Context, recordID uint64) (*dao.Prediction, error) {
	return s.predictionRepository.GetByRecordID(ctx, recordID)
}

func (s *PredictionService) ListByUserID(ctx context.Context, userID uint) ([]dao.Prediction, error) {
	return s.predictionRepository.ListByUserID(ctx, userID)
}

func (s *PredictionService) CreateQuestion(ctx context.Context, question *dao.PredictionQuestion) error {
	return s.predictionRepository.CreateQuestion(ctx, question)
}

func (s *PredictionService) BatchCreateQuestions(ctx context.Context, questions []*dao.PredictionQuestion) error {
	return s.predictionRepository.BatchCreateQuestions(ctx, questions)
}

func (s *PredictionService) ListQuestions(ctx context.Context, predictionID uint64) ([]dao.PredictionQuestion, error) {
	return s.predictionRepository.ListQuestions(ctx, predictionID)
}
