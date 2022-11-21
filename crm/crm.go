package crm

import (
	"github.com/on2itsecurity/go-auxo/apiclient"
)

// CRM is the main object for the CRM API
// Please use the NewCRM function to create a new object
type CRM struct {
	apiEndpoint string `default:"/v3/crm/"`
	apiClient   *apiclient.APIClient
}

// NewCRM creates a new CRM object to work with the API
// returns CRM object (pointer)
func NewCRM(address, token string, debug bool) *CRM {
	crm := new(CRM)
	crm.apiEndpoint = "/v3/crm/"
	crm.apiClient, _ = apiclient.NewAPIClient(address, token, debug)
	return crm
}
