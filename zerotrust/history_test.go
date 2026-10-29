package zerotrust

import (
	"reflect"
	"testing"
	"time"
)

// historyFixture follows the shape returned by the API (Tailspin, October 2026):
//   - location L1 created (1,1) at 90, changed to (2,2) at 100, renamed and (3,3) at 200
//   - location L2 changed at 100, deleted at 300
//   - protect surface PS1 with empty customer labels ([]), one measure implemented at 150
//   - state S1 without any change
const historyFixture = `{"items":[{
 "locations":{"items":[
  {"relation_id":"1","uniqueness_key":"L1","deleted":false,"deleted_at":null,"timestamp":null,
   "name":"loc-renamed","coords":{"lat":3,"long":3},
   "snapshots":[
    {"timestamp":200,"name":"loc-renamed","coords":{"lat":3,"long":3}},
    {"timestamp":100,"name":"loc","coords":{"lat":2,"long":2}}]},
  {"relation_id":"1","uniqueness_key":"L2","deleted":true,"deleted_at":300,"timestamp":null,
   "name":"gone","coords":{"lat":5,"long":5},
   "snapshots":[
    {"timestamp":300,"name":"gone","coords":{"lat":5,"long":5}},
    {"timestamp":100,"name":"gone","coords":{"lat":5,"long":5}}]}]},
 "protectsurfaces":{"items":[
  {"relation_id":"1","uniqueness_key":"PS1","deleted":false,"deleted_at":null,"timestamp":null,
   "name":"P_EHR","description":"","main_contact_person_id":null,"security_contact_person_id":"42",
   "in_control_boundary":true,"in_zero_trust_focus":true,"relevance":60,"confidentiality":5,"integrity":5,"availability":5,
   "data_tags":["PHI"],"compliance_tags":[],"soc_tags":[],"customer_labels":[],
   "flows_to_other_ps":{"PS2":{"allow":true}},"flows_from_other_ps":[],"flows_to_outside":{"allow":false},"flows_from_outside":[],
   "measures":{"data-dlp":{"assignment":{"assigned":true,"last_determined_by_person_id":"a@example.com","last_determined_timestamp":120},"implementation":{"implemented":true,"comment":"done in Q3","last_determined_by_person_id":"b@example.com","last_determined_timestamp":150},"risk_acceptance":{"risk_no_implementation_accepted":false,"risk_no_evidence_accepted":true,"risk_accepted_comment":"no evidence possible","last_determined_by_person_id":"c@example.com","last_determined_timestamp":150}}},
   "maturity":{"step_1":1,"step_2":1,"step_3":1,"step_4":1,"step_5":1},
   "snapshots":[
    {"timestamp":150,"name":"P_EHR","description":"","main_contact_person_id":null,"security_contact_person_id":"42",
     "in_control_boundary":true,"in_zero_trust_focus":true,"relevance":60,"confidentiality":5,"integrity":5,"availability":5,
     "data_tags":["PHI"],"compliance_tags":[],"soc_tags":[],"customer_labels":[],
     "flows_to_other_ps":{"PS2":{"allow":true}},"flows_from_other_ps":[],"flows_to_outside":{"allow":false},"flows_from_outside":[],
     "measures":{"data-dlp":{"assignment":{"assigned":true,"last_determined_by_person_id":"a@example.com","last_determined_timestamp":120},"implementation":{"implemented":true,"comment":"done in Q3","last_determined_by_person_id":"b@example.com","last_determined_timestamp":150},"risk_acceptance":{"risk_no_implementation_accepted":false,"risk_no_evidence_accepted":true,"risk_accepted_comment":"no evidence possible","last_determined_by_person_id":"c@example.com","last_determined_timestamp":150}}},
     "maturity":{"step_1":1,"step_2":1,"step_3":1,"step_4":1,"step_5":1}},
    {"timestamp":120,"name":"P_EHR","description":"","main_contact_person_id":null,"security_contact_person_id":"42",
     "in_control_boundary":true,"in_zero_trust_focus":true,"relevance":60,"confidentiality":5,"integrity":5,"availability":5,
     "data_tags":["PHI"],"compliance_tags":[],"soc_tags":[],"customer_labels":[],
     "flows_to_other_ps":{"PS2":{"allow":true}},"flows_from_other_ps":[],"flows_to_outside":{"allow":false},"flows_from_outside":[],
     "measures":{"data-dlp":{"assignment":{"assigned":true,"last_determined_by_person_id":"a@example.com","last_determined_timestamp":120}}},
     "maturity":{"step_1":1,"step_2":1,"step_3":1,"step_4":1,"step_5":1}}]}]},
 "states":{"items":[
  {"relation_id":"1","uniqueness_key":"S1","deleted":false,"deleted_at":null,"timestamp":null,
   "protectsurface_uniqueness_key":"PS1","description":"ehr-ipv4","content_type":"ipv4","location_uniqueness_key":"L1",
   "exists_on_asset_ids":[],"maintainer":"portal_manual","content":["10.0.0.0/24"],"snapshots":[]}]}
}]}`

