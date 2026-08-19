//go:build wireinject
// +build wireinject

package backend

import (
	Init2 "awesomeProject4/backend/Init"
	"awesomeProject4/backend/api/handler"
	"awesomeProject4/backend/api/server"
	"awesomeProject4/backend/event"
	"awesomeProject4/backend/repository"
	"awesomeProject4/backend/repository/dao"
	"awesomeProject4/backend/service"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var providerSet = wire.NewSet(
	Init2.InitLogger,
	Init2.InitMysql,
	Init2.InitRedis,
	Init2.NewKafka,
	Init2.NewKafkaProducer,
	wire.Value("interview-evaluation-consumer"),
	Init2.NewConsumer,
	dao.NewUserDAO,
	dao.NewUserModelDAO,
	dao.NewResumeDAO,
	dao.NewInterviewRecordDAO,
	dao.NewInterviewDialogueDAO,
	dao.NewInterviewEvaluationDAO,
	dao.NewEvaluationDetailDAO,
	dao.NewPredictionDAO,
	repository.NewUserRepository,
	repository.NewUserModelRepository,
	repository.NewResumeRepository,
	repository.NewInterviewRecordRepository,
	repository.NewInterviewDialogueRepository,
	repository.NewInterviewEvaluationRepository,
	repository.NewEvaluationDetailRepository,
	repository.NewPredictionRepository,
	service.NewUserService,
	service.NewUserModelService,
	service.NewResumeService,
	service.NewInterviewRecordService,
	service.NewInterviewDialogueService,
	service.NewInterviewEvaluationService,
	service.NewEvaluationDetailService,
	service.NewPredictionService,
	event.NewInterviewEvaluationProducer,
	event.NewInterviewEvaluationConsumer,
	handler.NewHandler,
	server.NewGinEngine,
)

func InitializeServer() (*gin.Engine, sarama.ConsumerGroup, *event.InterviewEvaluationConsumer, error) {
	wire.Build(providerSet)
	return nil, nil, nil, nil
}
