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

func (v *TxtPrinter) wrapFmt(s string) string {
	return util.Spaces(2*v.depth) + s + "\n"
}

func (v *TxtPrinter) Depth() int {
	return v.depth
}

func (v *TxtPrinter) Down() {
	v.depth += 1
}

func (v *TxtPrinter) Up() {
	v.depth -= 1
}

func (v *TxtPrinter) P(s string, a ...interface{}) {
	fmt.Fprintf(v.w, v.wrapFmt(s), a...)
}

func (v *TxtPrinter) VisitCommandBlock(b *CommandBlock) {
	v.P("%s --- %s...", b.Name(), util.SampleString(string(b.Code()), 60))
}

func (v *TxtPrinter) VisitLesson(l *Lesson) {
	v.P("%s", l.Name())
	v.Down()
	for _, x := range l.Children() {
		x.Accept(v)
	}
	v.Up()
}

func (v *TxtPrinter) VisitCourse(c *Course) {
	v.P("%s", c.Name())
	v.Down()
	for _, x := range c.Children() {
		x.Accept(v)
	}
	v.Up()
}

func (v *TxtPrinter) VisitTopCourse(t *TopCourse) {
	for _, x := range t.Children() {
		x.Accept(v)
	}
}
