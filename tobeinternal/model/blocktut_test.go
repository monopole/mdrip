package model

import (
	"testing"

	"github.com/monopole/mdrip/tobeinternal/base"
)

type btTest struct {
	name       string
	block      BlockParsed
	nameWanted string
}

var btTests = []btTest{
	{"empty",
		BlockParsed{bb, []base.Label{}},
		AnonBlockName},
	{"anylabel",
		BlockParsed{bb, []base.Label{base.WildCardLabel}},
		AnonBlockName},
	{"sleeplabel",
		BlockParsed{bb, []base.Label{base.SleepLabel, base.WildCardLabel}},
		"sleep"},
	{"wildFirst",
		BlockParsed{bb, []base.Label{base.WildCardLabel, base.Label("hoser"), base.SleepLabel}},
		"hoser"},
	{"xFirst",
		BlockParsed{bb, []base.Label{base.Label("shazam"), base.WildCardLabel, base.SleepLabel}},
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
