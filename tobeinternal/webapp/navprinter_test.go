package webapp

import (
	"bytes"
	"testing"

	"github.com/monopole/mdrip/tobeinternal/base"
	"github.com/monopole/mdrip/tobeinternal/model"
)

type npTest struct {
	name  string
	input model.Tutorial
	want  string
}

var emptyLesson = model.NewLessonTutForTests(
	base.FilePath(""),
	[]*model.BlockTut{})

var course1 = model.NewCourse(base.FilePath("hey"),
	[]model.Tutorial{emptyLesson})

var npTests = []npTest{
	{"emptyLesson",
		emptyLesson,
		`<div class='navItemTop'>
  <div id='NL0' class='navLessonTitleOff'
      onclick='lessonController.assureActiveLesson(0)'
      data-path='.'>
    .
  </div>
</div>
`}, {"smallCourse",
		course1,
		`<div class='navItemTop'>
  <div class='navCourseTitle' onclick='lessonController.ncToggle(0)'>
    hey
  </div>
  <div id='NC0' class='navCourseContent'
      style='display: none;'>
    <div class='navItemBox'>
      <div id='NL0' class='navLessonTitleOff'
          onclick='lessonController.assureActiveLesson(0)'
          data-path='hey/.'>
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
