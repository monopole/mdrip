package webapp

import (
	"github.com/monopole/mdrip/tutorial"
	"io"
)

// NavPrinter creates leftnav html. for use in a Go template.
type NavPrinter struct {
	tutorial.TxtPrinter
	courseCounter int
	lessonCounter int
}

func NewTutorialNavPrinter(w io.Writer) *NavPrinter {
	return &NavPrinter{
		*tutorial.NewTutorialTxtPrinter(w),
		-1, -1}
}

func (v *NavPrinter) navItemStyle() string {
	if v.Depth() > 1 {
		return "navItemBox"
	}
	return "navItemTop"
}

// Not expanding blocks in the nav - too busy looking.
func (v *NavPrinter) VisitCommandBlock(x *tutorial.CommandBlock) {
}

func (v *NavPrinter) VisitLesson(x *tutorial.Lesson) {
	v.lessonCounter++
	v.P("<div class='%s'>", v.navItemStyle())
	v.Down()
	v.P("<div id='NL%d' class='navLessonTitleOff'", v.lessonCounter)
	v.P("    onclick='assureActiveLesson(%d)'", v.lessonCounter)
	v.P("    data-path='%s'>", x.Path())
	// Could loop over children here - decided not to.
	v.Down()
	v.P("%s", x.Name())
	v.Up()
	v.P("</div>")
	v.Up()
	v.P("</div>")
}

func (v *NavPrinter) VisitCourse(x *tutorial.Course) {
	v.courseCounter++
	v.P("<div class='%s'>", v.navItemStyle())
	v.Down()
	v.P("<div class='navCourseTitle' onclick='toggleNC(%d)'>", v.courseCounter)
	v.Down()
	v.P("%s", x.Name())
	v.Up()
	v.P("</div>")
	v.P("<div id='NC%d' style='display: none;'>", v.courseCounter)
	v.Down()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.Up()
	v.P("</div>")
	v.Up()
	v.P("</div>")
}

func (v *NavPrinter) VisitTopCourse(x *tutorial.TopCourse) {
	for _, c := range x.Children() {
		c.Accept(v)
	}
}
