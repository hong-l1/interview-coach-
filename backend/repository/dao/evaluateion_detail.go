package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type EvaluationDetail struct {
	ID          uint64            `gorm:"primaryKey;autoIncrement"`
	UserID      uint              `gorm:"type:int;index:idx_user_record,priority:1"`
	RecordID    uint64            `gorm:"type:bigint unsigned;index:idx_user_record,priority:2"`
	Evaluations []*EvaluationItem `gorm:"type:json;serializer:json"`
	Status      string            `gorm:"type:varchar(20);not null;default:'pending'"`
	CreatedAt   time.Time         `gorm:"autoCreateTime:milli"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime:milli"`
}
type EvaluationItem struct {
	Comment []*Comment  `gorm:"type:json;serializer:json"`
	Message []*Dialogue `gorm:"type:json;serializer:json"`
}
type Dialogue struct {
	Question string `gorm:"type:text"`
	Answer   string `gorm:"type:text"`
}
type Comment struct {
	Score      int32  `gorm:"type:int"`
	KnowPoints string `gorm:"type:text"`
	Strengths  string `gorm:"type:text"`
	Weaknesses string `gorm:"type:text"`
	Suggestion string `gorm:"type:text"`
	Reference  string `gorm:"type:text"`
}

type EvaluationDetailDAO struct {
	db *gorm.DB
}

func NewEvaluationDetailDAO(db *gorm.DB) *EvaluationDetailDAO {
	return &EvaluationDetailDAO{db: db}
}

func (d *EvaluationDetailDAO) Create(ctx context.Context, detail *EvaluationDetail) error {
	return d.db.WithContext(ctx).Create(detail).Error
}

func (d *EvaluationDetailDAO) GetByID(ctx context.Context, id uint64) (*EvaluationDetail, error) {
	var detail EvaluationDetail
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&detail).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}

func (d *EvaluationDetailDAO) GetByRecordID(ctx context.Context, recordID uint64) (*EvaluationDetail, error) {
	var detail EvaluationDetail
	if err := d.db.WithContext(ctx).Where("record_id = ?", recordID).First(&detail).Error; err != nil {
		return nil, err
	}
	return &detail, nil
}

func (d *EvaluationDetailDAO) ListByUserID(ctx context.Context, userID uint) ([]EvaluationDetail, error) {
	var details []EvaluationDetail
	if err := d.db.WithContext(ctx).Where("user_id = ?", userID).Find(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}

func (d *EvaluationDetailDAO) Update(ctx context.Context, detail *EvaluationDetail) error {
	return d.db.WithContext(ctx).Save(detail).Error
}

func (d *EvaluationDetailDAO) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return d.db.WithContext(ctx).Model(&EvaluationDetail{}).Where("id = ?", id).Update("status", status).Error
}
