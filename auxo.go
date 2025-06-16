package auxo

import (
	"github.com/on2itsecurity/go-auxo/asset"
	"github.com/on2itsecurity/go-auxo/caseintegration"
	"github.com/on2itsecurity/go-auxo/crm"
	"github.com/on2itsecurity/go-auxo/eventflow"
	"github.com/on2itsecurity/go-auxo/zerotrust"
)

// Client struct, contains the different "Auxo API sections"
type Client struct {
	Asset           *asset.Asset
	CaseIntegration *caseintegration.CaseIntegration
	CRM             *crm.CRM
	EventFlow       *eventflow.EventFlow
	ZeroTrust       *zerotrust.ZeroTrust
}

// NewClient constructor returns a client which can be used to make the calls.
// returns *client, error
func NewClient(address, token string, debug bool) (*Client, error) {
	client := new(Client)
	client.Asset = asset.NewAsset(address, token, debug)
	client.CaseIntegration = caseintegration.NewCaseIntegration(address, token, debug)
	client.CRM = crm.NewCRM(address, token, debug)
	client.EventFlow = eventflow.NewEventFlow(address, token, debug)
	client.ZeroTrust = zerotrust.NewZeroTrust(address, token, debug)

	return client, nil
}
