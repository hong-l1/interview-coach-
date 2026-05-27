package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type InterviewDialogue struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	UserID    uint      `gorm:"type:int unsigned;not null;index:idx_user_report,priority:1"`
	ReportID  uint64    `gorm:"type:bigint unsigned;not null;index:idx_user_report,priority:2"`
	Question  string    `gorm:"type:text"`
	Answer    string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime:milli"`
}

type InterviewDialogueDAO struct {
	db *gorm.DB
}

func NewInterviewDialogueDAO(db *gorm.DB) *InterviewDialogueDAO {
	return &InterviewDialogueDAO{db: db}
}

func (d *InterviewDialogueDAO) BatchCreate(ctx context.Context, dialogues []*InterviewDialogue) error {
	if len(dialogues) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Create(&dialogues).Error
}

func (d *InterviewDialogueDAO) Create(ctx context.Context, dialogue *InterviewDialogue) error {
	return d.db.WithContext(ctx).Create(dialogue).Error
}

func (d *InterviewDialogueDAO) GetByID(ctx context.Context, id uint64) (*InterviewDialogue, error) {
	var dialogue InterviewDialogue
	if err := d.db.WithContext(ctx).Where("id = ?", id).First(&dialogue).Error; err != nil {
		return nil, err
	}
	return &dialogue, nil
}

func (d *InterviewDialogueDAO) ListByReportID(ctx context.Context, reportID uint64) ([]InterviewDialogue, error) {
	var dialogues []InterviewDialogue
	if err := d.db.WithContext(ctx).Where("report_id = ?", reportID).Order("created_at ASC").Find(&dialogues).Error; err != nil {
		return nil, err
	}
	return dialogues, nil
}

func (d *InterviewDialogueDAO) ListByUserID(ctx context.Context, userID uint) ([]InterviewDialogue, error) {
	var dialogues []InterviewDialogue
	if err := d.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&dialogues).Error; err != nil {
		return nil, err
	}
	return dialogues, nil
}
