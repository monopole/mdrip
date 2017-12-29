package webapp

import (
	"github.com/monopole/mdrip/model"
	"io"
	"strings"
)

// NavPrinter prints leftnav HTML to a Writer.
type NavPrinter struct {
	model.TxtPrinter
	courseCounter int
	lessonCounter int
	name          []string
}

// NewTutorialNavPrinter returns a new printer for the given writer.
func NewTutorialNavPrinter(w io.Writer) *NavPrinter {
	return &NavPrinter{
		*model.NewTutorialTxtPrinter(w),
		-1, -1, make([]string, 0)}
}

func (v *NavPrinter) navItemStyle() string {
	if v.Depth() > 1 {
		return "navItemBox"
	}
	return "navItemTop"
}

// VisitBlockTut does nothing.
// Not expanding blocks in the nav - too busy looking.
func (v *NavPrinter) VisitBlockTut(x *model.BlockTut) {
}

func (v *NavPrinter) addName(t model.Tutorial) {
	v.name = append(v.name, t.Name())
}

func (v *NavPrinter) rmName() {
	v.name = v.name[:len(v.name)-1]
}

func (v *NavPrinter) path() string {
	return strings.Join(v.name, "/")
}

// VisitLessonTut visits a lesson to print it.
func (v *NavPrinter) VisitLessonTut(x *model.LessonTut) {
	v.lessonCounter++
	v.addName(x)
	v.P("<div class='%s'>", v.navItemStyle())
	v.Down()
	v.P("<div id='NL%d' class='navLessonTitleOff'", v.lessonCounter)
	v.P("    onclick='lessonController.assureActiveLesson(%d)'", v.lessonCounter)
	v.P("    data-path='%s'>", v.path())
	// Could loop over children here - decided not to.
	v.Down()
	v.P("%s", x.Name())
	v.Up()
	v.P("</div>")
	v.Up()
	v.P("</div>")
	v.rmName()
}

// VisitCourse visits a course to print it.
func (v *NavPrinter) VisitCourse(x *model.Course) {
	v.courseCounter++
	v.addName(x)
	v.P("<div class='%s'>", v.navItemStyle())
	v.Down()
	v.P("<div class='navCourseTitle' onclick='lessonController.ncToggle(%d)'>",
		v.courseCounter)
	v.Down()
	v.P("%s", x.Name())
	v.Up()
	v.P("</div>")
	v.P("<div id='NC%d' class='navCourseContent'", v.courseCounter)
	v.P("    style='display: none;'>")
	v.Down()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.Up()
	v.P("</div>")
	v.Up()
	v.P("</div>")
	v.rmName()
}

// VisitTopCourse visits a course to print its children.
func (v *NavPrinter) VisitTopCourse(x *model.TopCourse) {
	for _, c := range x.Children() {
		c.Accept(v)
	}
}
