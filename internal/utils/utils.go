package utils

import (
	"bytes"
	"errors"
	"os"
	"regexp"
	"unicode"
)

const (
	PgmName = "mdrip"

	// AllowDebug is a global flag to hide or expose debugging tricks.
	AllowDebug = false
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

// Summarize a code block in one line.
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

// const blanks = "                                                                "
const blanks = "                                               " +
	"                                               "

// Spaces returns a string of length n with only spaces.
func Spaces(n int) string {
	if n < 1 {
		return ""
	}
	if n > len(blanks) {
		panic("too many blanks")
	}
	return blanks[:n]
}

type PathCase int

const (
	PathInUnknownState PathCase = iota
	PathIsAFile
	PathIsAFolder
	PathDoesNotExist
)

func PathStatus(path string) (PathCase, error) {
	fi, err := os.Stat(path)
	if err == nil {
		// path exists!
		if fi.IsDir() {
			return PathIsAFolder, nil
		}
		return PathIsAFile, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return PathDoesNotExist, nil
	}
	// File may or may not exist, depends on the error.
	return PathInUnknownState, err
}
