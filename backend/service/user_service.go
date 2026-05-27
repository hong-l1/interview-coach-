package service

import (
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"context"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		userRepository: repo,
	}
}

func (s *UserService) Register(ctx context.Context, user *dao.User) error {
	return s.userRepository.Create(ctx, user)
}

func (s *UserService) Login(ctx context.Context, email string) (*dao.User, error) {
	return s.userRepository.GetByEmail(ctx, email)
}
