package compactor

import (
	"fmt"
	"strings"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// MakeDisjointRelabelConfigForGroups returns a list of RelabelConfig for each group that ensures the action "keep" is applied to all label values within the group.
// It returns an error if a label value is duplicated across multiple groups.
// This function helps generate disjoint relabel configurations, ensuring each compactor only processes blocks from its designated group.
// For instance, in a multitenant setup using a shared object storage bucket where different tenants require varying retention periods,
// this function can craft the necessary relabel configurations for each compactor.
// You must use a label that is propagated to the blocks, like an external label.
func MakeDisjointRelabelConfigForGroups(labelName string, groups map[string][]string, maxCfgSize int) (map[string][]monv1.RelabelConfig, error) {
	ret := make(map[string][]monv1.RelabelConfig, len(groups))

	if maxCfgSize <= 0 {
		maxCfgSize = 1
	}

	if err := checkLabelValuesUnicity(groups); err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return ret, nil
	}

	addRelabelConfig := func(group string, labelValues []string) {
		if len(labelValues) == 0 {
			return
		}

		var valuesRegex strings.Builder
		valuesRegex.WriteString("'" + labelValues[0] + "'")

		for _, value := range labelValues[1:] {
			valuesRegex.WriteString("|'" + value + "'")
		}

		ret[group] = append(ret[group], monv1.RelabelConfig{
			Action:       "keep",
			SourceLabels: []monv1.LabelName{monv1.LabelName(labelName)},
			Regex:        valuesRegex.String(),
		})
	}

	for group, labelValues := range groups {
		for i := 0; i < len(labelValues); i += maxCfgSize {
			end := i + maxCfgSize
			if end > len(labelValues) {
				end = len(labelValues)
			}

			addRelabelConfig(group, labelValues[i:end])
		}
	}

	return ret, nil
}

func checkLabelValuesUnicity(groups map[string][]string) error {
	seen := make(map[string]struct{}, len(groups))

	for group, labelValues := range groups {
		for _, value := range labelValues {
			if _, ok := seen[value]; ok {
				return fmt.Errorf("label value %q is duplicated in group %q", value, group)
			}

			seen[value] = struct{}{}
		}
	}

	return nil
}
