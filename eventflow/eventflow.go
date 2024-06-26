package eventflow

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/on2itsecurity/go-auxo/apiclient"
)

// EventFlow is the main object for the EventFlow / Generic Event API
// Please use the NewEventFlow function to create a new object
type EventFlow struct {
	apiEndpoint string `default:"/v3/eventflow/"`
	apiClient   *apiclient.APIClient
}

// Event is the interface for the different event types
type Event interface {
	EventType() string
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

// EventType returns the type of the event
func (o Other) EventType() string {
	return o.Type
}

// EventType returns the type of the event
func (t ThreatNetwork) EventType() string {
	return t.Type
}

// EventQueue is a queue of events
type EventQueue struct {
	events []Event
}

// AddEventToQueue will add an event to the queue
// By using a queue, multiple events can be added at once, with the PostEventQueue function
func (q *EventQueue) AddEventToQueue(e Event) {
	q.events = append(q.events, e)
}

// GetEventQueue will return the queue of events
func (q *EventQueue) GetEventQueue() []Event {
	return q.events
}

// ClearEventQueue will empty the queue
func (q *EventQueue) ClearEventQueue() {
	q.events = nil
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

// PostEventQueue will post the events in the queue to the EventFlow API and empty the queue
func (e *EventFlow) PostEventQueue(q *EventQueue) error {

	if len(q.GetEventQueue()) == 0 {
		return fmt.Errorf("No events in queue")
	}

	postData := ""
	for _, event := range q.GetEventQueue() {
		//Convert to base64 and add to postData
		byteOutput, _ := json.Marshal(event)
		base64Output := base64.StdEncoding.EncodeToString(byteOutput)
		postData += base64Output + "\n"
	}

	//Clear eventQueue
	q.events = nil

	err := e.StoreEventsInBase64(postData)

	return err
}

// StoreEvents will post the events to eventflow.
// The events needs to be given in base64, multiple events at once, can be seperated by a new-line (`\n`).
// When switching from the AddEvent method a new API token is required.
func (e *EventFlow) StoreEventsInBase64(event string) error {
	uri := "store-events"

	_, err := e.apiClient.ApiCall(e.apiEndpoint+uri, "POST", event)

	if err != nil {
		return err
	}

	return nil
}

// AddEvent will post the event to eventflow, the event needs to given in base64.
// DEPRECATED: Use StoreEventsInBase64 instead
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
