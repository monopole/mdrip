package tutorial

import (
	"fmt"
	"github.com/monopole/mdrip/util"
	"io"
)

type TxtPrinter struct {
	depth int
	w     io.Writer
}

func NewTutorialTxtPrinter(w io.Writer) *TxtPrinter {
	return &TxtPrinter{0, w}
}

func (v *TxtPrinter) in(s string) string {
	return util.Spaces(2*v.depth) + s
}

func (v *TxtPrinter) down() {
	v.depth += 1
}

func (v *TxtPrinter) up() {
	v.depth -= 1
}

func (v *TxtPrinter) pf(s string, a ...interface{}) {
	fmt.Fprintf(v.w, v.in(s), a...)
}

func (v *TxtPrinter) VisitCommandBlock(b *CommandBlock) {
	v.pf("%s --- %s...\n", b.Name(), util.SampleString(string(b.Code()), 60))
}

func (v *TxtPrinter) VisitLesson(l *Lesson) {
	v.pf("%s\n", l.Name())
	v.down()
	for _, x := range l.Children() {
		x.Accept(v)
	}
	v.up()
}

func (v *TxtPrinter) VisitCourse(c *Course) {
	v.pf("%s\n", c.Name())
	v.down()
	for _, x := range c.Children() {
		x.Accept(v)
	}
	v.up()
}

func (v *TxtPrinter) VisitTopCourse(t *TopCourse) {
	for _, x := range t.Children() {
		x.Accept(v)
	}
}
