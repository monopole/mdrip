package model

import (
	"bytes"
	"testing"
)

type tpTest struct {
	name  string
	input Tutorial
	want  string
}

var emptyLesson = NewLesson(
	FilePath(""),
	[]*CommandBlock{})

var course1 = NewCourse(FilePath("hey"),
	[]Tutorial{emptyLesson})

var npTests = []tpTest{
	{"emptyLesson",
		emptyLesson,
		`.
`}, {"smallCourse",
		course1,
		`hey
  .
`}}

func TestTxtPrinter(t *testing.T) {
	for _, test := range npTests {
		var b bytes.Buffer
		v := NewTutorialTxtPrinter(&b)
		test.input.Accept(v)
		got := b.String()
		if got != test.want {
			t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", test.name, got, test.want)
		}
	}
}
