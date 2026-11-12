package zerotrust

import (
	"testing"

	"github.com/on2itsecurity/go-auxo/v2/utils"
)

const complianceFixture = `{"items":[
 {"name":"PCI DSS 4.0.1","family":"PCI DSS","compliant_count":12,"total_count":20,"verdict":false,
  "gaps":[{"name":"data-dlp","fitness_state":"not_implemented"},{"name":"data-backup","fitness_state":"no_evidence_risk_accepted"}],
  "not_assigned":["mfa"]},
 {"name":"ISO 27001:2022","family":"ISO 27001","compliant_count":8,"total_count":8,"verdict":true,"gaps":[],"not_assigned":[]}]}`

func TestParseCompliance(t *testing.T) {
	compliance, err := utils.UnwrapItems[Compliance]([]byte(complianceFixture))
	if err != nil {
		t.Fatalf("UnwrapItems: %v", err)
	}

	if len(compliance) != 2 {
		t.Fatalf("expected 2 framework versions, got %d", len(compliance))
	}

	pci := compliance[0]
	if pci.Name != "PCI DSS 4.0.1" || pci.Family != "PCI DSS" || pci.CompliantCount != 12 || pci.TotalCount != 20 || pci.Verdict {
		t.Errorf("unexpected framework version %+v", pci)
	}
	if len(pci.Gaps) != 2 || pci.Gaps[1].Name != "data-backup" || pci.Gaps[1].FitnessState != "no_evidence_risk_accepted" {
		t.Errorf("unexpected gaps %+v", pci.Gaps)
	}
	if len(pci.NotAssigned) != 1 || pci.NotAssigned[0] != "mfa" {
		t.Errorf("unexpected not assigned %v", pci.NotAssigned)
	}

	if !compliance[1].Verdict || len(compliance[1].Gaps) != 0 {
		t.Errorf("unexpected compliant framework version %+v", compliance[1])
	}
}
