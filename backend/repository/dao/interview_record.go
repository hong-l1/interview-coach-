package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type InterviewRecord struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement"`
	UserID       uint      `gorm:"type:int unsigned;index;not null"`
	Type         string    `gorm:"type:varchar(128);not null"`
	Difficulty   string    `gorm:"type:varchar(64);not null"`
	Domain       string    `gorm:"type:varchar(255)"`
	CompanyName  string    `gorm:"type:varchar(255)"`
	PositionName string    `gorm:"type:varchar(255)"`
	Status       string    `gorm:"type:varchar(20);not null;default:'pending'"`
	Duration     int64     `gorm:"type:bigint;not null;default:0"`
	CreatedAt    time.Time `gorm:"autoCreateTime:milli"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime:milli"`
}

type InterviewRecordDAO struct {
	db *gorm.DB
}

func NewInterviewRecordDAO(db *gorm.DB) *InterviewRecordDAO {
	return &InterviewRecordDAO{
		db: db,
	}
}

func (d *InterviewRecordDAO) Create(ctx context.Context, record *InterviewRecord) error {
	return d.db.WithContext(ctx).Create(record).Error
}
