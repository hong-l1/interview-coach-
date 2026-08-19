package handler

import (
	"awesomeProject4/backend/api/utils"
	"awesomeProject4/backend/api/validate"
	"awesomeProject4/backend/pkg/zapx"
	"awesomeProject4/backend/repository/dao"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type resumeListItem struct {
	ID        uint64 `json:"id"`
	UserID    uint   `json:"user_id"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	FileType  string `json:"file_type"`
	IsDefault int    `json:"is_default"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func toResumeListItem(resume dao.Resume) resumeListItem {
	return resumeListItem{
		ID:        resume.ID,
		UserID:    resume.UserID,
		FileName:  resume.FileName,
		FileSize:  resume.FileSize,
		FileType:  resume.FileType,
		IsDefault: resume.IsDefault,
		CreatedAt: resume.CreatedAt.Unix(),
		UpdatedAt: resume.UpdatedAt.Unix(),
	}
}

func (h *Handler) ListResumes(c *gin.Context) {
	userID, ok := validate.ParseUserIDHeader(c)
	if !ok {
		return
	}

	resumes, err := h.resumeService.ListResumes(c.Request.Context(), uint(userID))
	if err != nil {
		h.l.Error("list resumes failed", zapx.Error(err), zapx.Int64("user_id", userID))
		utils.InternalServerError(c, err.Error())
		return
	}

	items := make([]resumeListItem, 0, len(resumes))
	for _, resume := range resumes {
		items = append(items, toResumeListItem(resume))
	}

	utils.SuccessWithMessage(c, "resume list", gin.H{
		"resumes": items,
		"total":   len(items),
	})
}

func (h *Handler) uploadResume(c *gin.Context) {
	userID, ok := validate.ParseUserIDHeader(c)
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "file is required")
		return
	}
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename)))
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		h.l.Error("save uploaded resume failed", zapx.Error(err), zapx.String("file_name", file.Filename))
		utils.InternalServerError(c, err.Error())
		return
	}
	defer os.Remove(tempPath)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	resume, _, err := h.resumeService.ParseResume(ctx, uint(userID), tempPath)
	if err != nil {
		h.l.Error("upload resume failed", zapx.Error(err), zapx.String("file_name", file.Filename))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "resume uploaded", toResumeListItem(*resume))
}

func (h *Handler) DeleteResume(c *gin.Context) {
	userID, ok := validate.ParseUserIDHeader(c)
	if !ok {
		return
	}

	id, ok := validate.ParseUintIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.resumeService.DeleteResume(c.Request.Context(), id, uint(userID)); err != nil {
		h.l.Error("delete resume failed", zapx.Error(err), zapx.Int64("id", int64(id)))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "resume deleted", gin.H{
		"id": id,
	})
}

func (h *Handler) SetDefaultResume(c *gin.Context) {
	userID, ok := validate.ParseUserIDHeader(c)
	if !ok {
		return
	}

	id, ok := validate.ParseUintIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.resumeService.SetDefaultResume(c.Request.Context(), id, uint(userID)); err != nil {
		h.l.Error("set default resume failed", zapx.Error(err), zapx.Int64("id", int64(id)))
		utils.InternalServerError(c, err.Error())
		return
	}

	utils.SuccessWithMessage(c, "resume set as default", gin.H{
		"id": id,
	})
}
