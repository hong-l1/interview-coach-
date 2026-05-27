package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Prediction struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement"`
	UserID   uint   `gorm:"type:int unsigned;index;not null"`
	RecordID uint64 `gorm:"type:bigint unsigned;index;not null"`
	ReportID uint64 `gorm:"type:bigint unsigned;index;not null"`
}
type PredictionQuestion struct {
	ID              uint64    `gorm:"primaryKey;autoIncrement"`
	PredictionID    uint64    `gorm:"index;not null"`
	Question        string    `gorm:"type:text;not null"`
	Focus           string    `gorm:"type:text"`
	ReferenceAnswer string    `gorm:"type:text"`
	FollowUp        string    `gorm:"type:text"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
}

type PredictionDAO struct {
	db *gorm.DB
}

func NewPredictionDAO(db *gorm.DB) *PredictionDAO {
	return &PredictionDAO{db: db}
}

func (d *PredictionDAO) Create(ctx context.Context, prediction *Prediction) error {
	return d.db.WithContext(ctx).Create(prediction).Error
}

func (d *PredictionDAO) GetByID(ctx context.Context, id uint64) (*Prediction, error) {
	var prediction Prediction
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&prediction).Error; err != nil {
		return nil, err
	}
	return &prediction, nil
}

func (d *PredictionDAO) GetByRecordID(ctx context.Context, recordID uint64) (*Prediction, error) {
	var prediction Prediction
	if err := d.db.WithContext(ctx).Where("record_id = ?", recordID).First(&prediction).Error; err != nil {
		return nil, err
	}
	return &prediction, nil
}

func (d *PredictionDAO) ListByUserID(ctx context.Context, userID uint) ([]Prediction, error) {
	var predictions []Prediction
	if err := d.db.WithContext(ctx).Where("user_id = ?", userID).Find(&predictions).Error; err != nil {
		return nil, err
	}
	return predictions, nil
}

func (d *PredictionDAO) CreateQuestion(ctx context.Context, question *PredictionQuestion) error {
	return d.db.WithContext(ctx).Create(question).Error
}

func (d *PredictionDAO) BatchCreateQuestions(ctx context.Context, questions []*PredictionQuestion) error {
	if len(questions) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Create(&questions).Error
}

func (d *PredictionDAO) ListQuestions(ctx context.Context, predictionID uint64) ([]PredictionQuestion, error) {
	var questions []PredictionQuestion
	if err := d.db.WithContext(ctx).Where("prediction_id = ?", predictionID).Find(&questions).Error; err != nil {
		return nil, err
	}
	return questions, nil
}
