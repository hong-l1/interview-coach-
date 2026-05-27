package Init

import (
	"awesomeProject4/backend/repository/dao"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var mysql_dsn = "root:root@tcp(43.142.57.35:13316)/my_interview?charset=utf8mb4&parseTime=True&loc=Local"

func InitMysql() (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(mysql_dsn), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(
		dao.InterviewRecord{},
		dao.InterviewDialogue{},
		dao.InterviewEvaluation{},
		dao.EvaluationDetail{},
		dao.Prediction{},
		dao.PredictionQuestion{},
		dao.Resume{},
		dao.User{},
		dao.UserModel{},
	)
	if err != nil {
		panic(err)
	}
	return db, nil
}
