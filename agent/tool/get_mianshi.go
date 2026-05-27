package tool

import (
	"awesomeProject4/backend/repository/dao"
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type MianShiInfoInput struct {
	ReportID uint64 `json:"report_id"`
}

type MianShiInfoOutput struct {
	Data map[string]any `json:"data"`
}

type MianShiDialogueGetter interface {
	ListByReportID(ctx context.Context, reportID uint64) ([]dao.InterviewDialogue, error)
}

type MianShiInfoGetter struct {
	dialogueService MianShiDialogueGetter
}

func NewMianShiInfoGetter(dialogueService MianShiDialogueGetter) *MianShiInfoGetter {
	return &MianShiInfoGetter{
		dialogueService: dialogueService,
	}
}

func (g *MianShiInfoGetter) GetMianShiInfo(ctx context.Context, req *MianShiInfoInput) (*MianShiInfoOutput, error) {
	if req == nil {
		return nil, fmt.Errorf("get_mianshi_info: request is nil")
	}
	dialogues, err := g.dialogueService.ListByReportID(ctx, req.ReportID)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(dialogues))
	for _, dialogue := range dialogues {
		items = append(items, map[string]any{
			"id":         dialogue.ID,
			"user_id":    dialogue.UserID,
			"report_id":  dialogue.ReportID,
			"question":   dialogue.Question,
			"answer":     dialogue.Answer,
			"created_at": dialogue.CreatedAt,
		})
	}
	return &MianShiInfoOutput{
		Data: map[string]any{
			"report_id":  req.ReportID,
			"dialogues":  items,
			"total_size": len(items),
		},
	}, nil
}

func CreatGetMianShiInfoTool(dialogueService MianShiDialogueGetter) tool.InvokableTool {
	getter := NewMianShiInfoGetter(dialogueService)
	tl, err := utils.InferTool(
		"get_mianshi_info",
		"Get the full interview dialogue history from the database. Input requires report_id.",
		getter.GetMianShiInfo,
	)
	if err != nil {
		panic(err)
	}
	return tl
}
