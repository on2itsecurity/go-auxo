package zerotrust

import (
	"encoding/json"

	"github.com/on2itsecurity/go-auxo/utils"
)

// ReadinessQuestions holds all the questions
type ReadinessQuestions struct {
	CreateDate  int64       `json:"create_date"` // The date the question was created in Unix time
	Strategical []Questions `json:"strategical"` // The strategical questions
	Tactical    []Questions `json:"tactical"`    // The tactical questions
	Operational []Questions `json:"operational"` // The operational questions
	Scoping     []Questions `json:"scoping"`     // The scoping question(s)
	Version     int         `json:"version"`     // The version of the questions
}

// Questions hold the question format as is used in Strategical, Tactical, Operational
type Questions struct {
	Answers               map[int]string `json:"answers"`                 // The possible answers for the question (normally 1-5 CMMI)
	Caption               string         `json:"caption"`                 // The question itself
	Explanation           string         `json:"explanation"`             // The explanation of the question
	QuestionID            string         `json:"question_id"`             // The ID of the question
	QuestionScope         string         `json:"question_scope"`          // The scope of the question (strategical, tactical, operational)
	RequiredLevelGuidance string         `json:"required-level-guidance"` // The required level of guidance
}

// ReadinessAnswers is the body for the post request
type ReadinessAnswers struct {
	ID          string   `json:"id,omitempty"`         // The ID of the assessment
	Timestamp   int64    `json:"assessment_timestamp"` // The timestamp of the assessment in Unix time
	Version     int      `json:"version"`              // The version of the questions (see ReadinessQuestions)
	Strategical []Answer `json:"strategical"`          // The strategical answers
	Tactical    []Answer `json:"tactical"`             // The tactical answers
	Operational []Answer `json:"operational"`          // The operational answers
	Scoping     Scope    `json:"scope"`                // The scoping answer(s)
}

// Scope holds the scope, goal answer.
type Scope struct {
	Goal int `json:"goal"` // The goal/ambition in the number of desired protect surfaces
}

// Answer holds the answer for a question
type Answer struct {
	QuestionID string `json:"question_id"`      // The ID of the question (zee ReadinessQuestions)
	Actual     int    `json:"actual"`           // The actual condition (1-5 CMMI)
	Timestamp  string `json:"answer_timestamp"` // The timestamp of the answer in Unix time
	AnsweredBy string `json:"answered_by"`      // The e-mail (of the user) who answered the question
	Goal       int    `json:"goal"`             // The goal/ambition (1-5 CMMI)
	Comment    string `json:"comment"`          // Additional comment on the answer
}

// AssessmentSummary holds the summary of the assessments
type AssessmentSummary struct {
	ID                  string `json:"id"`                   // The ID of the assessment
	AssessmentTimestamp int64  `json:"assessment_timestamp"` // The timestamp of the assessment in Unix time
}

// --- Functions ---

// GetAssessmentQuestions will return all questions
func (zt *ZeroTrust) GetReadinessQuestions() (*ReadinessQuestions, error) {
	call := "get-base-questions"
	method := "GET"

	result, err := zt.apiClient.ApiCall(zt.apiEndpoint+call, method, "")

	if err != nil {
		return nil, err
	}

	questions, err := utils.UnwrapItems[ReadinessQuestions](result)

	if err != nil {
		return nil, err
	}

	return questions[0], nil
}

// PostAssessmentAnswers will post the answers (ReadinessAnswers as object)
// Returns the posted ReadinessAnswers object or an error.
func (zt *ZeroTrust) PostReadinessAnswers(answers ReadinessAnswers) (*ReadinessAnswers, error) {
	call := "create-assessment"
	method := "POST"

	data, err := json.Marshal(utils.WrapItems(answers))
	if err != nil {
		return nil, err
	}

	result, err := zt.apiClient.ApiCall(zt.apiEndpoint+call, method, string(data))

	if err != nil {
		return nil, err
	}

	response, err := utils.UnwrapItems[ReadinessAnswers](result)

	if err != nil {
		return nil, err
	}

	return response[0], nil
}

// GetAssessmentSummary will return the summary of the assessments
func (zt *ZeroTrust) GetReadinessAssessmentsSummary() ([]*AssessmentSummary, error) {
	call := "get-assessments-summary"
	method := "GET"

	result, err := zt.apiClient.ApiCall(zt.apiEndpoint+call, method, "")

	if err != nil {
		return nil, err
	}

	assessments, err := utils.UnwrapItems[AssessmentSummary](result)

	if err != nil {
		return nil, err
	}

	return assessments, nil
}

// GetAssessments will return all the assessments
func (zt *ZeroTrust) GetReadinessAssessments() ([]*ReadinessAnswers, error) {
	call := "get-assessments"
	method := "GET"

	result, err := zt.apiClient.ApiCall(zt.apiEndpoint+call, method, "")

	if err != nil {
		return nil, err
	}

	assessments, err := utils.UnwrapItems[ReadinessAnswers](result)

	if err != nil {
		return nil, err
	}

	return assessments, nil
}

// GetAssessmentByID will return the assessment with the given ID
func (zt *ZeroTrust) GetReadinessAssessmentByID(assessmentID string) (*ReadinessAnswers, error) {
	call := "get-assessment-by-id?id=" + assessmentID
	method := "GET"

	result, err := zt.apiClient.ApiCall(zt.apiEndpoint+call, method, "")

	if err != nil {
		return nil, err
	}

	assessments, err := utils.UnwrapItems[ReadinessAnswers](result)

	if err != nil {
		return nil, err
	}

	return assessments[0], nil
}

// DeleteAssessmentByID will delete the assessment with the given ID
func (zt *ZeroTrust) DeleteReadinessAssessmentByID(assessmentID string) error {
	call := "remove-assessment-by-id?id=" + assessmentID
	method := "POST"

	_, err := zt.apiClient.ApiCall(zt.apiEndpoint+call, method, "")

	if err != nil {
		return err
	}

	return nil
}

// DeleteAssessments will delete ALL assessments
func (zt *ZeroTrust) DeleteReadinessAssessments() error {
	assessments, err := zt.GetReadinessAssessmentsSummary()

	if err != nil {
		return err
	}

	for _, a := range assessments {
		err := zt.DeleteReadinessAssessmentByID(a.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
