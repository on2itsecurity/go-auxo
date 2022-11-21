package asset

import (
	"github.com/on2itsecurity/go-auxo/apiclient"
)

// Asset is the main object for the Asset API
// Please use the NewAsset function to create a new object
type Asset struct {
	apiEndpoint string `default:"/v3/asset/"`
	apiClient   *apiclient.APIClient
}

// NewAsset creates a new Asset object to work with the API
// returns Asset object (pointer)
func NewAsset(address, token string, debug bool) *Asset {
	asset := new(Asset)
	asset.apiEndpoint = "/v3/asset/"
	asset.apiClient, _ = apiclient.NewAPIClient(address, token, debug)
	return asset
}
