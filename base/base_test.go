package base

import (
	"testing"
)

type btest struct {
	name  string
	input string
	want  string
}

var btests = []btest{
	{"empty", "", "."},
	{"one", "foo", "foo"},
	{"five", "dir1/dir2/mississippi.md", "mississippi"},
	{"onlymd", "dir1/dir2/mississippi.txt", "mississippi.txt"},
	{"onlymd", "dir1/v1.2", "v1.2"},
}

func TestBase(t *testing.T) {
	for _, test := range btests {
		f := FilePath(test.input)
		got := f.Base()
		if got != test.want {
			t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", test.name, got, test.want)
		}
	}
}
