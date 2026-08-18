package ragkitdb

import "gorm.io/gorm"

// Migrate 注册 ragkit 两张表。由 CLI / 服务端各自调用，不进 InitMysql。
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&StrategyProfileRow{}, &AuditEventRow{})
}
