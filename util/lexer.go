// Other than the custom stateFn's, much of this is copied from
// https://golang.org/src/pkg/text/template/parse/lex.go. Cannot use
// stuct embedding to reuse, since all the good parts are private.

package util

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

type Pos int

type CommandBlock struct {
	labels   []string
	codeText string
}

func (x CommandBlock) GetLabels() []string {
	return x.labels
}
func (x CommandBlock) GetCodeText() string {
	return x.codeText
}

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
	case i.typ == itemCommandBlock:
		return "--------\n" + i.val + "--------\n"
	case len(i.val) > 10:
		return fmt.Sprintf("%.30s...", i.val)
	}
	return fmt.Sprintf("%s", i.val)
}

type itemType int

const (
	itemError        itemType = iota
	itemBlockLabel            // Label for a command block
	itemCommandBlock          // All lines between codeFence marks
	itemEOF
)

const (
	labelMarker  = '@'
	commentOpen  = "<!--"
	commentClose = "-->"
	codeFence    = "```\n"
)

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input string    // string being scanned
	state stateFn   // the next lexing function to enter
	pos   Pos       // current position in 'input'
	start Pos       // start of this item
	width Pos       // width of last rune read
	items chan item // channel of scanned items
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) ignore() {
	l.start = l.pos
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

func (l *lexer) acceptWord() {
	l.acceptRun("012345789abcdefghijklmnopqrstuvwxyz_ABCDEFGHIJKLMNOPQRSTUVWXYZ")
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
		if strings.HasPrefix(l.input[l.pos:], commentOpen) {
			return lexPutativeComment
		}
		if l.next() == eof {
			l.ignore()
			l.emit(itemEOF)
			return nil
		}
	}
}

// Move to lexing a command block intended for a particular script, or to
// lexing a simple comment.  Comment opener known to be present.
func lexPutativeComment(l *lexer) stateFn {
	l.pos += Pos(len(commentOpen))
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
	i := strings.Index(l.input[l.pos:], commentClose)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += Pos(i + len(commentClose))
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
			l.acceptWord()
			if l.width == 0 {
				return l.errorf("empty block label")
			}
			l.emit(itemBlockLabel)
		default:
			l.backup()
			if !strings.HasPrefix(l.input[l.pos:], commentClose) {
				return l.errorf("improperly closed block label sequence")
			}
			l.pos += Pos(len(commentClose))
			l.ignore()
			l.acceptRun(" \t")
			l.ignore()
			r := l.next()
			if r != '\n' && r != '\r' {
				return l.errorf("Expected command block marker at start of line.")
			}
			l.ignore()
			if !strings.HasPrefix(l.input[l.pos:], codeFence) {
				return l.errorf("Expected command block mark, got: " + l.input[l.pos:])
			}
			return lexCommandBlock
		}
	}
	return lexText
}

// lexCommandBlock scans a command block.  Initial marker known to be present.
func lexCommandBlock(l *lexer) stateFn {
	l.pos += Pos(len(codeFence))
	l.ignore()
	for {
		if strings.HasPrefix(l.input[l.pos:], codeFence) {
			if l.pos > l.start {
				l.emit(itemCommandBlock)
			}
			l.pos += Pos(len(codeFence))
			l.ignore()
			return lexText
		}
		if l.next() == eof {
			return l.errorf("unclosed command block")
		}
	}
}

func shouldSleep(labels []string) bool {
	for _, label := range labels {
		if label == "sleep" {
			return true
		}
	}
	return false
}

// Parse lexes the incoming string into a mapping from block label to
// CommandBlock array.  The labels are the strings after a labelMarker in
// a comment preceding a command block.  Arrays hold command blocks in the
// order they appeared in the input.
func Parse(s string) (result map[string][]*CommandBlock) {
	result = make(map[string][]*CommandBlock)
	currentLabels := make([]string, 0, 10)
	l := newLex(s)
	for {
		item := l.nextItem()
		switch {
		case item.typ == itemEOF || item.typ == itemError:
			return
		case item.typ == itemBlockLabel:
			currentLabels = append(currentLabels, item.val)
		case item.typ == itemCommandBlock:
			if len(currentLabels) == 0 {
				fmt.Println("Have an unlabelled command block:\n " + item.val)
				os.Exit(1)
			}
			// If the command block has a 'sleep' label, add a brief sleep
			// at the end.  This is hack to give servers placed in the
			// background time to start.
			if shouldSleep(currentLabels) {
				item.val = item.val + "sleep 2s # Added by mdrip\n"
			}
			newBlock := &CommandBlock{currentLabels, item.val}
			for _, label := range currentLabels {
				blocks, ok := result[label]
				if ok {
					blocks = append(blocks, newBlock)
				} else {
					blocks = []*CommandBlock{newBlock}
				}
				result[label] = blocks
			}
			currentLabels = make([]string, 0, 10)
		}
	}
}