// loadPages merges the given history pages like GetHistory does
func loadPages(t *testing.T, pages ...string) *History {
	t.Helper()
	m := newHistoryMerger(time.Unix(0, 0))
	for _, page := range pages {
		p, err := parseHistoryPage([]byte(page))
		if err != nil {
			t.Fatalf("parseHistoryPage: %v", err)
		}
		m.add(p)
	}
	return m.history
}

func loadFixture(t *testing.T) *History {
	t.Helper()
	return loadPages(t, historyFixture)
}

func TestParseHistory(t *testing.T) {
	h := loadFixture(t)

	if len(h.ProtectSurfaces) != 1 || len(h.States) != 1 || len(h.Locations) != 2 {
		t.Fatalf("unexpected bucket sizes %d/%d/%d", len(h.ProtectSurfaces), len(h.States), len(h.Locations))
	}

	ps := h.ProtectSurfaces[0]
	if ps.UniquenessKey != "PS1" || ps.Timestamp != 0 || len(ps.Versions) != 2 {
		t.Errorf("unexpected protect surface %+v", ps.HistoryMeta)
	}
	if ps.CustomerLabels == nil || len(ps.CustomerLabels) != 0 {
		t.Errorf("customer_labels [] should be an empty map, got %#v", ps.CustomerLabels)
	}
	if ps.FlowsFromOutside.Allow != nil || ps.FlowsToOutside.Allow == nil || *ps.FlowsToOutside.Allow {
		t.Errorf("unexpected outside flows %+v %+v", ps.FlowsFromOutside, ps.FlowsToOutside)
	}

	l2 := h.Locations[1]
	if !l2.Deleted || l2.DeletedAt != 300 {
		t.Errorf("unexpected deletion %+v", l2.HistoryMeta)
	}
}

func TestAsOf(t *testing.T) {
	h := loadFixture(t)

	tests := []struct {
		at        int64
		l1Name    string
		l1Lat     float64
		l1Exact   bool
		locations int
	}{
		{at: 95, l1Name: "loc", l1Lat: 2, l1Exact: false, locations: 2}, //Before first recorded change, creation content (1,1) unknown
		{at: 100, l1Name: "loc", l1Lat: 2, l1Exact: true, locations: 2},
		{at: 199, l1Name: "loc", l1Lat: 2, l1Exact: true, locations: 2},
		{at: 200, l1Name: "loc-renamed", l1Lat: 3, l1Exact: true, locations: 2},
		{at: 300, l1Name: "loc-renamed", l1Lat: 3, l1Exact: true, locations: 1}, //L2 deleted
	}

	for _, tc := range tests {
		pit, err := h.AsOf(time.Unix(tc.at, 0))
		if err != nil {
			t.Fatalf("AsOf(%d): %v", tc.at, err)
		}
		if len(pit.Locations) != tc.locations {
			t.Errorf("AsOf(%d): %d locations, want %d", tc.at, len(pit.Locations), tc.locations)
		}
		l1 := pit.Locations[0]
		if l1.UniquenessKey != "L1" || l1.Name != tc.l1Name || l1.Coords.Latitude != tc.l1Lat || l1.Exact != tc.l1Exact {
			t.Errorf("AsOf(%d): L1 = %s %v exact=%v, want %s %v exact=%v", tc.at, l1.Name, l1.Coords.Latitude, l1.Exact, tc.l1Name, tc.l1Lat, tc.l1Exact)
		}
		if len(pit.States) != 1 || !pit.States[0].Exact || pit.States[0].ProtectSurfaceUniquenessKey != "PS1" {
			t.Errorf("AsOf(%d): unchanged state should be exact current content", tc.at)
		}
	}

	pit, _ := h.AsOf(time.Unix(130, 0))
	if m := pit.ProtectSurfaces[0].Measures["data-dlp"]; m.Implementation != nil {
		t.Errorf("AsOf(130): data-dlp should not be implemented yet")
	}

	h.From = time.Unix(100, 0)
	if _, err := h.AsOf(time.Unix(50, 0)); err == nil {
		t.Errorf("AsOf before From should fail")
	}
}

