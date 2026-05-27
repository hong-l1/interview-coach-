package handler

import (
	"awesomeProject4/backend/api/utils"
	"awesomeProject4/backend/event"
	"awesomeProject4/backend/repository/dao"
	userservice "awesomeProject4/backend/service"
	"awesomeProject4/pkg/zapx"
	"github.com/IBM/sarama"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	l                        zapx.Log
	userService              *userservice.UserService
	userModelService         *userservice.UserModelService
	resumeService            *userservice.ResumeService
	interviewService         *userservice.InterviewRecordService
	InterviewDialogueService *userservice.InterviewDialogueService
	evaluationDetailService  *userservice.EvaluationDetailService
	predictionService        *userservice.PredictionService
	evaluationPublisher      *event.InterviewEvaluationProducer
	redisClient              redis.Cmdable
	kafka                    sarama.Client
	cancelRegistry           *CancelRegistry
}

func NewHandler(l zapx.Log, userSvc *userservice.UserService, userModelSvc *userservice.UserModelService,
	InterviewDialogueService *userservice.InterviewDialogueService, evaluationDetailService *userservice.EvaluationDetailService,
	predictionService *userservice.PredictionService, evaluationPublisher *event.InterviewEvaluationProducer,
	resumeSvc *userservice.ResumeService, interviewSvc *userservice.InterviewRecordService, dialogueDAO *dao.InterviewDialogueDAO, redisClient redis.Cmdable) *Handler {
	return &Handler{
		l:                        l,
		userService:              userSvc,
		userModelService:         userModelSvc,
		resumeService:            resumeSvc,
		interviewService:         interviewSvc,
		InterviewDialogueService: InterviewDialogueService,
		evaluationDetailService:  evaluationDetailService,
		predictionService:        predictionService,
		evaluationPublisher:      evaluationPublisher,
		redisClient:              redisClient,
		cancelRegistry:           NewCancelRegistry(),
	}
}

func (h *Handler) Register(server *gin.Engine) {
	server.Use(gin.Recovery())
	server.Use(utils.CorsMiddleWare())
	user := server.Group("/user")
	{
		user.POST("/register", h.registerUser)
		user.POST("/login", h.loginUser)
	}
	authorized := server.Group("/")
	authorized.Use(utils.JWTMiddleWare())
	{
		//简历
		resume := authorized.Group("/resume")
		{
			resume.GET("/list", h.ListResumes)
			resume.POST("/upload", h.uploadResume)
			resume.DELETE("/:id", h.DeleteResume)
			resume.PUT("/:id/default", h.SetDefaultResume)
		}

		mianshi := authorized.Group("/mianshi")
		{
			mianshi.POST("/stream/start", h.startInterview)
			mianshi.POST("/stream/resume", h.ReconnectInterview)
			mianshi.POST("/answer/submit", h.SubmitAnswer)
			mianshi.POST("/interview/end", h.EndInterview)
			mianshi.GET("/session/info", h.GetSession)
		}

		evaluation := authorized.Group("/evaluation")
		{
			evaluation.POST("/report", h.getEvaluationReport)
			evaluation.POST("/predict", h.getEvaluationPrediction)
		}
		contribution := authorized.Group("/contribute")
		{
			contribution.POST("/")
		}
	}
}
