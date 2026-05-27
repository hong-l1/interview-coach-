package service

import (
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"context"
)

type InterviewRecordService struct {
	interviewRecordRepository *repository.InterviewRecordRepository
}

func NewInterviewRecordService(repo *repository.InterviewRecordRepository) *InterviewRecordService {
	return &InterviewRecordService{
		interviewRecordRepository: repo,
	}
}

func (s *InterviewRecordService) Create(ctx context.Context, record *dao.InterviewRecord) error {
	return s.interviewRecordRepository.Create(ctx, record)
}
