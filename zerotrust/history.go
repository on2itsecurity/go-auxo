package zerotrust

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/on2itsecurity/go-auxo/v2/utils"
)

// History holds the current content of every protect surface, state and location of the relation,
// each with the dated versions recorded at or after From.
//
// Entities are linked by uniqueness key only, internal IDs are not part of the history.
// A version holds the content as it was right AFTER the change at its timestamp,
// so the newest version equals the current content.
// Creation also records a version, at CreateDate. CreateDate is only returned for protect surfaces.
// A deletion also records a version (at DeletedAt), holding the content at the time of deletion.
// Entities deleted before From are not returned.
// The API returns the versions in pages, GetHistory requests all pages and merges them.
type History struct {
	From            time.Time                // Lower bound used for the request
	ProtectSurfaces []*ProtectSurfaceHistory // All protect surfaces, including the ones deleted at or after From
	States          []*StateHistory          // All states, including the ones deleted at or after From
	Locations       []*LocationHistory       // All locations, including the ones deleted at or after From
}

// HistoryMeta holds the identity and deletion status of an entity in the history
type HistoryMeta struct {
	RelationID    string `json:"relation_id"`    //Relation the entity belongs to
	UniquenessKey string `json:"uniqueness_key"` //Uniqueness key, stable over the life of the entity
	Deleted       bool   `json:"deleted"`        //Is the entity deleted
	DeletedAt     int64  `json:"deleted_at"`     //Unix time of the deletion, 0 if not deleted or unknown (deletions from before the history was recorded)
	CreateDate    int64  `json:"create_date"`    //Unix time of the creation, CreationRecordedFrom if unknown, 0 if not returned (states and locations)
}

// CreationRecordedFrom (16 September 2026 00:00 UTC) is the CreateDate of a protect surface whose creation time is unknown:
// no stored creation time and no time-based uniqueness key. Most were created before this moment, but a protect surface
// with a manually set uniqueness key created after it, before the API stored the creation time, gets it as well.
const CreationRecordedFrom int64 = 1789516800

// ProtectSurfaceHistory holds the current content of a protect surface plus its versions
type ProtectSurfaceHistory struct {
	HistoryMeta
	ProtectSurfaceVersion                         //Current content (Timestamp is 0)
	Versions              []ProtectSurfaceVersion `json:"snapshots"` //Versions recorded at or after From, newest first
}

// ProtectSurfaceVersion holds the content of a protect surface at a moment in time.
// Flows are keyed by the uniqueness key of the other protect surface (not by ID).
type ProtectSurfaceVersion struct {
	Timestamp               int64                   `json:"timestamp"`                  //Unix time this version was recorded, 0 for the current content
	Name                    string                  `json:"name"`                       //Name of the protect surface
	Description             string                  `json:"description"`                //Description of the protect surface
	MainContactPersonID     string                  `json:"main_contact_person_id"`     //Contact Person ID, empty if not set
	SecurityContactPersonID string                  `json:"security_contact_person_id"` //Security Person ID, empty if not set
	InControlBoundary       bool                    `json:"in_control_boundary"`        //Is this protect surface in the control boundary
	InZeroTrustFocus        bool                    `json:"in_zero_trust_focus"`        //Is this protect surface in focus
	Relevance               int                     `json:"relevance"`                  //How important (0=not-100=very) is this protect service
	Confidentiality         int                     `json:"confidentiality"`            //Score normally between 1-5
	Integrity               int                     `json:"integrity"`                  //Score normally between 1-5
	Availability            int                     `json:"availability"`               //Score normally between 1-5
	DataTags                []string                `json:"data_tags"`                  //What data holds this protect suface, f.e. PII, PCI
	ComplianceTags          []string                `json:"compliance_tags"`            //What compliance frameworks should be respected
	SocTags                 []string                `json:"soc_tags"`                   //Set by the SOC
	CustomerLabels          ObjectMap[string]       `json:"customer_labels"`            //Customer specific Key=Value pairs
	FlowsToOtherPS          ObjectMap[Flow]         `json:"flows_to_other_ps"`          //Other PS uniqueness key -> Flow
	FlowsFromOtherPS        ObjectMap[Flow]         `json:"flows_from_other_ps"`        //Other PS uniqueness key -> Flow
	FlowsToOutside          Flow                    `json:"flows_to_outside"`           //Flow
	FlowsFromOutside        Flow                    `json:"flows_from_outside"`         //Flow
	Measures                ObjectMap[MeasureState] `json:"measures"`                   //Measure name -> Measure
	Maturity                Maturity                `json:"maturity"`                   //Maturity (per zero trust step)
}

