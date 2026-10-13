package zerotrust

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// PointInTime holds the content of all protect surfaces, states and locations at a given moment,
// reconstructed from a History.
type PointInTime struct {
	At              time.Time           // Moment of the reconstruction
	ProtectSurfaces []*ProtectSurfaceAt // Protect surfaces not deleted at At
	States          []*StateAt          // States not deleted at At
	Locations       []*LocationAt       // Locations not deleted at At
}

// ProtectSurfaceAt holds the content of a protect surface at a given moment
type ProtectSurfaceAt struct {
	UniquenessKey string
	ProtectSurfaceVersion
	Exact bool // false: the entity was first recorded after the moment, the content shown is its oldest recorded version
}

// StateAt holds the content of a state at a given moment
type StateAt struct {
	UniquenessKey               string
	ProtectSurfaceUniquenessKey string
	StateVersion
	Exact bool // false: the entity was first recorded after the moment, the content shown is its oldest recorded version
}

// LocationAt holds the content of a location at a given moment
type LocationAt struct {
	UniquenessKey string
	LocationVersion
	Exact bool // false: the entity was first recorded after the moment, the content shown is its oldest recorded version
}

// EntityType is the type of entity in the history
type EntityType string

const (
	EntityProtectSurface EntityType = "protectsurface"
	EntityState          EntityType = "state"
	EntityLocation       EntityType = "location"
)

// ChangeType is the type of a change
type ChangeType string

const (
	ChangeUpdate ChangeType = "update" // Content was changed
	ChangeDelete ChangeType = "delete" // Entity was deleted
)

// Change describes a single recorded change of an entity
type Change struct {
	Timestamp     int64      // Unix time of the change
	EntityType    EntityType // Type of the changed entity
	UniquenessKey string     // Uniqueness key of the changed entity
	Name          string     // Name (protect surface, location) or description (state) of the entity
	Type          ChangeType // Update or delete
	// Changed fields as json names, map entries as "field.key" (f.e. "measures.data-dlp").
	// nil when Before is unknown, empty when no visible field changed.
	Fields []string
	// Content before and after the change: *ProtectSurfaceVersion, *StateVersion or *LocationVersion.
	// Before is nil for the first recorded change (content before it was not recorded), After is nil for a delete.
	Before any
	After  any
}

// versioned is implemented by the version types
type versioned interface {
	ProtectSurfaceVersion | StateVersion | LocationVersion
}

// --- Functions ---

// AsOf reconstructs the content of all entities at moment t, which must not be before h.From.
//
// Limitations of the recorded history:
//   - Creation is not recorded, entities created after t but never deleted are included.
//   - The content before the first recorded change is unknown, those entities are returned with Exact set to false
//     and the content of their oldest recorded version.
func (h *History) AsOf(t time.Time) (*PointInTime, error) {
	if t.Before(h.From) {
		return nil, fmt.Errorf("Cannot reconstruct %s, history starts at %s", t.Format(time.RFC3339), h.From.Format(time.RFC3339))
	}

	ts := t.Unix()
	pit := &PointInTime{At: t}

	for _, ps := range h.ProtectSurfaces {
		if v, exact, ok := versionAt(ps.HistoryMeta, ps.ProtectSurfaceVersion, ps.Versions, ts); ok {
			pit.ProtectSurfaces = append(pit.ProtectSurfaces, &ProtectSurfaceAt{UniquenessKey: ps.UniquenessKey, ProtectSurfaceVersion: v, Exact: exact})
		}
	}

	for _, s := range h.States {
		if v, exact, ok := versionAt(s.HistoryMeta, s.StateVersion, s.Versions, ts); ok {
			pit.States = append(pit.States, &StateAt{UniquenessKey: s.UniquenessKey, ProtectSurfaceUniquenessKey: s.ProtectSurfaceUniquenessKey, StateVersion: v, Exact: exact})
		}
	}

	for _, l := range h.Locations {
		if v, exact, ok := versionAt(l.HistoryMeta, l.LocationVersion, l.Versions, ts); ok {
			pit.Locations = append(pit.Locations, &LocationAt{UniquenessKey: l.UniquenessKey, LocationVersion: v, Exact: exact})
		}
	}

	return pit, nil
}

