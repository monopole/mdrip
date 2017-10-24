package lexer

import (
	"fmt"
	"testing"
)

type lexTest struct {
	name  string // Name of the sub-test.
	input string // Input string to be lexed.
	want  []item // Expected items produced by lexer.
}

const (
	block1 = "echo $PATH\n" +
		"echo $GOPATH"
	block2       = "kill -9 $pid"
	indentedCode = "> ```\n" +
		"> hey\n" +
		"> ```\n"
)

var (
	tEOF = item{itemEOF, ""}
)

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t\n", []item{{itemProse, " \t\n"}, tEOF}},
	{"text", "blah blah",
		[]item{{itemProse, "blah blah"}, tEOF}},
	{"comment1", "<!-- -->", []item{tEOF}},
	{"comment2", "a <!-- --> b", []item{{itemProse, "a "}, {itemProse, " b"}, tEOF}},
	{"block1", "fred <!-- @1 -->\n" +
		"```\n" + block1 + "```\n bbb",
		[]item{
			{itemProse, "fred "},
			{itemBlockLabel, "1"},
			{itemCodeBlock, block1},
			{itemProse, "\n bbb"},
			tEOF}},
	{"block2", "aa <!-- @1 @2-->\n" +
		"```\n" + block1 + "```\n bb cc\n" +
		"dd <!-- @3 @4-->\n" +
		"```\n" + block2 + "```\n ee ff\n",
		[]item{
			{itemProse, "aa "},
			{itemBlockLabel, "1"},
			{itemBlockLabel, "2"},
			{itemCodeBlock, block1},
			{itemProse, "\n bb cc\ndd "},
			{itemBlockLabel, "3"},
			{itemBlockLabel, "4"},
			{itemCodeBlock, block2},
			{itemProse, "\n ee ff\n"},
			tEOF}},
	{"blockWithLangName", "Hello <!-- @1 -->\n" +
		"```java\nvoid main whatever\n```",
		[]item{
			{itemProse, "Hello "},
			{itemBlockLabel, "1"},
			{itemCodeBlock, "void main whatever\n"},
			tEOF}},
	{"blockNoLabel", "fred\n" +
		"```\n" + block1 + "```\n bbb",
		[]item{
			{itemProse, "fred\n"},
			{itemCodeBlock, block1},
			{itemProse, "\n bbb"},
			tEOF}},
	{"blockQuote",
		"fred\n" +
			indentedCode +
			"bbb",
		[]item{
			{itemProse, "fred\n" + indentedCode + "bbb"},
			tEOF}},
}

var XXlexTests = []lexTest{
}

var DRAIN_lexTests = []lexTest{}

var ORIGINAL_lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t\n", []item{{itemProse, " \t\n"}, tEOF}},
	{"text", "blah blah",
		[]item{{itemProse, "blah blah"}, tEOF}},
	{"comment1", "<!-- -->", []item{tEOF}},
	{"comment2", "a <!-- --> b", []item{{itemProse, "a "}, {itemProse, " b"}, tEOF}},
	{"block1", "fred <!-- @1 -->\n" +
		"```\n" + block1 + "```\n bbb",
		[]item{
			{itemProse, "fred "},
			{itemBlockLabel, "1"},
			{itemCodeBlock, block1},
			{itemProse, "\n bbb"},
			tEOF}},
	{"block2", "aa <!-- @1 @2-->\n" +
		"```\n" + block1 + "```\n bb cc\n" +
		"dd <!-- @3 @4-->\n" +
		"```\n" + block2 + "```\n ee ff\n",
		[]item{
			{itemProse, "aa "},
			{itemBlockLabel, "1"},
			{itemBlockLabel, "2"},
			{itemCodeBlock, block1},
			{itemProse, "\n bb cc\ndd "},
			{itemBlockLabel, "3"},
			{itemBlockLabel, "4"},
			{itemCodeBlock, block2},
			{itemProse, "\n ee ff\n"},
			tEOF}},
	{"blockWithLangName", "Hello <!-- @1 -->\n" +
		"```java\nvoid main whatever\n```",
		[]item{
			{itemProse, "Hello "},
			{itemBlockLabel, "1"},
			{itemCodeBlock, "void main whatever\n"},
			tEOF}},
	{"blockNoLabel", "fred\n" +
		"```\n" + block1 + "```\n bbb",
		[]item{
			{itemProse, "fred\n"},
			{itemCodeBlock, block1},
			{itemProse, "\n bbb"},
			tEOF}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []item) {
	l := newLex(t.input)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}

func equal(i1, i2 []item) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			fmt.Printf("types not equal - got : %v\n", i1[k].typ)
			fmt.Printf("types not equal - want: %v\n", i2[k].typ)
			fmt.Printf("\n")
			return false
		}
		if i1[k].val != i2[k].val {
			fmt.Printf("vals not equal - got : %q\n", i1[k].val)
			fmt.Printf("vals not equal - want: %q\n", i2[k].val)
			fmt.Printf("\n")
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		got := collect(&test)
		if !equal(got, test.want) {
			t.Errorf("%s:\ngot\n\t%+v\nwant\n\t%v\n", test.name, got, test.want)
			t.Errorf("Details - got:\n")
			for i, c := range got {
				t.Errorf("   %d %s\n\"%s\"\n\n", i, textType(c.typ), c.val)
			}
			t.Errorf("Details - want:\n")
			for i, c := range test.want {
				t.Errorf("   %d %s\n\"%s\"\n\n", i, textType(c.typ), c.val)
			}
		}
	}
}