// StateHistory holds the current content of a state plus its versions
type StateHistory struct {
	HistoryMeta
	ProtectSurfaceUniquenessKey string         `json:"protectsurface_uniqueness_key"` //Owning protect surface, never changes over the life of the state
	StateVersion                               //Current content (Timestamp is 0)
	Versions                    []StateVersion `json:"snapshots"` //Versions recorded at or after From, newest first
}

// StateVersion holds the content of a state at a moment in time
type StateVersion struct {
	Timestamp             int64    `json:"timestamp"`               //Unix time this version was recorded, 0 for the current content
	Description           string   `json:"description"`             //Description of the state
	ContentType           string   `json:"content_type"`            //Content Type: `ipv4`, `ipv6`, `azure_cloud`, `aws_cloud`, `gcp_cloud`, `container`, `hostname`, `user_identity`
	LocationUniquenessKey string   `json:"location_uniqueness_key"` //Uniqueness key of the attached location, empty if not set
	ExistsOnAssetIDs      []string `json:"exists_on_asset_ids"`     //IDs of the managed assets providing this information
	Maintainer            string   `json:"maintainer"`              //Maintainer of the state, empty if not set
	Content               []string `json:"content"`                 //The actual content f.e. `["10.10.10.1", "10.10.10.2"]`
}

// LocationHistory holds the current content of a location plus its versions
type LocationHistory struct {
	HistoryMeta
	LocationVersion                   //Current content (Timestamp is 0)
	Versions        []LocationVersion `json:"snapshots"` //Versions recorded at or after From, newest first
}

// LocationVersion holds the content of a location at a moment in time
type LocationVersion struct {
	Timestamp int64  `json:"timestamp"` //Unix time this version was recorded, 0 for the current content
	Name      string `json:"name"`      //Name of the location
	Coords    Coords `json:"coords"`    //Coordinates
}

// ObjectMap is a map that also accepts an empty JSON array ([]) as an empty map,
// as the API returns empty objects as arrays in some places.
type ObjectMap[V any] map[string]V

// UnmarshalJSON accepts both a JSON object and an empty JSON array
func (m *ObjectMap[V]) UnmarshalJSON(data []byte) error {
	if isEmptyJSONArray(data) {
		*m = ObjectMap[V]{}
		return nil
	}

	var plain map[string]V
	if err := json.Unmarshal(data, &plain); err != nil {
		return err
	}
	*m = plain
	return nil
}

// UnmarshalJSON accepts both a JSON object and an empty JSON array (as an unset flow)
func (f *Flow) UnmarshalJSON(data []byte) error {
	if isEmptyJSONArray(data) {
		*f = Flow{}
		return nil
	}

	type plain Flow
	return json.Unmarshal(data, (*plain)(f))
}

func isEmptyJSONArray(data []byte) bool {
	return bytes.Equal(bytes.Join(bytes.Fields(data), nil), []byte("[]"))
}

// historyPayload is the single object returned (wrapped in items) by a page of the history call
type historyPayload struct {
	ProtectSurfaces struct {
		Items []*ProtectSurfaceHistory `json:"items"`
	} `json:"protectsurfaces"`
	States struct {
		Items []*StateHistory `json:"items"`
	} `json:"states"`
	Locations struct {
		Items []*LocationHistory `json:"items"`
	} `json:"locations"`
	HasMore bool `json:"has_more"`
}

