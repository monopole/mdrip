package base

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FilePath holds a file path.
type FilePath string

// ReadDir treats the path as a directory name and reads it.
func (n FilePath) ReadDir() ([]os.FileInfo, error) {
	return ioutil.ReadDir(string(n))
}

// Base returns the last part of the file path, e.g. foo in /usr/local/foo.
func (n FilePath) Base() string {
	arg := string(n)
	ext := filepath.Ext(arg)
	if len(ext) < 1 || ext != ".md" {
		return filepath.Base(arg)
	}
	return filepath.Base(arg[:strings.Index(arg, ext)])

}

// Join joins two file name parts using the OS-specific file path joiner.
func (n FilePath) Join(info os.FileInfo) FilePath {
	return FilePath(filepath.Join(string(n), info.Name()))
}

// Read reads the file into a string.
func (n FilePath) Read() (string, error) {
	contents, err := ioutil.ReadFile(string(n))
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Label is used to select code blocks, and group them into
// categories, e.g. run these blocks under test, run these blocks to do setup, etc.
type Label string

// String form of the label.
func (l Label) String() string { return string(l) }

const (
	// WildCardLabel matches an label.
	WildCardLabel = Label(`__wildcard__`)
	// AnonLabel may be used as a label placeholder when a label is needed but not specified.
	AnonLabel = Label(`__anonymous__`)
	// SleepLabel indicates the author wants a sleep after the block in a test context
	// where there is no natural human caused pause.
	SleepLabel = Label(`sleep`)
)

// OpaqueCode is an opaque, uninterpreted, unknown block of text that
// is presumably shell commands parsed from markdown.  Fed into a
// shell interpreter, the entire thing either succeeds or fails.
type OpaqueCode string

// String form of OpaqueCode.
func (c OpaqueCode) String() string { return string(c) }

// Bytes of the code.
func (c OpaqueCode) Bytes() []byte { return []byte(c) }

// NoCode is a constructor for NoCode - easy to search for usage.
func NoCode() OpaqueCode { return "" }

// NoLabels is easer to read than the literal empty array.
func NoLabels() []Label { return []Label{} }

// MdProse is documentation (plain text or markdown) for OpaqueCode.
type MdProse []byte

// String form of MdProse.
func (x MdProse) String() string { return string(x) }

// Bytes of MdProse.
func (x MdProse) Bytes() []byte { return []byte(x) }

// NoProse is placeholder for no prose.
func NoProse() MdProse { return []byte{} }
