package model

import (
	"testing"

	"github.com/monopole/mdrip/base"
)

type ltTest struct {
	tName        string
	fName        string
	parsedBlocks []*BlockParsed
	childNames   []string
	expectedName string
	expectedPath string
}

var array1 = []*BlockParsed{
	{bb, []base.Label{}},
	{bb, []base.Label{base.AnyLabel}},
	{bb, []base.Label{base.SleepLabel, base.AnyLabel}},
}

var ltTests = []ltTest{
	{"emptyempty",
		"",
		[]*BlockParsed{},
		[]string{},
		".",
		""},
	{"foo",
		"foo",
		array1,
		[]string{"noName", "__AnyLabel__", "sleep"},
		"foo",
		"foo"},
	{"meh",
		"d1/d2/f3.md",
		array1,
		[]string{"noName", "__AnyLabel__", "sleep"},
		"f3",
		"d1/d2/f3.md"},
}

func TestLessonTut(t *testing.T) {
	for _, test := range ltTests {
		got := NewLessonTutFromBlockParsed(base.FilePath(test.fName), test.parsedBlocks)
		if got.Path() != base.FilePath(test.expectedPath) {
			t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", test.tName, got.Path(), test.expectedPath)
		}
		if got.Name() != test.expectedName {
			t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", test.tName, got.Name(), test.expectedName)
		}
		if len(got.Children()) != len(test.childNames) {
			t.Errorf("%s:\ngot n chilren = \n\"%d\"\nwant\n\"%d\"\n", test.tName, len(got.Children()), len(test.childNames))
		} else {
			for i, tut := range got.Children() {
				if test.childNames[i] != tut.Name() {
					t.Errorf("%s:\ngot = \n\"%s\"\nwant\n\"%s\"\n", test.tName, test.childNames[i], tut.Name())
				}
			}
		}
	}
}
