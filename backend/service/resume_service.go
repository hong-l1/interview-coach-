package service

import (
	schemas "awesomeProject4/agent/schema"
	agentservice "awesomeProject4/agent/service"
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

type ResumeService struct {
	resumeRepository *repository.ResumeRepository
}

func NewResumeService(repo *repository.ResumeRepository) *ResumeService {
	return &ResumeService{
		resumeRepository: repo,
	}
}

func (s *ResumeService) ParseResume(ctx context.Context, userID uint, filePath string) (*dao.Resume, *schemas.Resume, error) {
	data, err := agentservice.ParseResumeService(ctx, filePath)
	if err != nil {
		return nil, nil, err
	}

	content, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	var fileSize int64
	if info, statErr := os.Stat(filePath); statErr == nil {
		fileSize = info.Size()
	}
	model := &dao.Resume{
		UserID:    userID,
		Content:   string(content),
		FileName:  filepath.Base(filePath),
		FileSize:  fileSize,
		FileType:  filepath.Ext(filePath),
		IsDefault: 0,
	}
	if err := s.resumeRepository.Create(ctx, model); err != nil {
		return nil, nil, err
	}
	return model, data, nil
}

func (s *ResumeService) DeleteResume(ctx context.Context, id uint64, userID uint) error {
	return s.resumeRepository.Delete(ctx, id, userID)
}

func (s *ResumeService) SetDefaultResume(ctx context.Context, id uint64, userID uint) error {
	return s.resumeRepository.SetDefault(ctx, id, userID)
}

func (s *ResumeService) GetResume(ctx context.Context, id uint64) (*dao.Resume, error) {
	return s.resumeRepository.GetByID(ctx, id)
}

func (s *ResumeService) ListResumes(ctx context.Context, userID uint) ([]dao.Resume, error) {
	return s.resumeRepository.ListByUserID(ctx, userID)
}

func (s *ResumeService) GetDefaultResume(ctx context.Context, userID uint) (*dao.Resume, error) {
	return s.resumeRepository.GetDefaultByUserID(ctx, userID)
}
