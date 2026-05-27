package service

import (
	agentSchema "awesomeProject4/agent/schema"
	"awesomeProject4/backend/repository/dao"
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type InterviewDialogueLister interface {
	ListByReportID(ctx context.Context, reportID uint64) ([]dao.InterviewDialogue, error)
}

type EvaluationDetailStore interface {
	Create(ctx context.Context, detail *dao.EvaluationDetail) error
	GetByRecordID(ctx context.Context, recordID uint64) (*dao.EvaluationDetail, error)
}

type PredictionCreator interface {
	Create(ctx context.Context, prediction *dao.Prediction) error
	BatchCreateQuestions(ctx context.Context, questions []*dao.PredictionQuestion) error
}

func EvaluationAndPredictService(
	ctx context.Context,
	userID uint,
	recordID uint64,
	reportID uint64,
	dialogueService InterviewDialogueLister,
	evaluationDetailService EvaluationDetailStore,
	predictionService PredictionCreator,
) error {
	existing, err := evaluationDetailService.GetByRecordID(ctx, recordID)
	if err == nil && existing != nil {
		return nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	dialogues, err := dialogueService.ListByReportID(ctx, reportID)
	if err != nil {
		return err
	}
	if len(dialogues) == 0 {
		return fmt.Errorf("no interview dialogues found for report_id=%d", reportID)
	}

	dialoguePrompt := buildDialoguePrompt(dialogues)
	detailResult, err := evaluateDetail(ctx, dialoguePrompt)
	if err != nil {
		return err
	}

	detail := &dao.EvaluationDetail{
		UserID:      userID,
		RecordID:    recordID,
		Evaluations: buildEvaluationItems(detailResult),
		Status:      "completed",
	}
	if err := evaluationDetailService.Create(ctx, detail); err != nil {
		return err
	}

	predictPrompt := buildPredictionPrompt(detailResult)
	predictResult, err := predict(ctx, predictPrompt)
	if err != nil {
		return err
	}

	prediction := &dao.Prediction{
		UserID:   userID,
		RecordID: recordID,
		ReportID: reportID,
	}
	if err := predictionService.Create(ctx, prediction); err != nil {
		return err
	}
	questions := make([]*dao.PredictionQuestion, 0, len(predictResult.PredictQuestions))
	for _, item := range predictResult.PredictQuestions {
		questions = append(questions, &dao.PredictionQuestion{
			PredictionID:    prediction.ID,
			Question:        item.Question,
			Focus:           item.Focus,
			ReferenceAnswer: item.ReferenceAnswer,
			FollowUp:        item.FollowUp,
		})
	}
	return predictionService.BatchCreateQuestions(ctx, questions)
}

func buildDialoguePrompt(dialogues []dao.InterviewDialogue) string {
	var b strings.Builder
	b.WriteString("请根据以下完整面试问答记录，逐题输出详细评价。\n")
	for i, dialogue := range dialogues {
		b.WriteString(fmt.Sprintf("\n第%d题\n问题：%s\n回答：%s\n", i+1, dialogue.Question, dialogue.Answer))
	}
	return b.String()
}

func buildPredictionPrompt(detail agentSchema.EvaluationDetails) string {
	var b strings.Builder
	b.WriteString("请根据以下逐题评价中暴露的薄弱点，生成5道针对性预测练习题。\n")
	for i, item := range detail.Evaluation {
		b.WriteString(fmt.Sprintf(
			"\n第%d题评价\n问题：%s\n得分：%d\n知识点：%s\n不足：%s\n建议：%s\n参考：%s\n",
			i+1,
			item.Question,
			item.Score,
			item.KnowPoints,
			item.Weaknesses,
			item.Suggestion,
			item.Reference,
		))
	}
	return b.String()
}

func buildEvaluationItems(result agentSchema.EvaluationDetails) []*dao.EvaluationItem {
	items := make([]*dao.EvaluationItem, 0, len(result.Evaluation))
	for _, item := range result.Evaluation {
		items = append(items, &dao.EvaluationItem{
			Message: []*dao.Dialogue{
				{
					Question: item.Question,
					Answer:   item.Answer,
				},
			},
			Comment: []*dao.Comment{
				{
					Score:      item.Score,
					KnowPoints: item.KnowPoints,
					Strengths:  item.Strengths,
					Weaknesses: item.Weaknesses,
					Suggestion: item.Suggestion,
					Reference:  item.Reference,
				},
			},
		})
	}
	return items
}