// Changes returns all recorded changes, oldest first.
func (h *History) Changes() []*Change {
	changes := make([]*Change, 0)

	for _, ps := range h.ProtectSurfaces {
		changes = append(changes, entityChanges(EntityProtectSurface, ps.HistoryMeta, ps.Versions, func(v *ProtectSurfaceVersion) string { return v.Name })...)
	}

	for _, s := range h.States {
		changes = append(changes, entityChanges(EntityState, s.HistoryMeta, s.Versions, func(v *StateVersion) string { return v.Description })...)
	}

	for _, l := range h.Locations {
		changes = append(changes, entityChanges(EntityLocation, l.HistoryMeta, l.Versions, func(v *LocationVersion) string { return v.Name })...)
	}

	sort.SliceStable(changes, func(i, j int) bool {
		if changes[i].Timestamp != changes[j].Timestamp {
			return changes[i].Timestamp < changes[j].Timestamp
		}
		if changes[i].EntityType != changes[j].EntityType {
			return changes[i].EntityType < changes[j].EntityType
		}
		return changes[i].UniquenessKey < changes[j].UniquenessKey
	})

	return changes
}

// versionAt returns the content of an entity at ts, whether it is exact and whether the entity existed (was not deleted).
// A version holds the content after the change at its timestamp, so the content at ts is the newest version at or before ts.
func versionAt[V versioned](meta HistoryMeta, current V, versions []V, ts int64) (V, bool, bool) {
	var none V
	if meta.Deleted && meta.DeletedAt <= ts {
		return none, false, false
	}

	var before, after *V
	for i := range versions {
		vts := versionTimestamp(&versions[i])
		if vts <= ts && (before == nil || vts > versionTimestamp(before)) {
			before = &versions[i]
		}
		if vts > ts && (after == nil || vts < versionTimestamp(after)) {
			after = &versions[i]
		}
	}

	switch {
	case before != nil:
		return *before, true, true
	case after == nil:
		return current, true, true //Not changed since From
	default:
		return *after, false, true //Content before the first recorded change is unknown
	}
}

// entityChanges converts the versions of a single entity to changes
func entityChanges[V versioned](et EntityType, meta HistoryMeta, versions []V, name func(*V) string) []*Change {
	asc := make([]*V, len(versions))
	for i := range versions {
		asc[i] = &versions[i]
	}
	sort.SliceStable(asc, func(i, j int) bool { return versionTimestamp(asc[i]) < versionTimestamp(asc[j]) })

	changes := make([]*Change, 0, len(asc))
	for i, v := range asc {
		c := &Change{
			Timestamp:     versionTimestamp(v),
			EntityType:    et,
			UniquenessKey: meta.UniquenessKey,
			Name:          name(v),
			Type:          ChangeUpdate,
			After:         v,
		}

		if meta.Deleted && i == len(asc)-1 && c.Timestamp == meta.DeletedAt {
			//The deletion records the content at the time of deletion
			c.Type = ChangeDelete
			c.Before = v
			c.After = nil
		} else if i > 0 {
			c.Before = asc[i-1]
			c.Fields = diffFields(*asc[i-1], *v)
		}

		changes = append(changes, c)
	}

	return changes
}

// versionTimestamp returns the timestamp of a version
func versionTimestamp[V versioned](v *V) int64 {
	switch x := any(v).(type) {
	case *ProtectSurfaceVersion:
		return x.Timestamp
	case *StateVersion:
		return x.Timestamp
	case *LocationVersion:
		return x.Timestamp
	}
	return 0
}

// diffFields returns the json names of the fields that differ between a and b (Timestamp excluded),
// map fields are compared per key and reported as "field.key"
func diffFields[V versioned](a, b V) []string {
	fields := make([]string, 0)
	va, vb := reflect.ValueOf(a), reflect.ValueOf(b)

	for i := 0; i < va.NumField(); i++ {
		f := va.Type().Field(i)
		name := strings.Split(f.Tag.Get("json"), ",")[0]
		if name == "timestamp" {
			continue
		}

		fa, fb := va.Field(i), vb.Field(i)
		if f.Type.Kind() == reflect.Map {
			keys := make(map[string]bool)
			for _, k := range append(fa.MapKeys(), fb.MapKeys()...) {
				keys[k.String()] = true
			}
			changed := make([]string, 0)
			for k := range keys {
				ka, kb := fa.MapIndex(reflect.ValueOf(k)), fb.MapIndex(reflect.ValueOf(k))
				if ka.IsValid() != kb.IsValid() || (ka.IsValid() && !reflect.DeepEqual(ka.Interface(), kb.Interface())) {
					changed = append(changed, name+"."+k)
				}
			}
			sort.Strings(changed)
			fields = append(fields, changed...)
			continue
		}

		if f.Type.Kind() == reflect.Slice && fa.Len() == 0 && fb.Len() == 0 {
			continue //nil and empty are equal
		}

		if !reflect.DeepEqual(fa.Interface(), fb.Interface()) {
			fields = append(fields, name)
		}
	}

	return fields
}

