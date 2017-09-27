package base

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FilePath string

func (n FilePath) ReadDir() ([]os.FileInfo, error) {
	return ioutil.ReadDir(string(n))
}

func (n FilePath) Base() string {
	arg := string(n)
	ext := filepath.Ext(arg)
	if len(ext) < 1 {
		return filepath.Base(arg)
	}
	return filepath.Base(arg[:strings.Index(arg, ext)])

}
func (n FilePath) Join(info os.FileInfo) FilePath {
	return FilePath(filepath.Join(string(n), info.Name()))
}
func (n FilePath) Read() (string, error) {
	contents, err := ioutil.ReadFile(string(n))
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

// Labels are applied to code blocks to identify them and allow the
// blocks to be grouped into categories, e.g. tests or tutorials.
type Label string

func (l Label) String() string { return string(l) }

const (
	AnyLabel   = Label(`__AnyLabel__`)
	SleepLabel = Label(`sleep`)
)

// OpaqueCode is an opaque, uninterpreted, unknown block of text that
// is presumably shell commands parsed from markdown.  Fed into a
// shell interpreter, the entire thing either succeeds or fails.
type OpaqueCode string

func (c OpaqueCode) String() string { return string(c) }
func (c OpaqueCode) Bytes() []byte  { return []byte(c) }
