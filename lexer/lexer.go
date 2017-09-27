// Inspired by golang.org/src/pkg/text/template/parse/lex.go
//
// Cannot use stuct embedding to reuse, since all the good parts are
// private.

package lexer

import (
	"fmt"
	"github.com/monopole/mdrip/base"
	"github.com/monopole/mdrip/model"
	"strings"
	"unicode/utf8"
)

type position int

type itemType int

const (
	itemError         itemType = iota
	itemProse                  // Prose between command blocks.
	itemBlockLabel             // Label for a command block
	itemLabelledBlock          // All lines between codeFence marks
	itemEOF
)

type item struct {
	typ itemType // Type of this item.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ == itemBlockLabel:
		return string(labelMarker) + i.val
	case i.typ == itemLabelledBlock:
		return "--------\n" + i.val + "--------\n"
	case len(i.val) > 10:
		return fmt.Sprintf("%.30s...", i.val)
	}
	return fmt.Sprintf("%s", i.val)
}

const (
	labelMarker  = '@'
	commentOpen  = "<!--"
	commentClose = "-->"
	codeFence    = "```"
	// All punctuation except for <, so we can watch for markdown comments.
	mdPunct           = "!->@#$%^&*()_=+\\|`~{}[];:'`\",.?/ "
	lettersAndNumbers = "012345789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input   string    // string being scanned
	state   stateFn   // the next lexing function to enter
	current position  // current position in 'input'
	start   position  // start of this item
	width   position  // width of last rune read
	items   chan item // channel of scanned items
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
	l.items <- item{t, l.input[l.start:l.current]}
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
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	return item
}

// newLex creates a new scanner for the input string.
func newLex(input string) *lexer {
	l := &lexer{
		input: input,
		items: make(chan item),
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
	return r == ' ' || r == '\t'
}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// lexText scans until an opening comment delimiter.
func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.current:], commentOpen) {
			if l.current > l.start {
				l.emit(itemProse)
			}
			return lexPutativeComment
		}
		if l.next() == eof {
			if l.current > l.start {
				l.emit(itemProse)
			}
			l.ignore()
			l.emit(itemEOF)
			return nil
		}
		l.acceptRun(mdPunct + lettersAndNumbers)
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
			l.acceptRun("_" + lettersAndNumbers)
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
			l.acceptRun(" \t")
			l.ignore()
			r := l.next()
			if r != '\n' && r != '\r' {
				return l.errorf("Expected command block marker at start of line.")
			}
			l.ignore()
			if !strings.HasPrefix(l.input[l.current:], codeFence) {
				return l.errorf("Expected command block mark, got: " + l.input[l.current:])
			}
			return lexLabelledBlock
		}
	}
}

// lexLabelledBlock scans a command block.  Initial marker known to be present.
func lexLabelledBlock(l *lexer) stateFn {
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
				l.emit(itemLabelledBlock)
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

// Parse lexes the incoming string into a list of model.BlockParsed.
func Parse(s string) (result []*model.BlockParsed) {
	result = []*model.BlockParsed{}
	prose := ""
	labels := []base.Label{}
	l := newLex(s)
	for {
		item := l.nextItem()
		switch {
		case item.typ == itemEOF || item.typ == itemError:
			prose = strings.TrimSpace(prose)
			if len(prose) > 0 {
				// Hack to grab the last bit of prose.
				// The data structure returned by Parse needs redesign.
				result = append(result, model.NewBlockParsed(labels, []byte(prose), ""))
			}
			return
		case item.typ == itemBlockLabel:
			labels = append(labels, base.Label(item.val))
		case item.typ == itemProse:
			prose = item.val
		case item.typ == itemLabelledBlock:
			result = append(result, model.NewBlockParsed(labels, []byte(prose), item.val))
			labels = []base.Label{}
			prose = ""
		}
	}
}
