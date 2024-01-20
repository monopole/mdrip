// Inspired by golang.org/src/pkg/text/template/parse/lex.go

package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/monopole/mdrip/tobeinternal/base"
	"github.com/monopole/mdrip/tobeinternal/model"
)

type position int

type itemType int

// Things that the lexer emits.
const (
	itemError      itemType = iota
	itemBlockLabel          // Label for a command block
	itemCodeBlock           // All lines between codeFence marks
	itemHeader1             // Header1
	itemHeader2             // Header2
	itemHeader3             // Header3
	itemHeader4             // Header4
	itemHeader5             // Header5
	itemHeader6             // Header6
	itemProse               // Anything other than the above.
	itemEOF
)

func textType(t itemType) string {
	switch t {
	case itemError:
		return "ERROR"
	case itemBlockLabel:
		return "LABEL"
	case itemCodeBlock:
		return "BLOCK"
	case itemHeader1:
		return "H1"
	case itemHeader2:
		return "H2"
	case itemHeader3:
		return "H3"
	case itemHeader4:
		return "H4"
	case itemHeader5:
		return "H5"
	case itemHeader6:
		return "H6"
	case itemProse:
		return "PROSE"
	case itemEOF:
		return "EOF"
	default:
		return "UNKNOWN"
	}
}

type lexedItem struct {
	typ itemType // Type of this lexedItem.
	val string   // The value of this lexedItem.
}

func (i lexedItem) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ == itemBlockLabel:
		return string(labelMarker) + i.val
	case i.typ == itemCodeBlock:
		return "--------\n" + i.val + "--------\n"
	case i.typ == itemHeader1:
		return "# " + i.val
	case i.typ == itemHeader2:
		return "## " + i.val
	case i.typ == itemHeader3:
		return "### " + i.val
	case i.typ == itemHeader4:
		return "#### " + i.val
	case i.typ == itemHeader5:
		return "##### " + i.val
	case i.typ == itemHeader6:
		return "###### " + i.val
	case len(i.val) > 40-3:
		return fmt.Sprintf("%.40s...", i.val)
	}
	return fmt.Sprintf("%s", i.val)
}

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input   string         // string being scanned
	state   stateFn        // the next lexing function to enter
	current position       // current position in 'input'
	start   position       // start of this lexedItem
	width   position       // width of last rune read
	items   chan lexedItem // channel of scanned items
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.current) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.current:])
	l.width = position(w)
	l.current += l.width
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.current -= l.width
}

func (l *lexer) emit(t itemType) {
	l.items <- lexedItem{t, l.input[l.start:l.current]}
	l.start = l.current
}

func (l *lexer) ignore() {
	l.start = l.current
}

// Consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptEol() {
	if !isEndOfLine(l.next()) {
		l.backup()
	}
}

// Consumes a run of runes from the valid set
func (l *lexer) acceptRun(valid string) {
	// is the next character of the input an element
	// of the (defining) 'valid' set of runes (a string).
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- lexedItem{itemError, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next lexedItem from the input.
func (l *lexer) nextItem() lexedItem {
	item := <-l.items
	return item
}

// newLex creates a new scanner for the input string.
func newLex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan lexedItem),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for l.state = lexText; l.state != nil; {
		l.state = l.state(l)
	}
}

func isSpace(r rune) bool {
	return r == blank || r == '\t'
}

func isEndOfLine(r rune) bool {
	return r == carriageReturn || r == newLine
}

const (
	labelMarker      = '@'
	headerMarker     = '#'
	commentOpen      = "<!--"
	commentClose     = "-->"
	codeTick         = "`"
	codeFence        = codeTick + codeTick + codeTick
	blockQuoteIndent = ">"
	carriageReturn   = '\r'
	newLine          = '\n'
	blank            = ' '
	tabChar          = '\t'
	underScore       = "_"
	// < html comment start
	msSpecialChar        = "<" + string(headerMarker) + blockQuoteIndent + codeTick
	whiteSpace           = string(blank) + string(tabChar)
	miscChar             = ",.?!-$%^&*()=+|~{}[];:/'\"\\" + whiteSpace + string(labelMarker) + underScore
	lettersAndNumbers    = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	acceptableProse      = miscChar + string(headerMarker) + lettersAndNumbers
	acceptableBlockQuote = miscChar + msSpecialChar + lettersAndNumbers
	acceptableLabel      = underScore + lettersAndNumbers
)

func isBlockQuoteStart(remainder string) bool {
	return strings.HasPrefix(remainder, blockQuoteIndent) ||
		strings.HasPrefix(remainder, " "+blockQuoteIndent) ||
		strings.HasPrefix(remainder, "  "+blockQuoteIndent)
}

// Roughly this reads line by line and changes behavior.
func lexText(l *lexer) stateFn {
	for {
		l.acceptEol()
		remainder := l.input[l.current:]
		if strings.HasPrefix(remainder, commentOpen) {
			if l.current > l.start {
				l.emit(itemProse)
			}
			return lexPutativeComment
		}
		if strings.HasPrefix(remainder, codeFence) {
			if l.current > l.start {
				l.emit(itemProse)
			}
			return lexCodeBlock
		}
		if strings.HasPrefix(remainder, string(headerMarker)) {
			if l.current > l.start {
				l.emit(itemProse)
			}
			return lexHeader
		}
		if isBlockQuoteStart(remainder) {
			l.acceptRun(acceptableBlockQuote)
			return lexBlockQuote
		}
		if l.next() == eof {
			if l.current > l.start {
				l.emit(itemProse)
			}
			l.ignore()
			l.emit(itemEOF)
			return nil
		}
		l.acceptRun(acceptableProse)
	}
}