func TestChanges(t *testing.T) {
	h := loadFixture(t)
	changes := h.Changes()

	type summary struct {
		ts     int64
		uk     string
		typ    ChangeType
		fields []string
		before bool
	}
	got := make([]summary, 0, len(changes))
	for _, c := range changes {
		got = append(got, summary{c.Timestamp, c.UniquenessKey, c.Type, c.Fields, c.Before != nil})
	}

	want := []summary{
		{100, "L1", ChangeUpdate, nil, false},
		{100, "L2", ChangeUpdate, nil, false},
		{120, "PS1", ChangeUpdate, nil, false},
		{150, "PS1", ChangeUpdate, []string{"measures.data-dlp"}, true},
		{200, "L1", ChangeUpdate, []string{"name", "coords"}, true},
		{300, "L2", ChangeDelete, nil, true},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Changes:\n got  %+v\n want %+v", got, want)
	}

	if after, ok := changes[4].After.(*LocationVersion); !ok || after.Name != "loc-renamed" {
		t.Errorf("After should be the new location content, got %#v", changes[4].After)
	}
	if changes[5].After != nil || changes[5].Before.(*LocationVersion).Name != "gone" {
		t.Errorf("Delete should carry the deleted content as Before")
	}
}

func TestChangesSameTimestamp(t *testing.T) {
	meta := HistoryMeta{UniquenessKey: "L3"}
	versions := []LocationVersion{ //Newest first, as returned by the API
		{Timestamp: 400, Name: "third"},
		{Timestamp: 400, Name: "second"},
		{Timestamp: 300, Name: "first"},
	}

	changes := entityChanges(EntityLocation, meta, versions, func(v *LocationVersion) string { return v.Name })
	names := []string{changes[0].Name, changes[1].Name, changes[2].Name}
	if !reflect.DeepEqual(names, []string{"first", "second", "third"}) {
		t.Errorf("changes with the same timestamp out of order: %v", names)
	}
}

func TestToProtectSurface(t *testing.T) {
	h := loadFixture(t)
	v := h.ProtectSurfaces[0].ProtectSurfaceVersion

	if _, err := v.ToProtectSurface("PS1", map[string]string{}); err == nil {
		t.Errorf("unknown flow uniqueness key should fail")
	}

	ps, err := v.ToProtectSurface("PS1", map[string]string{"PS2": "ID2"})
	if err != nil {
		t.Fatalf("ToProtectSurface: %v", err)
	}
	if ps.UniquenessKey != "PS1" || ps.ID != "" || ps.SecurityContactPersonID != "42" {
		t.Errorf("unexpected identity %+v", ps)
	}
	if f, ok := ps.FlowsToOtherPS["ID2"]; !ok || !*f.Allow {
		t.Errorf("flows should be keyed by ID, got %+v", ps.FlowsToOtherPS)
	}

	m := ps.Measures["data-dlp"]
	if !m.Assignment.Assigned || m.Implementation == nil || !m.Implementation.Implemented || m.Implementation.Comment != "done in Q3" || m.Evidence != nil {
		t.Errorf("unexpected measure conversion %+v", m)
	}
	if m.Implementation.LastDeterminedByPersonID != "b@example.com" || m.RiskAcceptance.RiskAcceptedComment != "no evidence possible" {
		t.Errorf("measures should keep who/when and the risk accepted comment, got %+v %+v", m.Implementation, m.RiskAcceptance)
	}
}

