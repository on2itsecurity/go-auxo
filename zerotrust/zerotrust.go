package zerotrust

import (
	"github.com/on2itsecurity/go-auxo/apiclient"
)

// ZeroTrust is the main object for the ZeroTrust API
// Please use the NewZeroTrust function to create a new object
type ZeroTrust struct {
	apiEndpoint string `default:"/v3/protectsurface/"`
	apiClient   *apiclient.APIClient
}

// NewZeroTrust creates a new ZeroTrust object to work with the API
// returns ZeroTrust object (pointer)
func NewZeroTrust(address, token string, debug bool) *ZeroTrust {
	zt := new(ZeroTrust)
	zt.apiEndpoint = "/v3/zerotrust/"
	zt.apiClient, _ = apiclient.NewAPIClient(address, token, debug)
	return zt
}
