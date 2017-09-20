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
	make(map[model.Label][]*model.OldBlock))

var course1 = tutorial.NewCourse(model.FilePath("hey"),
	[]tutorial.Tutorial{emptyLesson})

var npTests = []npTest{
	{"emptyLesson",
		emptyLesson,
		`<div class='lnav1' data-name=".">
  <div onclick="assureActive('L0')">.</div>
  <div id='n1' style='display: block;'>
  </div>
</div>
`}, {"smallCourse",
		course1,
		`<div class='lnav1' data-name="hey">
  <div onclick="toggle('n1')">hey</div>
  <div id='n1' style='display: block;'>
    <div class='lnav1' data-name=".">
      <div onclick="assureActive('L0')">.</div>
      <div id='n2' style='display: none;'>
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
