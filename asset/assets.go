package asset

import "github.com/on2itsecurity/go-auxo/utils"

// AssetItem holds all the fields for the asset "object"
type AssetItem struct {
	ID              string `json:"id"`                     //The ID of the asset
	Name            string `json:"name"`                   //The name of the asset
	TypeName        string `json:"asset_type_name"`        //The type of the asset
	TypeDescription string `json:"asset_type_description"` //The description of the asset type
	IP              string `json:"ip"`                     //The IP address of the asset
	Port            string `json:"port"`                   //The port of the asset
	Status          string `json:"status"`                 //The status of the asset
}

// GetContacts will get all contacts of the relation (based on used API Token)
// It returns an array with all the Location objects.
func (asset *Asset) GetAssets() ([]*AssetItem, error) {
	call := "get-assets"
	method := "GET"

	result, err := utils.GetAllPages[AssetItem](asset.apiEndpoint+call, method, asset.apiClient)

	if err != nil {
		return nil, err
	}

	return result, nil
}
