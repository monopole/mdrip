package program

import (
	"fmt"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
)

// BlockPgm is input to execution.
type BlockPgm struct {
	// name is the command block name, e.g. printTheMountTable.
	// Not necessarily unique. Useful in rendering and logging.
	name string
	// Should a sleep be added?
	shouldAddSleep bool
	base.BlockBase
}

func NewEmptyBlockPgm() *BlockPgm {
	return NewBlockPgm("")
}

func NewBlockPgm(code string) *BlockPgm {
	return &BlockPgm{"noNameBlock", false,
		base.NewBlockBase([]byte{}, base.OpaqueCode(code))}
}

func NewBlockPgmFromBlockTut(b *model.BlockTut) *BlockPgm {
	return &BlockPgm{
		b.Name(),
		b.HasLabel(base.SleepLabel),
		base.NewBlockBase(b.Prose(), b.Code())}
}

func (x *BlockPgm) Name() string { return x.name }
func (x *BlockPgm) HtmlProse() template.HTML {
	return template.HTML(string(blackfriday.MarkdownCommon(x.Prose())))
}

func (x *BlockPgm) Print(
	w io.Writer, prefix string, n int, label base.Label, fileName base.FilePath) {
	fmt.Fprintf(w, "echo \"%s @%s (block #%d in %s) of %s\"\n\n",
		prefix, x.Name(), n, label, fileName)
	fmt.Fprint(w, x.Code())
	// Add a brief sleep at the end.
	// This hack gives servers placed in the background time to start, assuming
	// they can do so in the time added!  Yeah, bad.
	if x.shouldAddSleep {
		fmt.Fprint(w, "sleep 3s # Added by mdrip\n")
	}
}
