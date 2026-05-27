package validate

import (
	"awesomeProject4/backend/api/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserModel struct {
	Name         string `json:"name"`
	ModelName    string `json:"model_name"`
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	ProviderName string `json:"provider_name"`
	IsDefault    int32  `json:"is_default"`
}

type InterviewQuestionRequest struct {
	InterviewType string `json:"interview_type" binding:"required"`
	Domain        string `json:"domain"`
	ResumeID      uint64 `json:"resume_id"`
	Difficulty    string `json:"difficulty"`
	Company       string `json:"company"`
	Position      string `json:"position"`
}

type SubmitAnswerRequest struct {
	SessionID string `json:"session_id"`
	Answer    string `json:"answer"`
}

func UserValidate(c *gin.Context) *User {
	var user User
	if err := c.Bind(&user); err != nil {
		utils.BadRequest(c, err.Error())
		return nil
	}
	if user.Email == "" {
		utils.BadRequest(c, "email is required")
		return nil
	}
	if user.Password == "" {
		utils.BadRequest(c, "password is required")
		return nil
	}
	return &user
}

func UserModelValidate(c *gin.Context) *UserModel {
	var userModel UserModel
	if err := c.Bind(&userModel); err != nil {
		utils.BadRequest(c, err.Error())
		return nil
	}
	if userModel.Name == "" || userModel.ModelName == "" || userModel.BaseURL == "" || userModel.APIKey == "" ||
		userModel.ProviderName == "" {
		utils.BadRequest(c, "some parms is required")
		return nil
	}
	if userModel.IsDefault != 0 && userModel.IsDefault != 1 {
		utils.BadRequest(c, "is_default must be 0 or 1")
		return nil
	}
	return &userModel
}

func InterviewQuestionValidate(c *gin.Context) *InterviewQuestionRequest {
	var req InterviewQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return nil
	}
	if req.InterviewType == "" {
		utils.BadRequest(c, "interview_type is required")
		return nil
	}
	return &req
}

func SubmitAnswerValidate(c *gin.Context) *SubmitAnswerRequest {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return nil
	}
	if req.SessionID == "" || req.Answer == "" {
		utils.BadRequest(c, "session id is nil or answer is nil")
		return nil
	}
	return &req
}

func ParseUintIDParam(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return id, true
}

func ParseSessionIDParam(c *gin.Context) (string, bool) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		utils.BadRequest(c, "session_id is required")
		return "", false
	}
	return sessionID, true
}

func ParseUserIDHeader(c *gin.Context) (int64, bool) {
	userID, err := strconv.ParseInt(c.GetHeader("userid"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "invalid userid")
		return 0, false
	}
	return userID, true
}

func ParseSessionID(c *gin.Context) (string, bool) {
	var sessionID string
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&req); err == nil {
		sessionID = req.SessionID
	}
	if sessionID == "" {
		utils.BadRequest(c, "session_id is required")
		return "", false
	}
	return sessionID, true
}
