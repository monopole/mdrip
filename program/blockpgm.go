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
	id             int
	base.BlockBase
}

// NewEmptyBlockPgm returns an empty block.
func NewEmptyBlockPgm() *BlockPgm {
	return NewBlockPgm("")
}

// NewBlockPgm returns a block with the given code.
func NewBlockPgm(code string) *BlockPgm {
	return &BlockPgm{"noNameBlock", false, -1,
		base.NewBlockBase(base.NoProse(), base.OpaqueCode(code))}
}

// NewBlockPgmFromBlockTut converts a BlockTut to a BlockPgm.
func NewBlockPgmFromBlockTut(b *model.BlockTut) *BlockPgm {
	return &BlockPgm{
		b.Name(),
		b.HasLabel(base.SleepLabel), -1,
		base.NewBlockBase(b.Prose(), b.Code())}
}

// ID returns the block's ID.
func (x *BlockPgm) ID() int { return x.id }

// Name returns the block name.
func (x *BlockPgm) Name() string { return x.name }

// HTMLProse returns HTML that should precede the block.
func (x *BlockPgm) HTMLProse() template.HTML {
	return template.HTML(string(blackfriday.Run(x.Prose())))
}

// Print prints the block.
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
