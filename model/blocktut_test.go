package model

import (
	"testing"

	"github.com/monopole/mdrip/base"
)

type btTest struct {
	name       string
	block      BlockParsed
	nameWanted string
}

var btTests = []btTest{
	{"empty",
		BlockParsed{bb, []base.Label{}},
		"noName"},
	{"anylabel",
		BlockParsed{bb, []base.Label{base.AnyLabel}},
		"__AnyLabel__"},
	{"sleeplabel",
		BlockParsed{bb, []base.Label{base.SleepLabel, base.AnyLabel}},
		"sleep"},
	{"anyFirst",
		BlockParsed{bb, []base.Label{base.AnyLabel, base.SleepLabel}},
		"__AnyLabel__"},
	{"xFirst",
		BlockParsed{bb, []base.Label{base.Label("shazam"), base.AnyLabel, base.SleepLabel}},
		"shazam"},
}

func TestBlockTut(t *testing.T) {
	for _, test := range btTests {
		got := NewBlockTut(&test.block).Name()
		if got != test.nameWanted {
			t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", test.name, got, test.nameWanted)
		}
	}
}
