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
	ChangeCreate ChangeType = "create" // Entity was created
	ChangeUpdate ChangeType = "update" // Content was changed
	ChangeDelete ChangeType = "delete" // Entity was deleted
)

// Change describes a single recorded change of an entity
type Change struct {
	Timestamp     int64      // Unix time of the change
	EntityType    EntityType // Type of the changed entity
	UniquenessKey string     // Uniqueness key of the changed entity
	Name          string     // Name (protect surface, location) or description (state) of the entity
	Type          ChangeType // Create, update or delete
	// Changed fields as json names, map entries as "field.key" (f.e. "measures.data-dlp").
	// nil for a create or when Before is unknown, empty when no visible field changed.
	Fields []string
	// Content before and after the change: *ProtectSurfaceVersion, *StateVersion or *LocationVersion.
	// Before is nil for a create and for the first recorded change (content before it was not recorded), After is nil for a delete.
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
// Protect surfaces created after t are left out, together with their states.
//
// Limitations of the recorded history:
//   - The creation date is only known for protect surfaces, locations and states (of protect surfaces created before t) created after t are included.
//   - Protect surfaces with CreateDate CreationRecordedFrom have an unknown creation time, they are left out before and included after that moment.
//   - The content before the first recorded change is unknown, those entities are returned with Exact set to false
//     and the content of their oldest recorded version.
func (h *History) AsOf(t time.Time) (*PointInTime, error) {
	if t.Before(h.From) {
		return nil, fmt.Errorf("Cannot reconstruct %s, history starts at %s", t.Format(time.RFC3339), h.From.Format(time.RFC3339))
	}

	ts := t.Unix()
	pit := &PointInTime{At: t}

	createdLater := make(map[string]bool) //Protect surfaces created after t, by uniqueness key
	for _, ps := range h.ProtectSurfaces {
		if ps.CreateDate > ts {
			createdLater[ps.UniquenessKey] = true
		}
	}

	for _, ps := range h.ProtectSurfaces {
		if v, exact, ok := versionAt(ps.HistoryMeta, ps.ProtectSurfaceVersion, ps.Versions, ts); ok {
			pit.ProtectSurfaces = append(pit.ProtectSurfaces, &ProtectSurfaceAt{UniquenessKey: ps.UniquenessKey, ProtectSurfaceVersion: v, Exact: exact})
		}
	}

	for _, s := range h.States {
		if createdLater[s.ProtectSurfaceUniquenessKey] {
			continue //A state cannot exist before its protect surface
		}
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

// versionAt returns the content of an entity at ts, whether it is exact and whether the entity existed (created and not deleted).
// A version holds the content after the change at its timestamp, so the content at ts is the newest version at or before ts.
func versionAt[V versioned](meta HistoryMeta, current V, versions []V, ts int64) (V, bool, bool) {
	var none V
	if meta.CreateDate > ts {
		return none, false, false
	}
	if deletedAt, deleted := deletion(meta, versions); deleted && deletedAt <= ts {
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
	//Versions are newest first, reverse before the (stable) sort so versions with the same timestamp stay in order
	asc := make([]*V, len(versions))
	for i := range versions {
		asc[len(versions)-1-i] = &versions[i]
	}
	sort.SliceStable(asc, func(i, j int) bool { return versionTimestamp(asc[i]) < versionTimestamp(asc[j]) })

	deletedAt, deleted := deletion(meta, versions)

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

		if deleted && i == len(asc)-1 && c.Timestamp == deletedAt {
			//The deletion records the content at the time of deletion
			c.Type = ChangeDelete
			c.Before = v
			c.After = nil
		} else if i > 0 {
			c.Before = asc[i-1]
			c.Fields = diffFields(*asc[i-1], *v)
		} else if meta.CreateDate != 0 && c.Timestamp == meta.CreateDate {
			//Creation records a version at the creation date
			c.Type = ChangeCreate
		}

		changes = append(changes, c)
	}

	return changes
}

// deletion returns when the entity was deleted and whether it is deleted.
// When the deletion time is unknown (deleted before the history was recorded), the newest version is taken as the deletion.
func deletion[V versioned](meta HistoryMeta, versions []V) (int64, bool) {
	if !meta.Deleted || meta.DeletedAt != 0 {
		return meta.DeletedAt, meta.Deleted
	}

	var newest int64
	for i := range versions {
		if ts := versionTimestamp(&versions[i]); ts > newest {
			newest = ts
		}
	}

	return newest, true
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
		measures[name] = m
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