func TestToState(t *testing.T) {
	h := loadFixture(t)
	s := h.States[0].ToState("S1", "ID1", "LID1")

	if s.UniquenessKey != "S1" || s.ProtectSurface != "ID1" || s.Location != "LID1" || (*s.Content)[0] != "10.0.0.0/24" {
		t.Errorf("unexpected state %+v", s)
	}
}

// Two pages: every page repeats the current content, deleted_at is only set on the page holding the deletion
const historyPage1 = `{"items":[{"has_more":true,
 "protectsurfaces":{"items":[]},"states":{"items":[]},
 "locations":{"items":[
  {"relation_id":"1","uniqueness_key":"L1","deleted":false,"deleted_at":null,"timestamp":null,"name":"c","coords":{"lat":3,"long":3},
   "snapshots":[{"timestamp":300,"name":"c","coords":{"lat":3,"long":3}}]},
  {"relation_id":"1","uniqueness_key":"L2","deleted":true,"deleted_at":250,"timestamp":null,"name":"x","coords":{"lat":0,"long":0},
   "snapshots":[{"timestamp":250,"name":"x","coords":{"lat":0,"long":0}}]}]}}]}`

const historyPage2 = `{"items":[{"has_more":false,
 "protectsurfaces":{"items":[]},"states":{"items":[]},
 "locations":{"items":[
  {"relation_id":"1","uniqueness_key":"L1","deleted":false,"deleted_at":null,"timestamp":null,"name":"c","coords":{"lat":3,"long":3},
   "snapshots":[{"timestamp":200,"name":"b","coords":{"lat":2,"long":2}},{"timestamp":100,"name":"a","coords":{"lat":1,"long":1}}]},
  {"relation_id":"1","uniqueness_key":"L2","deleted":true,"deleted_at":null,"timestamp":null,"name":"x","coords":{"lat":0,"long":0},
   "snapshots":[{"timestamp":150,"name":"x","coords":{"lat":0,"long":0}}]},
  {"relation_id":"1","uniqueness_key":"L3","deleted":true,"deleted_at":null,"timestamp":null,"name":"old","coords":{"lat":0,"long":0},
   "snapshots":[{"timestamp":120,"name":"old","coords":{"lat":0,"long":0}}]}]}}]}`

func TestMergePages(t *testing.T) {
	h := loadPages(t, historyPage1, historyPage2)

	if len(h.Locations) != 3 {
		t.Fatalf("expected 3 locations, got %d", len(h.Locations))
	}

	l1 := h.Locations[0]
	ts := []int64{}
	for _, v := range l1.Versions {
		ts = append(ts, v.Timestamp)
	}
	if !reflect.DeepEqual(ts, []int64{300, 200, 100}) {
		t.Errorf("versions should be merged newest first, got %v", ts)
	}

	l2 := h.Locations[1]
	if !l2.Deleted || l2.DeletedAt != 250 || len(l2.Versions) != 2 {
		t.Errorf("deletion should survive the merge, got %+v with %d versions", l2.HistoryMeta, len(l2.Versions))
	}

	pit, _ := h.AsOf(time.Unix(240, 0))
	if len(pit.Locations) != 2 || pit.Locations[1].UniquenessKey != "L2" {
		t.Errorf("AsOf(240) should hold L1 and L2 (L3 deleted at its newest version, 120), got %d locations", len(pit.Locations))
	}

	changes := h.Changes()
	last := changes[len(changes)-1]
	if last.UniquenessKey != "L1" || last.Timestamp != 300 {
		t.Errorf("unexpected last change %+v", last)
	}
	for _, c := range changes {
		if c.UniquenessKey == "L3" && c.Type != ChangeDelete {
			t.Errorf("L3 without deleted_at should be deleted at its newest version, got %s", c.Type)
		}
	}
}

