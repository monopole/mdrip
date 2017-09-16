package tutorial

import (
	"fmt"
	"github.com/monopole/mdrip/util"
	"io"
	"strconv"
)

type Printer struct {
	indent int
	w      io.Writer
}

func NewTutorialPrinter(w io.Writer) *Printer {
	return &Printer{0, w}
}

func (v *Printer) spaces(indent int) string {
	if indent < 1 {
		return ""
	}
	return fmt.Sprintf("%"+strconv.Itoa(indent)+"s", " ")
}

func (v *Printer) VisitLesson(l *Lesson) {
	fmt.Fprintf(v.w,
		v.spaces(v.indent)+"%s --- %s...\n",
		l.Name(), util.SampleString(l.Content(), 60))
}

func (v *Printer) VisitCourse(c *Course) {
	fmt.Fprintf(v.w, v.spaces(v.indent)+"%s\n", c.Name())
	v.indent += 3
	for _, x := range c.children {
		x.Accept(v)
	}
	v.indent -= 3
}

func (v *Printer) VisitTopCourse(t *TopCourse) {
	for _, x := range t.children {
		x.Accept(v)
	}
}
