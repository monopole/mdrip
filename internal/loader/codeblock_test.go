package loader_test

import (
	. "github.com/monopole/mdrip/v2/internal/loader"
	"testing"
)

func Test_codeBlock_HasLabel(t *testing.T) {
	tests := map[string]struct {
		labels []Label
		label  Label
		found  bool
	}{
		"t1": {
			labels: nil,
			label:  "sss",
			found:  false,
		},
		"t2": {
			labels: []Label{"protein", SleepLabel},
			label:  "protein",
			found:  true,
		},
		"t3": {
			labels: []Label{"protein", SleepLabel},
			label:  WildCardLabel,
			found:  true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cb := NewCodeBlock(nil, "code", 2, "language")
			cb.AddLabels(tc.labels)
			if got := cb.HasLabel(tc.label); got != tc.found {
				t.Errorf("HasLabel(%s) = %v, want %v", tc.label, got, tc.found)
			}
		})
	}
}
