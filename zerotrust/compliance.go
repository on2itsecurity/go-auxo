package zerotrust

import (
	"context"
	"fmt"
	"net/url"

	"github.com/on2itsecurity/go-auxo/v2/utils"
)

// Compliance holds the compliance status of a protect surface for a single framework version
type Compliance struct {
	Name           string          `json:"name"`            //Framework version, f.e. `PCI DSS 4.0.1`
	Family         string          `json:"family"`          //Framework family the version belongs to, f.e. `PCI DSS`
	CompliantCount int             `json:"compliant_count"` //Number of required measures that are ok
	TotalCount     int             `json:"total_count"`     //Number of measures the framework version requires
	Verdict        bool            `json:"verdict"`         //True when every required measure is ok, a risk-accepted measure counts as not ok
	Gaps           []ComplianceGap `json:"gaps"`            //Required measures that are assigned but not ok, risk-accepted ones included
	NotAssigned    []string        `json:"not_assigned"`    //Names of required measures that are not assigned or whose assignment is unknown
}

// ComplianceGap is a required measure that is assigned but not ok
type ComplianceGap struct {
	Name         string       `json:"name"`          //Name of the measure, f.e. `data-dlp`
	FitnessState FitnessState `json:"fitness_state"` //Fitness state of the measure, see the FitnessState constants
}

// FitnessState is the fitness state of a measure that is assigned but not ok
type FitnessState string

const (
	FitnessNoEvidence                 FitnessState = "no_evidence"                   //Implemented, but not evidenced
	FitnessNoEvidenceRiskAccepted     FitnessState = "no_evidence_risk_accepted"     //Implemented, not evidenced and the risk is accepted
	FitnessNotImplemented             FitnessState = "not_implemented"               //Not implemented
	FitnessNotImplementedRiskAccepted FitnessState = "not_implemented_risk_accepted" //Not implemented and the risk is accepted
)

// --- Functions ---

// GetProtectSurfaceCompliance will get the compliance status of a protectsurface per framework version.
// The framework versions apply through the compliance tags of the protectsurface,
// it returns an empty slice when none of its compliance tags match a framework.
// Pass a nil context to use the default timeout of the API client.
func (zt *ZeroTrust) GetProtectSurfaceCompliance(ctx context.Context, protectSurfaceID string) ([]*Compliance, error) {
	if protectSurfaceID == "" {
		return nil, fmt.Errorf("No protect surface ID given")
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
