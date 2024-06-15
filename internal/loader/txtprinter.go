package loader

import (
	"fmt"
	"io"

	"github.com/monopole/mdrip/v2/internal/utils"
)

// TxtPrinter prints a tutorial as text.
type TxtPrinter struct {
	depth int
	w     io.Writer
}

// NewTutorialTxtPrinter makes a new TxtPrinter for the given writer.
func NewTutorialTxtPrinter(w io.Writer) *TxtPrinter {
	return &TxtPrinter{0, w}
}

func (v *TxtPrinter) wrapFmt(s string) string {
	return utils.Spaces(2*v.depth) + s + "\n"
}

// Depth is how deep we are in a tutorial tree.
func (v *TxtPrinter) Depth() int {
	return v.depth
}

// Down goes deeper.
func (v *TxtPrinter) Down() {
	v.depth++
}

// Up is opposite of Down.
func (v *TxtPrinter) Up() {
	v.depth--
}

// P does a formatted print.
func (v *TxtPrinter) P(s string, a ...interface{}) {
	_, _ = fmt.Fprintf(v.w, v.wrapFmt(s), a...)
}

//// VisitBlockTut prints a BlockTut.
//func (v *TxtPrinter) VisitBlockTut(b *BlockTut) {
//	v.P("%s --- %s...", b.Name(), utils.SampleString(string(b.Code()), 60))
//}

func (v *TxtPrinter) VisitTopFolder(fl *MyTopFolder) {
	v.P("%s", fl.Name())
	v.Down()
	fl.VisitChildren(v)
	v.Up()
}

func (v *TxtPrinter) VisitFolder(fl *MyFolder) {
	v.P("%s", fl.Name())
	v.Down()
	fl.VisitChildren(v)
	v.Up()
}

func (v *TxtPrinter) VisitFile(fi *MyFile) {
	v.P("%s", fi.Name())
	//v.Down()
	//for _, x := range l.Children() {
	//	x.Accept(v)
	//}
	//v.Up()
}

func (v *TxtPrinter) Error() error { return nil }
