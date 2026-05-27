package tool

import (
	"awesomeProject4/backend/repository/dao"
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GetResumeInput struct {
	ResumeID uint64 `json:"resume_id"`
}

type GetResumeOutput struct {
	Data map[string]any `json:"data"`
}

type ResumeGetter interface {
	GetResume(ctx context.Context, id uint64) (*dao.Resume, error)
}

type ResumeInfoGetter struct {
	resumeService ResumeGetter
}

func NewResumeInfoGetter(resumeService ResumeGetter) *ResumeInfoGetter {
	return &ResumeInfoGetter{
		resumeService: resumeService,
	}
}

func (g *ResumeInfoGetter) GetResumeInfo(ctx context.Context, req *GetResumeInput) (*GetResumeOutput, error) {
	if req == nil {
		return nil, fmt.Errorf("get_resume_info: request is nil")
	}
	resume, err := g.resumeService.GetResume(ctx, req.ResumeID)
	if err != nil {
		return nil, err
	}
	return &GetResumeOutput{
		Data: map[string]any{
			"id":         resume.ID,
			"user_id":    resume.UserID,
			"content":    resume.Content,
			"file_name":  resume.FileName,
			"file_size":  resume.FileSize,
			"file_type":  resume.FileType,
			"is_default": resume.IsDefault,
			"created_at": resume.CreatedAt,
			"updated_at": resume.UpdatedAt,
		},
	}, nil
}

func CreatGetResumeTool(resumeService ResumeGetter) tool.InvokableTool {
	getter := NewResumeInfoGetter(resumeService)
	tl, err := utils.InferTool(
		"get_resume_info",
		"Get parsed resume information from the database. Input requires resume_id.",
		getter.GetResumeInfo,
	)
	if err != nil {
		panic(err)
	}
	return tl
}
