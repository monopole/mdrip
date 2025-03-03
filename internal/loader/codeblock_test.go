package loader_test

import (
	. "github.com/monopole/mdrip/v2/internal/loader"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_codeBlock_HasLabel(t *testing.T) {
	tests := map[string]struct {
		labels []Label
		label  Label
		name   string
		found  bool
	}{
		"t1": {
			labels: nil,
			label:  "sss",
			found:  false,
			name:   "someNiceCode02",
		},
		"t2": {
			labels: []Label{"protein", SleepLabel},
			label:  "protein",
			found:  true,
			name:   "protein",
		},
		"t3": {
			labels: []Label{"protein", SleepLabel},
			label:  SleepLabel,
			found:  true,
			name:   "protein",
		},
		"t4": {
			labels: []Label{SkipLabel, SleepLabel},
			label:  "protein",
			found:  false,
			name:   "someNiceCode02",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cb := NewCodeBlock(
				nil, "some nice code in your face", 2, tc.labels...)
			if got := cb.HasLabel(tc.label); got != tc.found {
				t.Errorf("HasLabel(%s) = %v, want %v", tc.label, got, tc.found)
			}
			assert.Equal(t, tc.name, cb.Name())
		})
	}
}
