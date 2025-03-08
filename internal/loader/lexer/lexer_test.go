package lexer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type lexTest struct {
	input string // Input string to be lexed.
	want  []string
	id    string
}

var lexTests = map[string]lexTest{
	"text": {"blah1 zlAh2  ",
		[]string{"blah1", "zlah2"},
		"blah1Zlah2",
	},
	"empty":  {"", []string{}, ""},
	"spaces": {" \t\n", []string{}, ""},
	"comment1": {"<!-- cheese wHIz summer ocean-->",
		[]string{"cheese", "whiz", "summer", "ocean"},
		"cheesWhizSummer"},
	"command": {
		"sudo export FOOD=\"$meat\"",
		[]string{"sudo", "export", "food", "meat"},
		"exporFoodMeat",
	},
}

func equal(i1, i2 []string) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k] != i2[k] {
			fmt.Printf("vals not equal - got : %q\n", i1[k])
			fmt.Printf("vals not equal - want: %q\n", i2[k])
			fmt.Printf("\n")
			return false
		}
	}
	return true
}

func TestGather(t *testing.T) {
	for n, tc := range lexTests {
		t.Run(n, func(t *testing.T) {
			got := gatherAsLowerCase(tc.input)
			if !equal(got, tc.want) {
				t.Errorf(`%s:
   got %+v
  want %+v
`, n, got, tc.want)
			}
		})
	}
}

func TestMakeIdentifier(t *testing.T) {
	for n, tc := range lexTests {
		t.Run(n, func(t *testing.T) {
			got := MakeIdentifier(tc.input, 3, 5)
			assert.Equal(t, tc.id, got)
		})
	}
}
