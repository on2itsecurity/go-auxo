package caseintegration

import (
	"github.com/on2itsecurity/go-auxo/apiclient"
)

// CaseIntegration is the main object for the Case API
// Please use the NewCaseIntegration function to create a new object
type CaseIntegration struct {
	apiEndpoint string `default:"/v3/case/integration/"`
	apiClient   *apiclient.APIClient
}

// NewCaseIntegration creates a new CaseIntegration object to work with the API
// returns CaseIntegration object (pointer)
func NewCaseIntegration(address, token string, debug bool) *CaseIntegration {
	caseIntegration := new(CaseIntegration)
	caseIntegration.apiEndpoint = "/v3/case/"
	caseIntegration.apiClient, _ = apiclient.NewAPIClient(address, token, debug)
	return caseIntegration
}
