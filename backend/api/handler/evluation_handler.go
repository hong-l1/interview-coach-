package handler

import (
	"awesomeProject4/backend/api/utils"
	"awesomeProject4/backend/api/validate"
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) getEvaluationReport(c *gin.Context) {
	userID, ok := validate.ParseUserIDHeader(c)
	if !ok {
		return
	}
	recordID, ok := parseRecordID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	detail, err := h.evaluationDetailService.GetByRecordID(ctx, recordID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.NotFound(c, "evaluation report not generated")
		return
	}
	if err != nil {
		h.l.Error("get evaluation report failed")
		utils.InternalServerError(c, err.Error())
		return
	}
	if detail.UserID != uint(userID) {
		utils.Forbidden(c, "evaluation report is not yours")
		return
	}
	utils.Success(c, detail)
}

func (h *Handler) getEvaluationPrediction(c *gin.Context) {
	userID, ok := validate.ParseUserIDHeader(c)
	if !ok {
		return
	}
	recordID, ok := parseRecordID(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	prediction, err := h.predictionService.GetByRecordID(ctx, recordID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.NotFound(c, "prediction not generated")
		return
	}
	if err != nil {
		h.l.Error("get evaluation prediction failed")
		utils.InternalServerError(c, err.Error())
		return
	}
	if prediction.UserID != uint(userID) {
		utils.Forbidden(c, "prediction is not yours")
		return
	}
	questions, err := h.predictionService.ListQuestions(ctx, prediction.ID)
	if err != nil {
		h.l.Error("list prediction questions failed")
		utils.InternalServerError(c, err.Error())
		return
	}
	utils.Success(c, gin.H{
		"prediction": prediction,
		"questions":  questions,
	})
}

func parseRecordID(c *gin.Context) (uint64, bool) {
	var req struct {
		RecordID uint64 `json:"record_id"`
	}
	if err := c.ShouldBindJSON(&req); err == nil && req.RecordID != 0 {
		return req.RecordID, true
	}
	recordID := c.Query("record_id")
	if recordID == "" {
		utils.BadRequest(c, "record_id is required")
		return 0, false
	}
	id, err := strconv.ParseUint(recordID, 10, 64)
	if err != nil || id == 0 {
		utils.BadRequest(c, "invalid record_id")
		return 0, false
	}
	return id, true
}
