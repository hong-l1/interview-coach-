package schemas

import (
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/eino-contrib/jsonschema"
)

// --- 简历---

type BaseInfo struct {
	Name         string `json:"name" jsonschema:"required,description=Candidate full name from the resume"`
	Gender       string `json:"gender" jsonschema:"description=Candidate gender if explicitly mentioned"`
	Age          string `json:"age" jsonschema:"description=Candidate age if explicitly mentioned"`
	WorkYears    string `json:"work_years" jsonschema:"description=Total years of work experience if explicitly mentioned"`
	JobIntention string `json:"job_intention" jsonschema:"description=Target job or desired position"`
	ExpectedCity string `json:"expected_city" jsonschema:"description=Expected work city"`
	Email        string `json:"email" jsonschema:"description=Email address"`
	Phone        string `json:"phone" jsonschema:"description=Phone number"`
}

type Education struct {
	School string `json:"school" jsonschema:"required,description=School or university name"`
	Degree string `json:"degree" jsonschema:"description=Degree such as bachelor or master"`
	Major  string `json:"major" jsonschema:"description=Major or field of study"`
	Start  string `json:"start" jsonschema:"description=Education start time in the resume, keep original format if needed"`
	End    string `json:"end" jsonschema:"description=Education end time in the resume, keep original format if needed"`
}

type WorkExp struct {
	Company          string   `json:"company" jsonschema:"description=Company name"`
	Position         string   `json:"position" jsonschema:"description=Position title"`
	Duration         string   `json:"duration" jsonschema:"description=Work time range"`
	Responsibilities []string `json:"responsibilities" jsonschema:"description=Key responsibilities or achievements as bullet points"`
}

type Project struct {
	Name         string   `json:"name" jsonschema:"required,description=Project name"`
	Role         string   `json:"role" jsonschema:"description=Candidate role in the project"`
	Duration     string   `json:"duration" jsonschema:"description=Project time range"`
	Description  string   `json:"description" jsonschema:"description=Short project overview"`
	Contribution []string `json:"contribution" jsonschema:"description=Key contributions or responsibilities as bullet points"`
	TechStack    []string `json:"tech_stack" jsonschema:"description=Technologies used in the project as an array of keywords"`
}

type Resume struct {
	BaseInfo       BaseInfo    `json:"base_info" jsonschema:"required,description=Basic personal information"`
	Education      []Education `json:"education" jsonschema:"required,description=All education experiences in the resume, use empty array if none"`
	WorkExperience []WorkExp   `json:"work_experience" jsonschema:"required,description=Formal work experiences only, use empty array if none"`
	Projects       []Project   `json:"projects" jsonschema:"required,description=All major project experiences, use empty array if none"`
	Skills         []string    `json:"skills" jsonschema:"required,description=Skill keywords only, not long sentences"`
	Certifications []string    `json:"certifications" jsonschema:"required,description=Certificates or official certifications only"`
	Other          []string    `json:"other" jsonschema:"required,description=Other useful information such as awards or honors"`
}

// --- 面试---

type InterviewQuestion struct {
	Question string `json:"question" jsonschema:"required,description=Exactly one interview question in plain text"`
}

// --- 预测---

type PredictionQuestion struct {
	Question        string `json:"question" jsonschema:"required,description=The generated interview practice question"`
	Focus           string `json:"focus" jsonschema:"required,description=What knowledge point or skill this question is testing"`
	ReferenceAnswer string `json:"reference_answer" jsonschema:"required,description=A concise but solid reference answer"`
	FollowUp        string `json:"follow_up" jsonschema:"required,description=A likely follow-up question for deeper probing"`
}

type PredictionResult struct {
	PredictQuestions []PredictionQuestion `json:"predict_questions" jsonschema:"required,description=Exactly five targeted practice questions"`
}

// --- 总体评价---

type EvaluationAll struct {
	Comment    string      `json:"comment" jsonschema:"required,description=Overall evaluation summary and improvement suggestions"`
	Dimensions []Dimension `json:"dimensions" jsonschema:"required,description=Evaluation dimensions in order"`
}
type Dimension struct {
	Name    string `json:"name" jsonschema:"required,description=Dimension name in Chinese"`
	Content string `json:"content" jsonschema:"required,description=Detailed evaluation comment for this dimension"`
	Score   int    `json:"score" jsonschema:"required,description=Integer score between 0 and 100"`
}

// --- 总体评价---

type EvaluationDetails struct {
	Evaluation []Comments `json:"evaluation" jsonschema:"required,description=Evaluation comments"`
}
type Comments struct {
	Question   string `json:"question"    jsonschema:"description=the question  being evaluated"`
	Answer     string `json:"answer"      jsonschema:"description=user response to the question"`
	Order      string `json:"order"       jsonschema:"description=the display order of this comment"`
	Score      int32  `json:"score"       jsonschema:"description=User response scoring situation"`
	KnowPoints string `json:"know_points" jsonschema:"description=key knowledge points involved in the question"`
	Strengths  string `json:"strengths"   jsonschema:"description=advantages observed in the answer"`
	Weaknesses string `json:"weaknesses"  jsonschema:"description=weaknesses observed in the answer"`
	Suggestion string `json:"suggestion"  jsonschema:"description=recommendations for further improvement"`
	Reference  string `json:"reference"   jsonschema:"description=reference answer for the question"`
}

func NewResponseFormat(v any) *openai.ChatCompletionResponseFormat {
	return &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
		JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
			Strict:     true,
			JSONSchema: jsonschema.Reflect(v),
		},
	}
}
