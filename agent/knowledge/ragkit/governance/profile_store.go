package governance

import (
	"context"
	"encoding/json"
	"errors"

	"awesomeProject4/agent/knowledge/ragkit/ragkitdb"
	"gorm.io/gorm"
)

// ProfileStore 抽象策略档案持久化，便于测试用内存 store。
type ProfileStore interface {
	GetActive(ctx context.Context) (*StrategyProfile, error)
	Activate(ctx context.Context, id uint64) (*StrategyProfile, error)
	Rollback(ctx context.Context, id uint64) (*StrategyProfile, error)
	Create(ctx context.Context, p *StrategyProfile) error
}

type GormProfileStore struct {
	db *gorm.DB
}

func NewGormProfileStore(db *gorm.DB) ProfileStore { return &GormProfileStore{db: db} }

func (s *GormProfileStore) GetActive(ctx context.Context) (*StrategyProfile, error) {
	var row ragkitdb.StrategyProfileRow
	err := s.db.WithContext(ctx).Where("status = ?", "active").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return rowToProfile(row), nil
}

func (s *GormProfileStore) Activate(ctx context.Context, id uint64) (*StrategyProfile, error) {
	return s.inTx(ctx, func(tx *gorm.DB) (*StrategyProfile, error) {
		// 当前 active → baseline
		if err := tx.Model(&ragkitdb.StrategyProfileRow{}).
			Where("status = ?", "active").Update("status", "baseline").Error; err != nil {
			return nil, err
		}
		if err := tx.Model(&ragkitdb.StrategyProfileRow{}).
			Where("id = ?", id).Update("status", "active").Error; err != nil {
			return nil, err
		}
		var row ragkitdb.StrategyProfileRow
		if err := tx.First(&row, id).Error; err != nil {
			return nil, err
		}
		return rowToProfile(row), nil
	})
}

func (s *GormProfileStore) Rollback(ctx context.Context, id uint64) (*StrategyProfile, error) {
	return s.inTx(ctx, func(tx *gorm.DB) (*StrategyProfile, error) {
		// 找 baseline
		var base ragkitdb.StrategyProfileRow
		if err := tx.Where("status = ?", "baseline").First(&base).Error; err != nil {
			return nil, err
		}
		// 当前 active → archived（简化：改回 candidate），baseline → active
		if err := tx.Model(&ragkitdb.StrategyProfileRow{}).
			Where("status = ?", "active").Update("status", "candidate").Error; err != nil {
			return nil, err
		}
		if err := tx.Model(&ragkitdb.StrategyProfileRow{}).
			Where("id = ?", base.ID).Update("status", "active").Error; err != nil {
			return nil, err
		}
		var row ragkitdb.StrategyProfileRow
		if err := tx.First(&row, base.ID).Error; err != nil {
			return nil, err
		}
		return rowToProfile(row), nil
	})
}

func (s *GormProfileStore) Create(ctx context.Context, p *StrategyProfile) error {
	row := profileToRow(p)
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return err
	}
	// 回填自增 ID，调用方（测试/CLI）依赖 Create 之后 p.ID 可用。
	p.ID = row.ID
	return nil
}

func (s *GormProfileStore) inTx(ctx context.Context, fn func(*gorm.DB) (*StrategyProfile, error)) (*StrategyProfile, error) {
	var out *StrategyProfile
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		p, err := fn(tx)
		out = p
		return err
	})
	return out, err
}

func rowToProfile(row ragkitdb.StrategyProfileRow) *StrategyProfile {
	p := &StrategyProfile{
		ID: row.ID, Name: row.Name, Status: row.Status, FusionStrategy: row.FusionStrategy,
	}
	_ = json.Unmarshal([]byte(row.TopKConfig), &p.TopKConfig)
	_ = json.Unmarshal([]byte(row.EvidenceGateThresholds), &p.EvidenceGateThresholds)
	return p
}

func profileToRow(p *StrategyProfile) ragkitdb.StrategyProfileRow {
	row := ragkitdb.StrategyProfileRow{
		ID: p.ID, Name: p.Name, Status: p.Status, FusionStrategy: p.FusionStrategy,
	}
	topK, _ := json.Marshal(p.TopKConfig)
	gate, _ := json.Marshal(p.EvidenceGateThresholds)
	row.TopKConfig = string(topK)
	row.EvidenceGateThresholds = string(gate)
	return row
}
