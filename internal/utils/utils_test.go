package utils

import (
	"testing"
)

func TestSpaces(t *testing.T) {
	tests := map[string]struct {
		input int
		want  string
	}{
		"neg":   {-3, ""},
		"empty": {0, ""},
		"one":   {1, " "},
		"two":   {2, "  "},
		"three": {3, "   "},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := Spaces(tc.input)
			if got != tc.want {
				t.Errorf(
					"%s:\ngot\n\"%s\"\nwant\n\"%s\"\n",
					name, got, tc.want)
			}
		})
	}
}

func TestDropLeadingNumbers(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"t1": {"111_hey", "hey"},
		"t2": {"hey", "hey"},
		"t3": {"0_beans", "beans"},
		"t4": {"99999s_beans", "99999s_beans"},
		"t5": {"99999_beans", "beans"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := DropLeadingNumbers(tc.input); got != tc.want {
				t.Errorf(
					"got \"%s\"\n"+
						"want\"%s\"\n", got, tc.want)
			}
		})
	}
}
