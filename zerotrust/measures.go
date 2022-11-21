package zerotrust

import "github.com/on2itsecurity/go-auxo/utils"

type MeasureGroups struct {
	Groups []MeasureGroup `json:"groups"` //All measures are categorized in groups
}

type MeasureGroup struct {
	Caption  string    `json:"group_caption"` //The caption of the group
	Label    string    `json:"group_label"`   //The label of the group
	Name     string    `json:"group_name"`    //The name of the group
	Icon     string    `json:"icon"`          //The icon of the group
	Measures []Measure `json:"measures"`      //The measures of the group
}

type Measure struct {
	Caption     string              `json:"caption"`     //The caption of the measure
	Explanation string              `json:"explanation"` //The explanation of the measure
	Mappings    map[string][]string `json:"mappings"`    //The mappings of the measure (f.e. with Mitre ATT&CK)
	Name        string              `json:"name"`        //The name (=id) of the measure
}

// GetMeasures will get all measures of the relation (based on used API Token)
// It returns an array with all the Measure objects.
func (zt *ZeroTrust) GetMeasures() (*MeasureGroups, error) {
	call := "get-all-measures"
	method := "GET"

	result, err := zt.apiClient.ApiCall(zt.apiEndpoint+call, method, "")

	if err != nil {
		return nil, err
	}

	groups, err := utils.UnwrapItems[MeasureGroups](result)

	if err != nil {
		return nil, err
	}

	return groups[0], nil
}
