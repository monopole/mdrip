package util

import (
	"testing"
)

type stest struct {
	name  string
	input int
	want  string
}

var stests = []stest{
	{"empty", 0, ""},
	{"one", 1, " "},
	{"five", 5, "     "},
}

func TestSpaces(t *testing.T) {
	for _, test := range stests {
		got := Spaces(test.input)
		if got != test.want {
			t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", test.name, got, test.want)
		}
	}
}

type dtest struct {
	input string
	want  string
}

var dtests = []dtest{
	{"111_hey", "hey"},
	{"hey", "hey"},
	{"0_beans", "beans"},
	{"99999s_beans", "99999s_beans"},
	{"99999_beans", "beans"},
}

func TestDropLeadingSorter(t *testing.T) {
	for _, test := range dtests {
		got := DropLeadingNumbers(test.input)
		if got != test.want {
			t.Errorf(
				"got \"%s\"\n"+
					"want\"%s\"\n", got, test.want)
		}
	}
}
