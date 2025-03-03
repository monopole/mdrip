package lexer

import (
	"strings"
	"unicode/utf8"
)

type token string

type position int

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input   string     // string being scanned
	state   stateFn    // the next lexing function to enter
	current position   // current position in 'input'
	start   position   // start of this item
	width   position   // width of last rune read
	tCh     chan token // channel of scanned items
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

func (l *lexer) emit() {
	l.tCh <- token(l.input[l.start:l.current])
	l.start = l.current
}

func (l *lexer) ignore() {
	l.start = l.current
}

// newLex creates a new scanner for the input string.
func newLex(input string) *lexer {
	l := &lexer{
		input: input,
		tCh:   make(chan token),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for l.state = lexText; l.state != nil; {
		l.state = l.state(l)
	}
}

// Roughly this reads line by line and changes behavior.
func lexText(l *lexer) stateFn {
	for {
		r := l.next()
		if r == eof {
			l.ignore()
			close(l.tCh)
			return nil
		}
		if isLetterOrDigit(r) {
			l.backup()
			if l.current > l.start {
				l.ignore()
			}
			return lexWord
		}
	}
}

func lexWord(l *lexer) stateFn {
	for isLetterOrDigit(l.next()) {
	}
	l.backup()
	l.emit()
	return lexText
}

func isLetterOrDigit(r rune) bool {
	const lettersAndNumbers = "012345789" +
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return strings.IndexRune(lettersAndNumbers, r) >= 0
}

func gatherAsLowerCase(input string) (res []string) {
	l := newLex(input)
	for wut := range l.tCh {
		res = append(res, strings.ToLower(string(wut)))
	}
	return
}

func MakeIdentifier(input string, limit int) string {
	if limit < 1 {
		return ""
	}
	words := gatherAsLowerCase(input)
	words = dropBadWords(words)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(shorten(words[0]))
	if len(words) == 1 || limit == 1 {
		return b.String()
	}
	limit = limit - 1
	words = words[1:]
	c := 0
	for _, w := range words {
		c++
		b.WriteString(capitalized(shorten(w)))
		if c == limit {
			break
		}
	}
	return b.String()
}

const maxWordSize = 6

func shorten(word string) string {
	if len(word) <= maxWordSize {
		return word
	}
	return word[:maxWordSize]
}

func dropBadWords(words []string) (res []string) {
	first := true
	for _, w := range words {
		if first {
			first = false
			if len(w) > 1 && !isBadFirst(w) {
				// Drop sudo, but keep common linux commands like cp, ls
				res = append(res, w)
			}
		} else {
			if len(w) > 2 {
				res = append(res, w)
			}
		}
	}
	return
}

func isBadFirst(word string) bool {
	return word == "sudo"
}

// capitalized only works on english ascii, which is fine here.
func capitalized(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}
