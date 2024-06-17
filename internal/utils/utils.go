package utils

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	PgmName = "mdrip"
	Version = "v2.0.0-rc02"
)

var leading = regexp.MustCompile("^[0-9]+_")

// DropLeadingNumbers drops leading numbers and underscores.
func DropLeadingNumbers(s string) string {
	r := leading.FindStringIndex(s)
	if r == nil {
		return s
	}
	return s[r[1]:]
}

// SampleString converts a long multi-line string to a short one-line sample.
func SampleString(incoming string, max int) string {
	s := len(incoming)
	if s > max {
		s = max
	}
	return convertBadWhiteSpaceToBlanks(strings.TrimSpace(incoming[:s]))
}

// Summarize is better than SampleString?
func Summarize(c []byte) string {
	const mx = 60
	if len(c) > mx {
		c = c[:mx]
	}
	c = bytes.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, c)
	return string(c)
}

// Convert tabs, newlines, etc. to normal blanks.
func convertBadWhiteSpaceToBlanks(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case 0x000A, 0x000B, 0x000C, 0x000D, 0x0085, 0x2028, 0x2029:
			return ' '
		default:
			return r
		}
	}, s)
}

const blanks = "                                                                "

// Spaces returns a string of length n with only spaces.
func Spaces(n int) string {
	if n < 1 {
		return ""
	}
	// return fmt.Sprintf("%"+strconv.Itoa(n)+"s", " ")
	return blanks[:n]
}

// Check reports the error fatally if it's non-nil.
func Check(msg string, err error) {
	if err != nil {
		panic(fmt.Errorf("%s; %w", msg, err))
	}
}