// Same as lexText, except ignore comments and code fences.
func lexBlockQuote(l *lexer) stateFn {
	for {
		l.acceptEol()
		if !isBlockQuoteStart(l.input[l.current:]) {
			return lexText
		}
		if l.next() == eof {
			if l.current > l.start {
				l.emit(itemProse)
			}
			l.ignore()
			l.emit(itemEOF)
			return nil
		}
		l.acceptRun(acceptableBlockQuote)
	}
}

// Lex a header.
func lexHeader(l *lexer) stateFn {
	weight := 0
	for {
		switch r := l.next(); {
		case isSpace(r):
			l.ignore()
		case r == headerMarker:
			weight++
			l.ignore()
		default:
			l.acceptRun(acceptableProse)
			switch weight {
			case 1:
				l.emit(itemHeader1)
			case 2:
				l.emit(itemHeader2)
			case 3:
				l.emit(itemHeader3)
			case 4:
				l.emit(itemHeader4)
			case 5:
				l.emit(itemHeader5)
			default:
				l.emit(itemHeader6)
			}
			r = l.next()
			if isEndOfLine(r) {
				l.ignore()
			}
			return lexText
		}
	}
}

// Move to lexing a command block intended for a particular label, or to
// lexing a simple comment.  Comment opener known to be present.
func lexPutativeComment(l *lexer) stateFn {
	l.current += position(len(commentOpen))
	for {
		switch r := l.next(); {
		case isSpace(r):
			l.ignore()
		case r == labelMarker:
			l.backup()
			return lexBlockLabels
		default:
			l.backup()
			return lexCommentRemainder
		}
	}
}

// lexCommentRemainder assumes a comment opener was read,
// and eats everything up to and including the comment closer.
func lexCommentRemainder(l *lexer) stateFn {
	i := strings.Index(l.input[l.current:], commentClose)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.current += position(i + len(commentClose))
	l.ignore()
	return lexText
}

// lexBlockLabels scans a string like "@1 @hey" emitting the labels
// "1" and "hey".  LabelMarker known to be present.
func lexBlockLabels(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == eof || isEndOfLine(r):
			return l.errorf("unclosed block label sequence")
		case isSpace(r):
			l.ignore()
		case r == labelMarker:
			l.ignore()
			l.acceptRun(acceptableLabel)
			if l.width == 0 {
				return l.errorf("empty block label")
			}
			l.emit(itemBlockLabel)
		default:
			l.backup()
			if !strings.HasPrefix(l.input[l.current:], commentClose) {
				return l.errorf("improperly closed block label sequence")
			}
			l.current += position(len(commentClose))
			l.ignore()
			l.acceptRun(whiteSpace)
			l.ignore()
			r := l.next()
			if !isEndOfLine(r) {
				return l.errorf("Expected command block marker at start of line.")
			}
			l.ignore()
			if !strings.HasPrefix(l.input[l.current:], codeFence) {
				return l.errorf("Expected command block mark, got: " + l.input[l.current:])
			}
			return lexCodeBlock
		}
	}
}

// lexCodeBlock scans a command block.  Initial marker known to be present.
func lexCodeBlock(l *lexer) stateFn {
	l.current += position(len(codeFence))
	l.ignore()
	// Ignore any language specifier.
	if idx := strings.Index(l.input[l.current:], "\n"); idx > -1 {
		l.current += position(idx) + 1
		l.ignore()
	}
	for {
		if strings.HasPrefix(l.input[l.current:], codeFence) {
			if l.current > l.start {
				l.emit(itemCodeBlock)
			}
			l.current += position(len(codeFence))
			l.ignore()
			return lexText
		}
		if l.next() == eof {
			return l.errorf("unclosed command block")
		}
	}
}

func isHeader(x itemType) bool {
	return x == itemHeader1 ||
		x == itemHeader2 ||
		x == itemHeader3 ||
		x == itemHeader4 ||
		x == itemHeader5 ||
		x == itemHeader6
}

func headerWeight(x itemType) int {
	switch x {
	case itemHeader1:
		return 1
	case itemHeader2:
		return 2
	case itemHeader3:
		return 3
	case itemHeader4:
		return 4
	case itemHeader5:
		return 5
	default:
		return 6
	}
}

// Parse lexes the incoming string into a list of model.BlockParsed.
func Parse(s string) *model.MdContent {
	result := model.NewMdContent()
	prose := ""
	var labels []base.Label
	l := newLex(s)
	for {
		item := l.nextItem()
		switch {
		case item.typ == itemEOF || item.typ == itemError:
			prose = strings.TrimSpace(prose)
			if len(prose) > 0 {
				// Hack to grab the last bit of prose.
				// The data structure returned by Parse needs redesign.
				result.AddBlockParsed(
					model.NewBlockParsed(
						base.NoLabels(), base.MdProse(prose), base.NoCode()))
			}
			return result
		case item.typ == itemBlockLabel:
			labels = append(labels, base.Label(item.val))
		case item.typ == itemProse:
			prose += item.val
			result.AddProse(item.val)
		case isHeader(item.typ):
			prose += "#######"[:headerWeight(item.typ)] + " " + item.val + "\n"
			result.AddHeader(item.val, headerWeight(item.typ))
		case item.typ == itemCodeBlock:
			result.AddBlockParsed(
				model.NewBlockParsed(labels, base.MdProse(prose), base.OpaqueCode(item.val)))
			labels = []base.Label{}
			prose = ""
		}
	}
}
