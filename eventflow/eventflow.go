package eventflow

import (
	"fmt"

	"github.com/on2itsecurity/go-auxo/apiclient"
)

// EventFlow is the main object for the EventFlow / Generic Event API
// Please use the NewEventFlow function to create a new object
type EventFlow struct {
	apiEndpoint string `default:"/v3/eventflow/"`
	apiClient   *apiclient.APIClient
}

// Other represents the "other-event-type"
type Other struct {
	Type               string      `json:"type"`                // The type of the event
	DetectionTimestamp int64       `json:"detection_timestamp"` // The timestamp of the detection in Unix time
	Message            string      `json:"message"`             // The (short) message of the event
	Vendor             string      `json:"vendor"`              // The vendor of the event
	VendorEventID      string      `json:"vendor_event_id"`     // The vendor event (unique) ID
	Raw                interface{} `json:"raw"`                 // The raw event
}

// ThreatNetwork represents the "threat_network-event-type"
type ThreatNetwork struct {
	Type               string `json:"type"`                // The type of the event
	DetectionTimestamp int64  `json:"detection_timestamp"` // The timestamp of the detection in Unix time
	Severity           int    `json:"severity"`            // The severity of the event
	Blocked            bool   `json:"blocked"`             // The event is blocked or not
	Message            string `json:"message"`             // The (short) message of the event
	Vendor             string `json:"vendor"`              // The vendor of the event
	VendorEventID      string `json:"vendor_event_id"`     // The vendor event (unique) ID
	Attacker           Entity `json:"attacker"`            // The attacker
	Victim             Entity `json:"victim"`              // The victim
	Raw                string `json:"raw"`                 // The raw event
}

// Entity represents an Attacker of Victim
type Entity struct {
	Type  string `json:"type"`  // The type of the entity i.e. container-id or IPv4
	Value string `json:"value"` // The value of the entity
}

// NewZeroTrust creates a new ZeroTrust object to work with the API
// returns ZeroTrust object (pointer)
func NewEventFlow(address, token string, debug bool) *EventFlow {
	ef := new(EventFlow)
	ef.apiEndpoint = "/v3/eventflow/"
	ef.apiClient, _ = apiclient.NewAPIClient(address, token, debug)
	return ef
}

// AddEvent will post the event to eventflow, the event needs to given in base64.
// Multiple events can be added at once, by passing them in base64 seperated by a new-line (`\n`).
func (e *EventFlow) AddEvent(assetID string, eventInBase64 string) error {
	sensordatatype := "on2it_generic_webhook"
	uri := fmt.Sprintf("store?sensorid=on2itassetid:%s&sensordatatype=%s", assetID, sensordatatype)

	_, err := e.apiClient.ApiCall(e.apiEndpoint+uri, "POST", eventInBase64)

	if err != nil {
		return err
	}

	return nil
}

// SetTimeout, when large calls are created to add events, it might be usefull to extend the timeout.
func (e *EventFlow) SetTimeout(seconds int) {
	e.apiClient.SetTimeout(seconds)
}
