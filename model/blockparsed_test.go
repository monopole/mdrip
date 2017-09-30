package model

import (
	"testing"

	"github.com/monopole/mdrip/base"
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
		base.AnyLabel,
		false},
	{"test1",
		BlockParsed{bb, []base.Label{base.AnyLabel, base.SleepLabel}},
		base.AnyLabel,
		true},
	{"test2",
		BlockParsed{bb, []base.Label{base.SleepLabel, base.AnyLabel}},
		base.AnyLabel,
		true},
	{"test2",
		BlockParsed{bb, []base.Label{base.SleepLabel, base.SleepLabel}},
		base.AnyLabel,
		false},
}

func TestBlockParsed(t *testing.T) {
	for _, test := range bpTests {
		got := test.block.HasLabel(base.AnyLabel)
		if got != test.want {
			t.Errorf("%s:\ngot %v, want %v\n", test.name, got, test.want)
		}
	}
}
