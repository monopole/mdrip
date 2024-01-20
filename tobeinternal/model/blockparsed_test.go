package model

import (
	"testing"

	"github.com/monopole/mdrip/tobeinternal/base"
)

type bpTest struct {
	name  string
	block BlockParsed
	label base.Label
	want  bool
}

var bb = base.NewBlockBase(
	base.MdProse("// prints hey"),
	base.OpaqueCode("print hey"))

var bpTests = []bpTest{
	{"empty",
		BlockParsed{bb, []base.Label{}},
		base.WildCardLabel,
		false},
	{"test1",
		BlockParsed{bb, []base.Label{base.WildCardLabel, base.SleepLabel}},
		base.WildCardLabel,
		true},
	{"test2",
		BlockParsed{bb, []base.Label{base.SleepLabel, base.WildCardLabel}},
		base.WildCardLabel,
		true},
	{"test2",
		BlockParsed{bb, []base.Label{base.SleepLabel, base.SleepLabel}},
		base.WildCardLabel,
		false},
}

func TestBlockParsed(t *testing.T) {
	for _, test := range bpTests {
		got := test.block.HasLabel(base.WildCardLabel)
		if got != test.want {
			t.Errorf("%s:\ngot %v, want %v\n", test.name, got, test.want)
		}
	}
}
