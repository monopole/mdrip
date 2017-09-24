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
