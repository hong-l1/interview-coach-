package repository

import (
	"awesomeProject4/backend/repository/dao"
	"context"
)

type InterviewRecordRepository struct {
	interviewRecordDAO *dao.InterviewRecordDAO
}

func NewInterviewRecordRepository(interviewRecordDAO *dao.InterviewRecordDAO) *InterviewRecordRepository {
	return &InterviewRecordRepository{
		interviewRecordDAO: interviewRecordDAO,
	}
}

func (r *InterviewRecordRepository) Create(ctx context.Context, record *dao.InterviewRecord) error {
	return r.interviewRecordDAO.Create(ctx, record)
}