// historyPageSize is the number of versions requested per page (maximum allowed by the API)
const historyPageSize = 500

// --- Functions ---

// GetHistory will get the history of all protect surfaces, states and locations of the relation (based upon used API Token),
// with every version recorded at or after from (no upper bound).
// All pages are requested and merged, consider passing a context with a generous timeout for an early from.
func (zt *ZeroTrust) GetHistory(ctx context.Context, from time.Time) (*History, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ts := from.Unix()
	if ts < 0 {
		ts = 0
	}

	method := "GET"
	h := newHistoryMerger(time.Unix(ts, 0))

	for page := 1; ; page++ {
		call := fmt.Sprintf("state/history/?timestamp_from=%d&page_number=%d&page_size=%d", ts, page, historyPageSize)

		result, err := zt.apiClient.ApiCall(ctx, zt.apiEndpoint+call, method, "")

		if err != nil {
			return nil, err
		}

		payload, err := parseHistoryPage(result)

		if err != nil {
			return nil, err
		}

		//Stop on an empty page as well, to never loop endlessly
		if versions := h.add(payload); !payload.HasMore || versions == 0 {
			break
		}
	}

	return h.history, nil
}

// parseHistoryPage unwraps a page of the history call response
func parseHistoryPage(data []byte) (*historyPayload, error) {
	payloads, err := utils.UnwrapItems[historyPayload](data)

	if err != nil {
		return nil, err
	}

	if len(payloads) != 1 {
		return nil, fmt.Errorf("Expected exactly one history item, got %d", len(payloads))
	}

	return payloads[0], nil
}

// historyMerger merges the pages of the history call into a single History.
// Every page repeats the current content of all entities with only the versions on that page,
// pages are newest first, so appending the versions keeps them newest first.
type historyMerger struct {
	history         *History
	protectSurfaces map[string]*ProtectSurfaceHistory
	states          map[string]*StateHistory
	locations       map[string]*LocationHistory
}

func newHistoryMerger(from time.Time) *historyMerger {
	return &historyMerger{
		history:         &History{From: from},
		protectSurfaces: make(map[string]*ProtectSurfaceHistory),
		states:          make(map[string]*StateHistory),
		locations:       make(map[string]*LocationHistory),
	}
}

// add merges a page and returns the number of versions on it
func (m *historyMerger) add(p *historyPayload) int {
	versions := 0

	for _, e := range p.ProtectSurfaces.Items {
		versions += len(e.Versions)
		if known, ok := m.protectSurfaces[e.UniquenessKey]; ok {
			known.Versions = append(known.Versions, e.Versions...)
			known.merge(e.HistoryMeta)
			continue
		}
		m.protectSurfaces[e.UniquenessKey] = e
		m.history.ProtectSurfaces = append(m.history.ProtectSurfaces, e)
	}

	for _, e := range p.States.Items {
		versions += len(e.Versions)
		if known, ok := m.states[e.UniquenessKey]; ok {
			known.Versions = append(known.Versions, e.Versions...)
			known.merge(e.HistoryMeta)
			continue
		}
		m.states[e.UniquenessKey] = e
		m.history.States = append(m.history.States, e)
	}

	for _, e := range p.Locations.Items {
		versions += len(e.Versions)
		if known, ok := m.locations[e.UniquenessKey]; ok {
			known.Versions = append(known.Versions, e.Versions...)
			known.merge(e.HistoryMeta)
			continue
		}
		m.locations[e.UniquenessKey] = e
		m.history.Locations = append(m.history.Locations, e)
	}

	return versions
}

// merge combines the deletion status of the same entity from another page,
// deleted_at is only set on the page holding the deletion
func (m *HistoryMeta) merge(other HistoryMeta) {
	if m.CreateDate == 0 {
		m.CreateDate = other.CreateDate
	}
	m.Deleted = m.Deleted || other.Deleted
	if other.DeletedAt > m.DeletedAt {
		m.DeletedAt = other.DeletedAt
	}
}
