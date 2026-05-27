package repository

import (
	"awesomeProject4/backend/repository/dao"
	"context"
)

type UserModelRepository struct {
	userModelDAO *dao.UserModelDAO
}

func NewUserModelRepository(userModelDAO *dao.UserModelDAO) *UserModelRepository {
	return &UserModelRepository{
		userModelDAO: userModelDAO,
	}
}

func (r *UserModelRepository) Create(ctx context.Context, model *dao.UserModel) error {
	return r.userModelDAO.Create(ctx, model)
}

func (r *UserModelRepository) Update(ctx context.Context, model *dao.UserModel) error {
	return r.userModelDAO.Update(ctx, model)
}

func (r *UserModelRepository) Delete(ctx context.Context, id uint64) error {
	return r.userModelDAO.Delete(ctx, id)
}

func (r *UserModelRepository) Get(ctx context.Context, id uint64) (*dao.UserModel, error) {
	return r.userModelDAO.Get(ctx, id)
}

func (r *UserModelRepository) List(ctx context.Context) ([]*dao.UserModel, error) {
	return r.userModelDAO.List(ctx)
}
