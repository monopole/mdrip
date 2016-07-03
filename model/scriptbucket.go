package model

import (
	"fmt"
	"io"
	"strings"
)

// ScriptBucket associates a list of CommandBlocks with the name of the
// file they came from.
type ScriptBucket struct {
	fileName string
	script   []*CommandBlock
}

func (b ScriptBucket) GetFileName() string {
	return b.fileName
}

func (b ScriptBucket) GetScript() []*CommandBlock {
	return b.script
}

// Dump sends contents to the given Writer.
//
// If n <= 0, dump everything, else only dump the first n blocks.  n
// is a count not an index.  If you want the first two blocks dumped,
// pass n==2, not n==1.
func (bucket ScriptBucket) Dump(w io.Writer, label Label, n int) {
	fmt.Fprintf(w, "#\n# Script @%s from %s \n#\n", label, bucket.GetFileName())
	delimFmt := "#" + strings.Repeat("-", 70) + "#  %s %d of %d\n"
	blockCount := len(bucket.GetScript())
	for i, block := range bucket.GetScript() {
		if n > 0 && i >= n {
			break
		}
		fmt.Fprintf(w, delimFmt, "Start", i+1, blockCount)
		block.Dump(w, "#", i+1, label, bucket.GetFileName())
		fmt.Fprintf(w, delimFmt, "End", i+1, blockCount)
		fmt.Fprintln(w)
	}
}

func NewScriptBucket(fileName string, script []*CommandBlock) *ScriptBucket {
	return &ScriptBucket{fileName, script}
}
