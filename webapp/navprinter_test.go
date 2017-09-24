package webapp

import (
	"bytes"
	"testing"

	"github.com/monopole/mdrip/model"
	"github.com/monopole/mdrip/tutorial"
)

type npTest struct {
	name  string
	input tutorial.Tutorial
	want  string
}

var emptyLesson = tutorial.NewLesson(
	model.FilePath(""),
	[]*tutorial.CommandBlock{})

var course1 = tutorial.NewCourse(model.FilePath("hey"),
	[]tutorial.Tutorial{emptyLesson})

var npTests = []npTest{
	{"emptyLesson",
		emptyLesson,
		`<div class='navItemTop'>
  <div id='NL0' class='navLessonTitleOff'
      onclick='assureActiveLesson(0)'
      data-path=''>
    .
  </div>
</div>
`}, {"smallCourse",
		course1,
		`<div class='navItemTop'>
  <div class='navCourseTitle' onclick='toggleNC(0)'>
    hey
  </div>
  <div id='NC0' style='display: none;'>
    <div class='navItemBox'>
      <div id='NL0' class='navLessonTitleOff'
          onclick='assureActiveLesson(0)'
          data-path=''>
        .
      </div>
    </div>
  </div>
</div>
`}}

func TestNavPrinter(t *testing.T) {
	for _, test := range npTests {
		var b bytes.Buffer
		v := NewTutorialNavPrinter(&b)
		test.input.Accept(v)
		got := b.String()
		if got != test.want {
			t.Errorf("%s:\ngot\n\"%s\"\nwant\n\"%s\"\n", test.name, got, test.want)
		}
	}
}
