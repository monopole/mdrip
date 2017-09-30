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
	{bb, []base.Label{base.WildCardLabel}},
	{bb, []base.Label{base.SleepLabel, base.WildCardLabel}},
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
		[]string{AnonBlockName, AnonBlockName, "sleep"},
		"foo",
		"foo"},
	{"meh",
		"d1/d2/f3.md",
		array1,
		[]string{AnonBlockName, AnonBlockName, "sleep"},
		"f3",
		"d1/d2/f3.md"},
}

func TestLessonTut(t *testing.T) {
	for _, test := range ltTests {
		got := NewLessonTutFromBlockParsed(base.FilePath(test.fName), test.parsedBlocks)
		if got.Path() != base.FilePath(test.expectedPath) {
			t.Errorf("%s:\npath got\n\"%s\"\nwant\n\"%s\"\n", test.tName, got.Path(), test.expectedPath)
		}
		if got.Name() != test.expectedName {
			t.Errorf("%s:\nname got\n\"%s\"\nwant\n\"%s\"\n", test.tName, got.Name(), test.expectedName)
		}
		if len(got.Children()) != len(test.childNames) {
			t.Errorf("%s:\ngot n chilren = \n\"%d\"\nwant\n\"%d\"\n", test.tName, len(got.Children()), len(test.childNames))
		} else {
			for i, tut := range got.Children() {
				if test.childNames[i] != tut.Name() {
					t.Errorf("%s:\ngot = \n\"%s\"\nwant\n\"%s\"\n", test.tName, tut.Name(), test.childNames[i])
				}
			}
		}
	}
}
