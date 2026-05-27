package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Resume struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"`
	UserID    uint           `gorm:"index;not null"`
	Content   string         `gorm:"type:longtext"`
	FileName  string         `gorm:"type:varchar(255)"`
	FileSize  int64          `gorm:"type:int"`
	FileType  string         `gorm:"type:varchar(255)"`
	IsDefault int            `gorm:"type:int;default:0"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedAt time.Time      `gorm:"autoCreateTime:milli"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime:milli"`
}

type ResumeDAO struct {
	db *gorm.DB
}

func NewResumeDAO(db *gorm.DB) *ResumeDAO {
	return &ResumeDAO{
		db: db,
	}
}

func (d *ResumeDAO) Create(ctx context.Context, resume *Resume) error {
	return d.db.WithContext(ctx).Create(resume).Error
}

func (d *ResumeDAO) GetByID(ctx context.Context, id uint64) (*Resume, error) {
	var resume Resume
	if err := d.db.WithContext(ctx).
		Where("id = ?", id).
		First(&resume).Error; err != nil {
		return nil, err
	}
	return &resume, nil
}

func (d *ResumeDAO) GetByIDAndUserID(ctx context.Context, id uint64, userID uint) (*Resume, error) {
	var resume Resume
	if err := d.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&resume).Error; err != nil {
		return nil, err
	}
	return &resume, nil
}

func (d *ResumeDAO) ListByUserID(ctx context.Context, userID uint) ([]Resume, error) {
	var resumes []Resume
	if err := d.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("is_default DESC, created_at DESC").
		Find(&resumes).Error; err != nil {
		return nil, err
	}
	return resumes, nil
}

func (d *ResumeDAO) GetDefaultByUserID(ctx context.Context, userID uint) (*Resume, error) {
	var resume Resume
	if err := d.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ?", userID, 1).
		Order("updated_at DESC").
		First(&resume).Error; err != nil {
		return nil, err
	}
	return &resume, nil
}

func (d *ResumeDAO) Delete(ctx context.Context, id uint64, userID uint) error {
	tx := d.db.WithContext(ctx).
		Model(&Resume{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_default", 0)
	if tx.Error != nil {
		return tx.Error
	}
	return d.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&Resume{}).Error
}

func (d *ResumeDAO) ClearDefault(ctx context.Context, userID uint) error {
	return d.db.WithContext(ctx).
		Model(&Resume{}).
		Where("user_id = ?", userID).
		Update("is_default", 0).Error
}

func (d *ResumeDAO) SetDefault(ctx context.Context, id uint64, userID uint) error {
	return d.db.WithContext(ctx).
		Model(&Resume{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_default", 1).Error
}
