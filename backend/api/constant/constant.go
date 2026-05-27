package constant

import "time"

const (
	MaxInterviewQuestions        = 20
	DialogueBatchSize            = 5
	DialogueFlushInterval        = 2 * time.Minute
	InterviewMsgListTTl          = time.Minute
	Start                        = -10
	End                          = -1
	QuestionConsumeBlockDuration = 60 * time.Second
	EndedInterviewSessionTTL     = 60 * time.Second
)
