package loader

import (
	"bytes"
	"testing"
)

func TestTxtPrinter(t *testing.T) {
	type testCase struct {
		input *MyFolder
		want  string
	}
	for n, test := range map[string]testCase{
		"smallTree": {
			input: NewFolder("D0").
				AddFile(NewEmptyFile("F0")).
				AddFile(NewEmptyFile("F1")),
			want: `
D0
  F0
  F1
`[1:],
		},
		"biggerTree": {
			input: NewFolder("D0").
				AddFile(NewEmptyFile("F0")).
				AddFile(NewEmptyFile("F1")).
				AddFile(NewEmptyFile("F2")).
				AddFolder(NewFolder("D1").
					AddFile(NewEmptyFile("F3")).
					AddFile(NewEmptyFile("F4")).
					AddFolder(NewFolder("D2").
						AddFile(NewEmptyFile("F5")).
						AddFile(NewEmptyFile("F6")))).
				AddFolder(NewFolder("D3").
					AddFile(NewEmptyFile("F7")).
					AddFile(NewEmptyFile("F8"))),
			want: `
D0
  F0
  F1
  F2
  D1
    F3
    F4
    D2
      F5
      F6
  D3
    F7
    F8
`[1:],
		},
	} {
		t.Run(n, func(t *testing.T) {
			var b bytes.Buffer
			v := NewTutorialTxtPrinter(&b)
			test.input.Accept(v)
			got := b.String()
			if got != test.want {
				t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", n, got, test.want)
			}
		})
	}
}
