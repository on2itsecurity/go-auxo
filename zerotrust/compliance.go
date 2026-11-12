package zerotrust

import (
	"context"
	"net/url"

	"github.com/on2itsecurity/go-auxo/v2/utils"
)

// Compliance holds the compliance status of a protect surface for a single framework version
type Compliance struct {
	Name           string          `json:"name"`            //Framework version, f.e. `PCI DSS 4.0.1`
	Family         string          `json:"family"`          //Framework family the version belongs to, f.e. `PCI DSS`
	CompliantCount int             `json:"compliant_count"` //Number of required measures that are ok
	TotalCount     int             `json:"total_count"`     //Number of measures the framework version requires
	Verdict        bool            `json:"verdict"`         //True when every required measure is ok
	Gaps           []ComplianceGap `json:"gaps"`            //Required measures that are assigned but not ok, risk-accepted ones included
	NotAssigned    []string        `json:"not_assigned"`    //Names of required measures that are not assigned or whose assignment is unknown
}

// ComplianceGap is a required measure that is assigned but not ok
type ComplianceGap struct {
	Name         string `json:"name"`          //Name of the measure, f.e. `data-dlp`
	FitnessState string `json:"fitness_state"` //Either: `no_evidence`, `no_evidence_risk_accepted`, `not_implemented` or `not_implemented_risk_accepted`
}

// --- Functions ---

// GetProtectSurfaceCompliance will get the compliance status of a protectsurface per framework version.
// The framework versions apply through the compliance tags of the protectsurface,
// it returns an empty slice when none of its compliance tags match a framework.
func (zt *ZeroTrust) GetProtectSurfaceCompliance(ctx context.Context, protectSurfaceID string) ([]*Compliance, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	call := "get-protectsurface-compliance?id=" + url.QueryEscape(protectSurfaceID)
	method := "GET"

	result, err := zt.apiClient.ApiCall(ctx, zt.apiEndpoint+call, method, "")

	if err != nil {
		return nil, err
	}

	compliance, err := utils.UnwrapItems[Compliance](result)

	if err != nil {
		return nil, err
	}

	if compliance == nil {
		compliance = make([]*Compliance, 0)
	}

	return compliance, nil
}
