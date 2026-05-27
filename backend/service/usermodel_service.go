package service

import (
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"context"
)

type UserModelService struct {
	userModelRepository *repository.UserModelRepository
}

func NewUserModelService(repo *repository.UserModelRepository) *UserModelService {
	return &UserModelService{
		userModelRepository: repo,
	}
}

func (s *UserModelService) Create(ctx context.Context, model *dao.UserModel) error {
	return s.userModelRepository.Create(ctx, model)
}

func (s *UserModelService) Update(ctx context.Context, model *dao.UserModel) error {
	return s.userModelRepository.Update(ctx, model)
}

func (s *UserModelService) Delete(ctx context.Context, id uint64) error {
	return s.userModelRepository.Delete(ctx, id)
}

func (s *UserModelService) Get(ctx context.Context, id uint64) (*dao.UserModel, error) {
	return s.userModelRepository.Get(ctx, id)
}

func (s *UserModelService) List(ctx context.Context) ([]*dao.UserModel, error) {
	return s.userModelRepository.List(ctx)
}
