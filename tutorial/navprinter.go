package tutorial

import (
	"fmt"
	"github.com/monopole/mdrip/util"
	"io"
)

type NavPrinter struct {
	depth int
	id    int
	lCounter int
	w     io.Writer
}

func NewTutorialNavPrinter(w io.Writer) *NavPrinter {
	return &NavPrinter{0, 0, -1, w}
}

func (v *NavPrinter) in(s string) string {
	return util.Spaces(2*v.depth) + s
}

func (v *NavPrinter) incId() {
	v.id += 1
}

func (v *NavPrinter) down() {
	v.depth += 1
}

func (v *NavPrinter) up() {
	v.depth -= 1
}

func (v *NavPrinter) pf(s string, a ...interface{}) {
	fmt.Fprintf(v.w, v.in(s), a...)
}

func (v *NavPrinter) VisitCommandBlock(x *CommandBlock) {
	//v.incId()
	//v.pf("<div> %s </div>\n", x.Name())
	//v.pf("%s %s\n", b.Name(), util.SampleString(string(b.Code()), 60))
}

func (v *NavPrinter) VisitLesson(x *Lesson) {
	v.lCounter++
	v.incId()
	v.pf("<div class='lnav1' data-name=\"%s\">\n", x.Name())
	v.down()
	// instead of toggle, call assureOnlyThisGuyOn!
	v.pf("<div onclick=\"assureActive('L%d')\">%s</div>\n", v.lCounter, x.Name())
	v.pf("<div id='n%d' style='display: %s;'>\n", v.id, v.initStyle())
	v.down()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.up()
	v.pf("</div>\n")
	v.up()
	v.pf("</div>\n")
}

func (v *NavPrinter) initStyle() string {
	if v.depth > 1 {
		return "none"
	}
	return "block"
}

func (v *NavPrinter) VisitCourse(x *Course) {
	v.incId()
	v.pf("<div class='lnav1' data-name=\"%s\">\n", x.Name())
	v.down()
	v.pf("<div onclick=\"toggle('n%d')\">%s</div>\n", v.id, x.Name())
	v.pf("<div id='n%d' style='display: %s;'>\n", v.id, v.initStyle())
	v.down()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.up()
	v.pf("</div>\n")
	v.up()
	v.pf("</div>\n")
}

func (v *NavPrinter) VisitTopCourse(x *TopCourse) {
	v.pf("<div class='lnav0' data-name=\"%s\">\n", x.Path())
	v.down()
	for _, c := range x.Children() {
		c.Accept(v)
	}
	v.up()
	v.pf("</div>\n")
}