// --- Conversion to writable objects ---

// ToProtectSurface converts the version to a ProtectSurface which can be written with CreateProtectSurfaceByObject or UpdateProtectSurface.
// psIDs maps protect surface uniqueness keys to IDs, used to convert the flows to other protect surfaces.
// The ID is not set, set it to update an existing protect surface.
// Note that the history does not contain the risk accepted comment and who/when determined a measure.
func (v ProtectSurfaceVersion) ToProtectSurface(uniquenessKey string, psIDs map[string]string) (*ProtectSurface, error) {
	flowsTo, err := flowsByID(v.FlowsToOtherPS, psIDs)
	if err != nil {
		return nil, err
	}

	flowsFrom, err := flowsByID(v.FlowsFromOtherPS, psIDs)
	if err != nil {
		return nil, err
	}

	measures := make(map[string]MeasureState, len(v.Measures))
	for name, m := range v.Measures {
		measures[name] = m.ToMeasureState()
	}

	labels := make(map[string]string, len(v.CustomerLabels))
	for k, l := range v.CustomerLabels {
		labels[k] = l
	}

	return &ProtectSurface{
		UniquenessKey:           uniquenessKey,
		Name:                    v.Name,
		Description:             v.Description,
		MainContactPersonID:     v.MainContactPersonID,
		SecurityContactPersonID: v.SecurityContactPersonID,
		InControlBoundary:       v.InControlBoundary,
		InZeroTrustFocus:        v.InZeroTrustFocus,
		Relevance:               v.Relevance,
		Confidentiality:         v.Confidentiality,
		Integrity:               v.Integrity,
		Availability:            v.Availability,
		DataTags:                v.DataTags,
		ComplianceTags:          v.ComplianceTags,
		CustomerLabels:          labels,
		SocTags:                 v.SocTags,
		FlowsFromOtherPS:        flowsFrom,
		FlowsToOtherPS:          flowsTo,
		FlowsFromOutside:        v.FlowsFromOutside,
		FlowsToOutside:          v.FlowsToOutside,
		Measures:                measures,
		Maturity:                v.Maturity,
	}, nil
}

// ToMeasureState converts the flattened history measure to a MeasureState
func (m HistoryMeasure) ToMeasureState() MeasureState {
	ms := MeasureState{Assignment: &Assignment{Assigned: m.Assigned}}

	if m.Implemented != nil {
		ms.Implementation = &Implementation{Implemented: *m.Implemented}
	}

	if m.Evidenced != nil {
		ms.Evidence = &Evidence{Evidenced: *m.Evidenced}
	}

	if m.RiskNoImplementationAccepted != nil || m.RiskNoEvidenceAccepted != nil {
		ms.RiskAcceptance = &RiskAcceptance{
			RiskNoImplementationAccepted: m.RiskNoImplementationAccepted != nil && *m.RiskNoImplementationAccepted,
			RiskNoEvidenceAccepted:       m.RiskNoEvidenceAccepted != nil && *m.RiskNoEvidenceAccepted,
		}
	}

	return ms
}

// ToState converts the version to a State which can be written with CreateStateByObject or UpdateState.
// The ID is not set, set it to update an existing state.
func (v StateVersion) ToState(uniquenessKey, protectSurfaceID, locationID string) *State {
	content := append([]string{}, v.Content...)

	return &State{
		UniquenessKey:    uniquenessKey,
		Description:      v.Description,
		ProtectSurface:   protectSurfaceID,
		ContentType:      v.ContentType,
		Location:         locationID,
		ExistsOnAssetIDs: v.ExistsOnAssetIDs,
		Maintainer:       v.Maintainer,
		Content:          &content,
	}
}

// ToLocation converts the version to a Location which can be written with CreateLocationByObject or UpdateLocation.
// The ID is not set, set it to update an existing location.
func (v LocationVersion) ToLocation(uniquenessKey string) *Location {
	return &Location{
		UniquenessKey: uniquenessKey,
		Name:          v.Name,
		Coords:        v.Coords,
	}
}

// flowsByID re-keys flows from protect surface uniqueness key to ID
func flowsByID(flows ObjectMap[Flow], psIDs map[string]string) (map[string]Flow, error) {
	result := make(map[string]Flow, len(flows))

	for uk, flow := range flows {
		id, ok := psIDs[uk]
		if !ok {
			return nil, fmt.Errorf("No protect surface ID found for uniqueness key %s", uk)
		}
		result[id] = flow
	}

	return result, nil
}
