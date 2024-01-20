package model

import (
	"github.com/monopole/mdrip/tobeinternal/base"
)

type itemType int

const (
	itemHeader itemType = iota
	itemProse
	itemCode
)

type keyItem struct {
	typ   itemType
	index int
}

type mdHeader struct {
	text   string
	weight int
}

// MdContent represents markdown content.
type MdContent struct {
	ordering []*keyItem
	code     []base.OpaqueCode
	prose    []base.MdProse
	headers  []*mdHeader
	Blocks   []*BlockParsed
}

// NewMdContent makes a new instance of MdContent.
func NewMdContent() *MdContent {
	return &MdContent{
		[]*keyItem{},
		[]base.OpaqueCode{},
		[]base.MdProse{},
		[]*mdHeader{},
		[]*BlockParsed{}}
}

// HasTitle is true if a title can be discerned from the markdown.
func (md *MdContent) HasTitle() bool {
	return len(md.headers) > 0 && md.headers[0].weight == 1
}

// GetTitle returns the most likely title of the markdown.
func (md *MdContent) GetTitle() string {
	return md.headers[0].text
}

func (md *MdContent) addOrdering(x itemType, index int) {
	md.ordering = append(md.ordering, &keyItem{x, index})
}

// AddHeader adds a header.
func (md *MdContent) AddHeader(x string, w int) {
	md.addOrdering(itemHeader, len(md.headers))
	md.headers = append(md.headers, &mdHeader{x, w})
}

// AddCode adds code.
func (md *MdContent) AddCode(x string) {
	md.addOrdering(itemCode, len(md.code))
	md.code = append(md.code, base.OpaqueCode(x))
}

// AddProse adds prose.
func (md *MdContent) AddProse(x string) {
	md.addOrdering(itemProse, len(md.prose))
	md.prose = append(md.prose, base.MdProse(x))
}

// AddBlockParsed adds an instance of BlockParsed.
func (md *MdContent) AddBlockParsed(x *BlockParsed) {
	md.Blocks = append(md.Blocks, x)
}
