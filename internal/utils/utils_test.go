package utils

import (
	"testing"
)

func TestSpaces(t *testing.T) {
	type sTest struct {
		name  string
		input int
		want  string
	}
	for _, test := range []sTest{
		{"neg", -3, ""},
		{"empty", 0, ""},
		{"one", 1, " "},
		{"five", 5, "     "},
	} {
		got := Spaces(test.input)
		if got != test.want {
			t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", test.name, got, test.want)
		}
	}
}

func TestDropLeadingNumbers(t *testing.T) {
	type dTest struct {
		input string
		want  string
	}
	for _, test := range []dTest{
		{"111_hey", "hey"},
		{"hey", "hey"},
		{"0_beans", "beans"},
		{"99999s_beans", "99999s_beans"},
		{"99999_beans", "beans"},
	} {
		if got := DropLeadingNumbers(test.input); got != test.want {
			t.Errorf(
				"got \"%s\"\n"+
					"want\"%s\"\n", got, test.want)
		}
	}
}
