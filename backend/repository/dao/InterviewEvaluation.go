package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// 总体

type InterviewEvaluation struct {
	ID         uint64                 `gorm:"primaryKey;autoIncrement"`
	UserID     uint                   `gorm:"type:int;index:idx_user_record,priority:1"`
	RecordID   uint64                 `gorm:"type:bigint unsigned;Index:idx_user_record,priority:2"`
	Comment    string                 `gorm:"type:text"`
	Score      float64                `gorm:"type:decimal(5,2)"`
	Dimensions []*EvaluationDimension `gorm:"type:json;serializer:json"`
	Status     string                 `gorm:"type:varchar(20);not null;default:processing"`
	CreatedAt  time.Time              `gorm:"autoCreateTime:milli"`
	UpdatedAt  time.Time              `gorm:"autoUpdateTime:milli"`
}
type EvaluationDimension struct {
	DimensionName string  `gorm:"type:varchar(128)"`
	Evaluation    string  `gorm:"type:text"`
	Score         float64 `gorm:"type:decimal(5,2)"`
}

type InterviewEvaluationDAO struct {
	db *gorm.DB
}

func NewInterviewEvaluationDAO(db *gorm.DB) *InterviewEvaluationDAO {
	return &InterviewEvaluationDAO{db: db}
}

func (d *InterviewEvaluationDAO) Create(ctx context.Context, evaluation *InterviewEvaluation) error {
	return d.db.WithContext(ctx).Create(evaluation).Error
}

func (d *InterviewEvaluationDAO) GetByID(ctx context.Context, id uint64) (*InterviewEvaluation, error) {
	var evaluation InterviewEvaluation
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&evaluation).Error; err != nil {
		return nil, err
	}
	return &evaluation, nil
}

func (d *InterviewEvaluationDAO) GetByRecordID(ctx context.Context, recordID uint64) (*InterviewEvaluation, error) {
	var evaluation InterviewEvaluation
	if err := d.db.WithContext(ctx).Where("record_id = ?", recordID).First(&evaluation).Error; err != nil {
		return nil, err
	}
	return &evaluation, nil
}

func (d *InterviewEvaluationDAO) ListByUserID(ctx context.Context, userID uint) ([]InterviewEvaluation, error) {
	var evaluations []InterviewEvaluation
	if err := d.db.WithContext(ctx).Where("user_id = ?", userID).Find(&evaluations).Error; err != nil {
		return nil, err
	}
	return evaluations, nil
}

func (d *InterviewEvaluationDAO) Update(ctx context.Context, evaluation *InterviewEvaluation) error {
	return d.db.WithContext(ctx).Save(evaluation).Error
}

func (d *InterviewEvaluationDAO) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return d.db.WithContext(ctx).Model(&InterviewEvaluation{}).Where("id = ?", id).Update("status", status).Error
}