// Creation dates (live data, October 2026): the creation records a version at create_date,
// a protect surface from before the history keeps its create_date on every page.
//   - PS_OLD created at 50 (before From), first recorded change at 200
//   - PS_NEW created at 120, its state S_NEW never changed
//   - PS_OLD has state S_OLD, never changed
const historyCreatePage1 = `{"items":[{"has_more":true,
 "protectsurfaces":{"items":[
  {"relation_id":"1","uniqueness_key":"PS_OLD","create_date":50,"deleted":false,"deleted_at":null,"timestamp":null,"name":"old-renamed",
   "snapshots":[{"timestamp":200,"create_date":50,"name":"old-renamed"}]},
  {"relation_id":"1","uniqueness_key":"PS_NEW","create_date":120,"deleted":false,"deleted_at":null,"timestamp":null,"name":"new",
   "snapshots":[]}]},
 "states":{"items":[
  {"relation_id":"1","uniqueness_key":"S_OLD","protectsurface_uniqueness_key":"PS_OLD","deleted":false,"deleted_at":null,"timestamp":null,"description":"old","snapshots":[]},
  {"relation_id":"1","uniqueness_key":"S_NEW","protectsurface_uniqueness_key":"PS_NEW","deleted":false,"deleted_at":null,"timestamp":null,"description":"new","snapshots":[]}]},
 "locations":{"items":[]}}]}`

const historyCreatePage2 = `{"items":[{"has_more":false,
 "protectsurfaces":{"items":[
  {"relation_id":"1","uniqueness_key":"PS_OLD","create_date":50,"deleted":false,"deleted_at":null,"timestamp":null,"name":"old-renamed","snapshots":[]},
  {"relation_id":"1","uniqueness_key":"PS_NEW","create_date":120,"deleted":false,"deleted_at":null,"timestamp":null,"name":"new",
   "snapshots":[{"timestamp":120,"create_date":120,"name":"new"}]}]},
 "states":{"items":[]},"locations":{"items":[]}}]}`

func TestCreateDate(t *testing.T) {
	h := loadPages(t, historyCreatePage1, historyCreatePage2)
	h.From = time.Unix(100, 0)

	if h.ProtectSurfaces[0].CreateDate != 50 || h.ProtectSurfaces[1].CreateDate != 120 {
		t.Fatalf("unexpected create dates %d %d", h.ProtectSurfaces[0].CreateDate, h.ProtectSurfaces[1].CreateDate)
	}

	tests := []struct {
		at       int64
		pss      []string
		states   []string
		oldExact bool
	}{
		{at: 110, pss: []string{"PS_OLD"}, states: []string{"S_OLD"}, oldExact: false}, //PS_NEW and its state not created yet
		{at: 120, pss: []string{"PS_OLD", "PS_NEW"}, states: []string{"S_OLD", "S_NEW"}, oldExact: false},
		{at: 200, pss: []string{"PS_OLD", "PS_NEW"}, states: []string{"S_OLD", "S_NEW"}, oldExact: true},
	}

	for _, tc := range tests {
		pit, err := h.AsOf(time.Unix(tc.at, 0))
		if err != nil {
			t.Fatalf("AsOf(%d): %v", tc.at, err)
		}
		pss, states := []string{}, []string{}
		for _, ps := range pit.ProtectSurfaces {
			pss = append(pss, ps.UniquenessKey)
		}
		for _, s := range pit.States {
			states = append(states, s.UniquenessKey)
		}
		if !reflect.DeepEqual(pss, tc.pss) || !reflect.DeepEqual(states, tc.states) {
			t.Errorf("AsOf(%d): protect surfaces %v states %v, want %v %v", tc.at, pss, states, tc.pss, tc.states)
		}
		if pit.ProtectSurfaces[0].Exact != tc.oldExact {
			t.Errorf("AsOf(%d): PS_OLD exact=%v, want %v", tc.at, pit.ProtectSurfaces[0].Exact, tc.oldExact)
		}
		if len(pit.ProtectSurfaces) > 1 && !pit.ProtectSurfaces[1].Exact {
			t.Errorf("AsOf(%d): PS_NEW should be exact from its creation", tc.at)
		}
	}

	types := map[string]ChangeType{}
	for _, c := range h.Changes() {
		types[c.UniquenessKey] = c.Type
		if c.Before != nil || c.Fields != nil {
			t.Errorf("%s: first change should have no Before and no Fields", c.UniquenessKey)
		}
	}
	if types["PS_NEW"] != ChangeCreate || types["PS_OLD"] != ChangeUpdate {
		t.Errorf("PS_NEW should be a create and PS_OLD an update (created before From), got %v", types)
	}
}
