package lexer

import (
	"fmt"
	"testing"
)

type lexTest struct {
	name  string      // Name of the sub-test.
	input string      // Input string to be lexed.
	want  []lexedItem // Expected items produced by lexer.
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
	tEOF = lexedItem{itemEOF, ""}
)

var lexTests = []lexTest{
	{"empty",
		"",
		[]lexedItem{tEOF}},
	{"spaces",
		" \t\n",
		[]lexedItem{{itemProse, " \t\n"}, tEOF}},
	{"text",
		"blah blah",
		[]lexedItem{{itemProse, "blah blah"}, tEOF}},
	{"header1",
		"#cheese",
		[]lexedItem{{itemHeader1, "cheese"}, tEOF}},
	{"comment1",
		"<!-- -->",
		[]lexedItem{tEOF}},
	{"comment2",
		"a <!-- --> b",
		[]lexedItem{{itemProse, "a "}, {itemProse, " b"}, tEOF}},
	{"block1",
		"fred <!-- @1 -->\n" + "```\n" + block1 + "```\n bbb",
		[]lexedItem{
			{itemProse, "fred "},
			{itemBlockLabel, "1"},
			{itemCodeBlock, block1},
			{itemProse, "\n bbb"},
			tEOF}},
	{"block2",
		"aa <!-- @1 @2-->\n" +
			"```\n" + block1 + "```\n bb cc\n" +
			"dd <!-- @3 @4-->\n" +
			"```\n" + block2 + "```\n ee ff\n",
		[]lexedItem{
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
	{"blockWithLangName",
		"Hello <!-- @1 -->\n" +
			"```java\nvoid main whatever\n```",
		[]lexedItem{
			{itemProse, "Hello "},
			{itemBlockLabel, "1"},
			{itemCodeBlock, "void main whatever\n"},
			tEOF}},
	{"blockNoLabel",
		"fred\n" +
			"```\n" + block1 + "```\n bbb",
		[]lexedItem{
			{itemProse, "fred\n"},
			{itemCodeBlock, block1},
			{itemProse, "\n bbb"},
			tEOF}},
	{"blockQuote",
		"fred\n" + indentedCode + "bbb",
		[]lexedItem{
			{itemProse, "fred\n" + indentedCode + "bbb"},
			tEOF}},
	{"header1",
		"#cheese",
		[]lexedItem{{itemHeader1, "cheese"}, tEOF}},
	{"header2",
		"##     carrot celery",
		[]lexedItem{{itemHeader2, "carrot celery"}, tEOF}},
	{"header6IsMostest",
		"######## #x",
		[]lexedItem{{itemHeader6, "x"}, tEOF}},
	{"notHeaderIfNotAtStart",
		"  ## x",
		[]lexedItem{{itemProse, "  ## x"}, tEOF}},
	{"notHeaderIfNotAtStart",
		"  hey\n### x\nbob",
		[]lexedItem{
			{itemProse, "  hey\n"},
			{itemHeader3, "x"},
			{itemProse, "bob"},
			tEOF}},
	{"headerHeader",
		"  hey\n### x3\n#### x4\n# x1\nbob",
		[]lexedItem{
			{itemProse, "  hey\n"},
			{itemHeader3, "x3"},
			{itemHeader4, "x4"},
			{itemHeader1, "x1"},
			{itemProse, "bob"},
			tEOF}},
	{"headerInHeader",
		"  hey\n### x3 ###\n#### x4",
		[]lexedItem{
			{itemProse, "  hey\n"},
			{itemHeader3, "x3 ###"},
			{itemHeader4, "x4"},
			tEOF}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []lexedItem) {
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

func equal(i1, i2 []lexedItem) bool {
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
