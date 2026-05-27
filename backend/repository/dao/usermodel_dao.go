package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserModel struct {
	ID              uint64         `gorm:"primaryKey;autoIncrement"`
	UserID          int64          `gorm:"type:bigint;index;not null"`
	Name            string         `gorm:"type:varchar(255);not null"`
	ModelName       string         `gorm:"type:varchar(255);not null"`
	Protocol        string         `gorm:"type:varchar(255);not null"`
	BaseURL         string         `gorm:"type:varchar(255);not null"`
	APIKeyEncrypted string         `gorm:"type:text;not null"`
	ProviderName    string         `gorm:"type:varchar(255)"`
	IsDefault       int            `gorm:"type:int;default:0"`
	CreatedAt       time.Time      `gorm:"autoCreateTime:milli"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime:milli"`
	Deleted         gorm.DeletedAt `gorm:"index"`
}

type UserModelDAO struct {
	db *gorm.DB
}

func NewUserModelDAO(db *gorm.DB) *UserModelDAO {
	return &UserModelDAO{
		db: db,
	}
}

func (d *UserModelDAO) Create(ctx context.Context, model *UserModel) error {
	return d.db.WithContext(ctx).Create(model).Error
}

func (d *UserModelDAO) Update(ctx context.Context, model *UserModel) error {
	return d.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", model.ID).Updates(model).Error
}

func (d *UserModelDAO) Delete(ctx context.Context, id uint64) error {
	return d.db.WithContext(ctx).Where("id = ?", id).Delete(&UserModel{}).Error
}

func (d *UserModelDAO) Get(ctx context.Context, id uint64) (*UserModel, error) {
	var model UserModel
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (d *UserModelDAO) List(ctx context.Context) ([]*UserModel, error) {
	var models []*UserModel
	if err := d.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	return models, nil
}
