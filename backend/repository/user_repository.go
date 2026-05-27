package repository

import (
	"awesomeProject4/backend/repository/dao"
	"context"
)

type UserRepository struct {
	userDAO *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		userDAO: dao,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *dao.User) error {
	return r.userDAO.Create(ctx, user)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*dao.User, error) {
	return r.userDAO.GetByEmail(ctx, email)
}
