package model

import "github.com/monopole/mdrip/base"

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

type MdContent struct {
	ordering []*keyItem
	code     []base.OpaqueCode
	prose    []base.MdProse
	headers  []*mdHeader
	Blocks   []*BlockParsed
}

func NewMdContent() *MdContent {
	return &MdContent{
		[]*keyItem{},
		[]base.OpaqueCode{},
		[]base.MdProse{},
		[]*mdHeader{},
		[]*BlockParsed{}}
}

func (md *MdContent) HasTitle() bool {
	return len(md.headers) > 0 && md.headers[0].weight == 1
}

func (md *MdContent) GetTitle() string {
	return md.headers[0].text
}

func (md *MdContent) addOrdering(x itemType, index int) {
	md.ordering = append(md.ordering, &keyItem{x, index})
}

func (md *MdContent) AddHeader(x string, w int) {
	md.addOrdering(itemHeader, len(md.headers))
	md.headers = append(md.headers, &mdHeader{x, w})
}

func (md *MdContent) AddCode(x string) {
	md.addOrdering(itemCode, len(md.code))
	md.code = append(md.code, base.OpaqueCode(x))
}

func (md *MdContent) AddProse(x string) {
	md.addOrdering(itemProse, len(md.prose))
	md.prose = append(md.prose, base.MdProse(x))
}

func (md *MdContent) AddBlockParsed(x *BlockParsed) {
	md.Blocks = append(md.Blocks, x)
}
