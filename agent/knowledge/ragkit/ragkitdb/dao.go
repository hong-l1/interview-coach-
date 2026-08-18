package ragkitdb

import "time"

// StrategyProfileRow 是策略档案表（DB 持久化，修正源项目内存态缺陷）。
type StrategyProfileRow struct {
	ID                     uint64 `gorm:"primaryKey;autoIncrement"`
	Name                   string `gorm:"type:varchar(128);uniqueIndex;not null"`
	Status                 string `gorm:"type:varchar(32);index;not null"` // active/candidate/baseline
	FusionStrategy         string `gorm:"type:varchar(64)"`
	TopKConfig             string `gorm:"type:text"` // JSON
	EvidenceGateThresholds string `gorm:"type:text"` // JSON
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (StrategyProfileRow) TableName() string { return "ragkit_strategy_profile" }

// AuditEventRow 是审计事件表。
type AuditEventRow struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement"`
	AuditTraceID    string `gorm:"type:varchar(128);uniqueIndex;not null"`
	Operator        string `gorm:"type:varchar(128)"`
	Action          string `gorm:"type:varchar(64);index;not null"`
	ResourceType    string `gorm:"type:varchar(64)"`
	ResourceID      string `gorm:"type:varchar(128)"`
	Before          string `gorm:"type:text"`
	After           string `gorm:"type:text"`
	Result          string `gorm:"type:varchar(32)"`
	Reason          string `gorm:"type:text"`
	IP              string `gorm:"type:varchar(64)"`
	SensitiveMasked bool
	CreatedAt       time.Time
}

func (AuditEventRow) TableName() string { return "ragkit_audit_event" }
