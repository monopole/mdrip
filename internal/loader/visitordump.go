package loader

import (
	"fmt"
	"github.com/monopole/mdrip/v2/internal/utils"
	"io"
)

type VisitorDump struct {
	wr     io.Writer
	indent int
}

func NewVisitorDump(wr io.Writer) *VisitorDump {
	return &VisitorDump{
		wr:     wr,
		indent: 0,
	}
}

const blanks = "                                                                "

func (v *VisitorDump) VisitTopFolder(fl *MyTopFolder) {
	_, _ = fmt.Fprint(v.wr, blanks[:v.indent])
	_, _ = fmt.Fprint(v.wr, fl.Name()) // This is always blank
	//if !fl.IsRoot() {
	//	_, _ = fmt.Fprint(v.wr, RootSlash)
	//}
	_, _ = fmt.Fprintln(v.wr)
	v.indent += 2
	fl.VisitChildren(v)
	v.indent -= 2
}

func (v *VisitorDump) VisitFolder(fl *MyFolder) {
	_, _ = fmt.Fprint(v.wr, blanks[:v.indent])
	_, _ = fmt.Fprint(v.wr, fl.Name())
	_, _ = fmt.Fprintln(v.wr)
	v.indent += 2
	fl.VisitChildren(v)
	v.indent -= 2
}

func (v *VisitorDump) VisitFile(fi *MyFile) {
	_, _ = fmt.Fprint(v.wr, blanks[:v.indent])
	_, _ = fmt.Fprint(v.wr, fi.Name())
	_, _ = fmt.Fprint(v.wr, " : ")
	_, _ = fmt.Fprintln(v.wr, utils.Summarize(fi.C())+"...")
}

func (v *VisitorDump) Error() error { return nil }
