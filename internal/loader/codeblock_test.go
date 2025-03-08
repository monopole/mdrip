package loader_test

import (
	"testing"

	. "github.com/monopole/mdrip/v2/internal/loader"
	"github.com/stretchr/testify/assert"
)

func Test_codeBlock_HasLabel(t *testing.T) {
	tests := []struct {
		tn           string
		labels       []Label
		code         string
		labToCheck   Label
		expectedName string
		found        bool
	}{
		{
			tn:           "t1",
			labels:       nil,
			code:         "apt get meat ball",
			labToCheck:   "sss",
			found:        false,
			expectedName: "aptGetMeatBall",
		},
		{
			tn:           "t2",
			labels:       []Label{"protein", SleepLabel},
			code:         "sudo apt get meat ball",
			labToCheck:   "protein",
			found:        true,
			expectedName: "protein",
		},
		{
			tn:           "t3",
			labels:       []Label{"protein", SleepLabel},
			code:         "apt get meat ball",
			labToCheck:   SleepLabel,
			found:        true,
			expectedName: "protein2",
		},
		{
			tn:           "t4",
			labels:       []Label{SkipLabel, SleepLabel},
			code:         "apt get meat ball",
			labToCheck:   "protein",
			found:        false,
			expectedName: "aptGetMeatBall2",
		},
		{
			tn:           "t5",
			code:         "apt get meat ball",
			labToCheck:   "protein",
			found:        false,
			expectedName: "aptGetMeatBall3",
		},
		{
			tn:           "t6",
			code:         "apt get meat balloon",
			labToCheck:   "protein",
			found:        false,
			expectedName: "aptGetMeatBalloon",
		},
	}
	disAmbig := make(map[string]int)
	for _, tc := range tests {
		t.Run(tc.tn, func(t *testing.T) {
			cb := NewCodeBlock(
				nil, tc.code, 2, tc.labels...)
			if got := cb.HasLabel(tc.labToCheck); got != tc.found {
				t.Errorf("HasLabel(%s) = %v, want %v",
					tc.labToCheck, got, tc.found)
			}
			cb.ResetTitle(disAmbig)
			assert.Equal(t, tc.expectedName, cb.UniqName())
		})
	}
}
