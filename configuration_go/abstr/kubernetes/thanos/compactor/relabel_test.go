package compactor_test

import (
	"testing"

	"github.com/observatorium/observatorium/configuration_go/abstr/kubernetes/thanos/compactor"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestMakeDisjointRelabelConfigForGroups(t *testing.T) {
	labelName := monv1.LabelName("mylabel")

	makeRelabelCfg := func(regex string) monv1.RelabelConfig {
		return monv1.RelabelConfig{
			Action:       "keep",
			SourceLabels: []monv1.LabelName{labelName},
			Regex:        regex,
		}
	}

	testCases := map[string]struct {
		groups     map[string][]string
		maxCfgSize int
		expect     map[string][]monv1.RelabelConfig
		expectErr  bool
	}{
		"empty": {
			groups: map[string][]string{},
			expect: map[string][]monv1.RelabelConfig{},
		},
		"one group": {
			groups: map[string][]string{
				"group1": {"value1"},
			},
			expect: map[string][]monv1.RelabelConfig{
				"group1": {makeRelabelCfg("'value1'")},
			},
		},
		"two groups": {
			groups: map[string][]string{
				"group1": {"value1", "value2"},
			},
			expect: map[string][]monv1.RelabelConfig{
				"group1": {
					makeRelabelCfg("'value1'"),
					makeRelabelCfg("'value2'"),
				},
			},
		},
		"group split": {
			groups: map[string][]string{
				"group1": {"value1", "value2", "value3", "value4", "value5"},
			},
			maxCfgSize: 2,
			expect: map[string][]monv1.RelabelConfig{
				"group1": {
					makeRelabelCfg("'value1'|'value2'"),
					makeRelabelCfg("'value3'|'value4'"),
					makeRelabelCfg("'value5'"),
				},
			},
		},

		"duplicated label value": {
			groups: map[string][]string{
				"group1": {"value1", "value1"},
			},
			expectErr: true,
		},
		"duplicated label value in different groups": {
			groups: map[string][]string{
				"group1": {"value1"},
				"group2": {"value1"},
			},
			expectErr: true,
		},
		"multi groups": {
			groups: map[string][]string{
				"group1": {"value1", "value2", "value3", "value4"},
				"group2": {"value11", "value12", "value13"},
			},
			maxCfgSize: 2,
			expect: map[string][]monv1.RelabelConfig{
				"group1": {
					makeRelabelCfg("'value1'|'value2'"),
					makeRelabelCfg("'value3'|'value4'"),
				},
				"group2": {
					makeRelabelCfg("'value11'|'value12'"),
					makeRelabelCfg("'value13'"),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := compactor.MakeDisjointRelabelConfigForGroups(string(labelName), tc.groups, tc.maxCfgSize)
			if err != nil {
				if !tc.expectErr {
					t.Fatalf("unexpected error: %v", err)
				}

				return
			}

			if len(got) != len(tc.expect) {
				t.Fatalf("expected %d relabel configs, got %d", len(tc.expect), len(got))
			}

			for group, expect := range tc.expect {
				got, ok := got[group]
				if !ok {
					t.Fatalf("expected relabel config for group %q", group)
				}

				if len(got) != len(expect) {
					t.Fatalf("expected %d relabel configs for group %q, got %d", len(expect), group, len(got))
				}

				for i := range expect {
					if got[i].Regex != expect[i].Regex {
						t.Fatalf("expected regex %q, got %q", expect[i].Regex, got[i].Regex)
					}

					if got[i].Action != expect[i].Action {
						t.Fatalf("expected action %q, got %q", expect[i].Action, got[i].Action)
					}

					if len(got[i].SourceLabels) != len(expect[i].SourceLabels) {
						t.Fatalf("expected %d source labels, got %d", len(expect[i].SourceLabels), len(got[i].SourceLabels))
					}

					for j := range expect[i].SourceLabels {
						if got[i].SourceLabels[j] != expect[i].SourceLabels[j] {
							t.Fatalf("expected source label %q, got %q", expect[i].SourceLabels[j], got[i].SourceLabels[j])
						}
					}
				}
			}
		})
	}
}
