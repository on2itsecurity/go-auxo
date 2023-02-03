package zerotrust

import (
	"encoding/json"

	"github.com/on2itsecurity/go-auxo/utils"
)

type Upsert struct {
	Protectsurface_uk        string        `json:"protectsurface_uniqueness_key,omitempty"` //Unique value. If no id is given it will be generated (only allowed on creation).
	Location_uk              string        `json:"location_uniqueness_key,omitempty"`       //Unique value. If no id is given it will be generated (only allowed on creation).
	Protectsurface_name      string        `json:"protectsurface_name"`                     //Name of the ProtectSurface
	Protectsurface_relevance int           `json:"protectsurface_relevance"`                //How important (0=not-100=very) is this protect service
	Location_name            string        `json:"location_name"`                           //Name of the location
	Location_coords          Coords        `json:"location_coords,omitempty"`               //Coordinates of the location
	States                   []UpsertState `json:"states"`                                  //Array of (minimal/upsert) states
}

type UpsertState struct {
	Description string    `json:"description,omitempty"`  //Description of the state
	ContentType string    `json:"content_type,omitempty"` //Contect Type: `ipv4`, `ipv6`, `azure_cloud`, `aws_cloud`, `gcp_cloud`, `container`, `hostname`, `user_identity`
	Content     *[]string `json:"content,omitempty"`      //The actual content f.e. `["10.10.10.1", "10.10.10.2"]`
}

// --- Functions ---

// Upsert creates a minimal protectsurface with location and state,
// and inserts the location and the protectsurface (and never update it if it already exists),
// and replaces the state in its entirety every time
func (zt *ZeroTrust) UpsertByObject(upsert Upsert) error {
	call := "upsert-protectsurface-location-state"
	method := "POST"

	data, err := json.Marshal(utils.WrapItems(upsert))
	if err != nil {
		return err
	}

	_, err = zt.apiClient.ApiCall(zt.apiEndpoint+call, method, string(data))

	if err != nil {
		return err
	}

	return nil
}
