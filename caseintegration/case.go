package caseintegration

import (
	"encoding/json"
	"fmt"

	"github.com/on2itsecurity/go-auxo/utils"
)

type Case struct {
	ID                  string `json:"id"`                    // The Unique ID for the case (to track it across systems)
	Subject             string `json:"subject"`               // The subject of the case
	Note                string `json:"note"`                  // Represents a note or update added to the case
	Priority            int    `json:"priority"`              // Case priority 1 to 4, where 1 is the highest priority
	CaseType            string `json:"case_type"`             // The type of the case; `securityincident`, `incident`, `change`, `standardchange`, `inforequest`.
	PrimaryContactEmail string `json:"primary_contact_email"` // The email address of the primary contact for the case, this should match a user in the system.
	Attachment          string `json:"attachment,omitempty"`  // A base64-encoded file attachment as a data URI. The MIME and encoding type (e.g., data:application/vnd.openxmlformats-officedocument.spreadsheetml.sheet;base64) should be included in the data URI.
}

// AddNoteToCase adds a note to an existing case
func (ci *CaseIntegration) AddNoteToCase(caseID string, note string) error {
	call := "integration/" + caseID + "/notes"

	method := "POST"

	data, err := json.Marshal(utils.WrapItems(map[string]string{"note": note}))
	if err != nil {
		return err
	}

	// Will return an empty response
	_, err = ci.apiClient.ApiCall(ci.apiEndpoint+call, method, string(data))

	if err != nil {
		return err
	}

	return nil
}

// CreateCaseByObject creates a new case in the system
func (ci *CaseIntegration) CreateCaseByObject(c Case) error {
	call := "integration"

	method := "POST"

	data, err := json.Marshal(utils.WrapItems(c))
	if err != nil {
		return err
	}

	// Will return an empty response
	_, err = ci.apiClient.ApiCall(ci.apiEndpoint+call, method, string(data))

	if err != nil {
		return err
	}

	return nil
}

// DeescalateCase deescalates an existing case
func (ci *CaseIntegration) DeescalateCase(caseID string) error {
	call := "integration/" + caseID + "/escalation-status"

	method := "DELETE"

	// Will return an empty response
	_, err := ci.apiClient.ApiCall(ci.apiEndpoint+call, method, "")

	if err != nil {
		return err
	}

	return nil
}

// EscalateCase escalates an existing case
func (ci *CaseIntegration) EscalateCase(caseID string) error {
	call := "integration/" + caseID + "/escalation-status"

	method := "PATCH"

	// Will return an empty response
	_, err := ci.apiClient.ApiCall(ci.apiEndpoint+call, method, "")

	if err != nil {
		return err
	}

	return nil
}

// GetCaseByID Get case by ID
func (ci *CaseIntegration) GetCaseByID(caseID string) (*Case, error) {
	call := "integration/" + caseID

	method := "GET"

	response, err := ci.apiClient.ApiCall(ci.apiEndpoint+call, method, "")
	if err != nil {
		return nil, err
	}

	cases, err := utils.UnwrapItems[Case](response)
	if err != nil {
		return nil, err
	}

	if len(cases) == 0 {
		return nil, fmt.Errorf("Case with ID %s not found", caseID)
	}

	return cases[0], nil
}

// RequestCaseClose Requests the closure of an existing case
func (ci *CaseIntegration) RequestCaseClose(caseID string) error {
	call := "integration/" + caseID + "/request-close"

	method := "POST"

	// Will return an empty response
	_, err := ci.apiClient.ApiCall(ci.apiEndpoint+call, method, "")

	if err != nil {
		return err
	}

	return nil
}

// UpdatePrimaryContactOfCase updates the primary contact of an existing case
func (ci *CaseIntegration) UpdatePrimaryContactOfCase(caseID string, email string) error {
	call := "integration/" + caseID + "/primary-contact"

	method := "PATCH"

	data, err := json.Marshal(utils.WrapItems(map[string]string{"primary_contact_email": email}))
	if err != nil {
		return err
	}

	// Will return an empty response
	_, err = ci.apiClient.ApiCall(ci.apiEndpoint+call, method, string(data))

	if err != nil {
		return err
	}

	return nil
}

// UpdatePriorityOfCase updates the priority (1-4) of an existing case
func (ci *CaseIntegration) UpdatePriorityOfCase(caseID string, priority int) error {
	call := "integration/" + caseID + "/priority"

	method := "PATCH"

	data, err := json.Marshal(utils.WrapItems(map[string]int{"priority": priority}))
	if err != nil {
		return err
	}

	// Will return an empty response
	_, err = ci.apiClient.ApiCall(ci.apiEndpoint+call, method, string(data))

	if err != nil {
		return err
	}

	return nil
}

// UpdateSubjectOfCase updates the subject of an existing case
func (ci *CaseIntegration) UpdateSubjectOfCase(caseID string, subject string) error {
	call := "integration/" + caseID + "/subject"

	method := "PATCH"

	data, err := json.Marshal(utils.WrapItems(map[string]string{"subject": subject}))
	if err != nil {
		return err
	}

	// Will return an empty response
	_, err = ci.apiClient.ApiCall(ci.apiEndpoint+call, method, string(data))

	if err != nil {
		return err
	}

	return nil
}
