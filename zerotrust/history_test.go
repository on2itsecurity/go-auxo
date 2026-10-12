package zerotrust

import (
	"reflect"
	"testing"
	"time"
)

// historyFixture follows the shape returned by the API (Tailspin, October 2026):
//   - location L1 created (1,1) at 90, changed to (2,2) at 100, renamed and (3,3) at 200
//   - location L2 changed at 100, deleted at 300
//   - protect surface PS1 with empty customer labels ([]), one measure changed at 150
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
   "measures":{"data-dlp":{"assigned":true,"assigned_comment":null,"implemented":true,"implemented_comment":null,"evidenced":null,"evidenced_comment":null,"risk_no_implementation_accepted":null,"risk_no_evidence_accepted":false}},
   "maturity":{"step_1":1,"step_2":1,"step_3":1,"step_4":1,"step_5":1},
   "snapshots":[
    {"timestamp":150,"name":"P_EHR","description":"","main_contact_person_id":null,"security_contact_person_id":"42",
     "in_control_boundary":true,"in_zero_trust_focus":true,"relevance":60,"confidentiality":5,"integrity":5,"availability":5,
     "data_tags":["PHI"],"compliance_tags":[],"soc_tags":[],"customer_labels":[],
     "flows_to_other_ps":{"PS2":{"allow":true}},"flows_from_other_ps":[],"flows_to_outside":{"allow":false},"flows_from_outside":[],
     "measures":{"data-dlp":{"assigned":true,"assigned_comment":null,"implemented":true,"implemented_comment":null,"evidenced":null,"evidenced_comment":null,"risk_no_implementation_accepted":null,"risk_no_evidence_accepted":false}},
     "maturity":{"step_1":1,"step_2":1,"step_3":1,"step_4":1,"step_5":1}},
    {"timestamp":120,"name":"P_EHR","description":"","main_contact_person_id":null,"security_contact_person_id":"42",
     "in_control_boundary":true,"in_zero_trust_focus":true,"relevance":60,"confidentiality":5,"integrity":5,"availability":5,
     "data_tags":["PHI"],"compliance_tags":[],"soc_tags":[],"customer_labels":[],
     "flows_to_other_ps":{"PS2":{"allow":true}},"flows_from_other_ps":[],"flows_to_outside":{"allow":false},"flows_from_outside":[],
     "measures":{"data-dlp":{"assigned":true,"assigned_comment":null,"implemented":null,"implemented_comment":null,"evidenced":null,"evidenced_comment":null,"risk_no_implementation_accepted":null,"risk_no_evidence_accepted":null}},
     "maturity":{"step_1":1,"step_2":1,"step_3":1,"step_4":1,"step_5":1}}]}]},
 "states":{"items":[
  {"relation_id":"1","uniqueness_key":"S1","deleted":false,"deleted_at":null,"timestamp":null,
   "protectsurface_uniqueness_key":"PS1","description":"ehr-ipv4","content_type":"ipv4","location_uniqueness_key":"L1",
   "exists_on_asset_ids":[],"maintainer":"portal_manual","content":["10.0.0.0/24"],"snapshots":[]}]}
}]}`

func loadFixture(t *testing.T) *History {
	t.Helper()
	h, err := parseHistory([]byte(historyFixture), time.Unix(0, 0))
	if err != nil {
		t.Fatalf("parseHistory: %v", err)
	}
	return h
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
	if m := pit.ProtectSurfaces[0].Measures["data-dlp"]; m.Implemented != nil {
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
	if !m.Assignment.Assigned || m.Implementation == nil || !m.Implementation.Implemented || m.Evidence != nil || m.RiskAcceptance == nil || m.RiskAcceptance.RiskNoEvidenceAccepted {
		t.Errorf("unexpected measure conversion %+v", m)
	}
}

func TestToState(t *testing.T) {
	h := loadFixture(t)
	s := h.States[0].ToState("S1", "ID1", "LID1")

	if s.UniquenessKey != "S1" || s.ProtectSurface != "ID1" || s.Location != "LID1" || (*s.Content)[0] != "10.0.0.0/24" {
		t.Errorf("unexpected state %+v", s)
	}
}
