package model

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type TypeSessId string

type FilePath string

func (n FilePath) ReadDir() ([]os.FileInfo, error) {
	return ioutil.ReadDir(string(n))
}

func (n FilePath) Base() string {
	return filepath.Base(string(n))
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

const (
	AnyLabel   = Label(`__AnyLabel__`)
	SleepLabel = Label(`sleep`)
)

func (l Label) String() string { return string(l) }

// OpaqueCode is an opaque, uninterpreted, unknown block of text that
// is presumably shell commands parsed from markdown.  Fed into a
// shell interpreter, the entire thing either succeeds or fails.
type OpaqueCode string

func (c OpaqueCode) String() string { return string(c) }
func (c OpaqueCode) Bytes() []byte  { return []byte(c) }

// Block groups OpaqueCode with its labels.
type Block struct {
	labels []Label
	// prose is presumably human language documentation for the OpaqueCode.
	prose []byte
	code  OpaqueCode
}

func NewBlock(labels []Label, p []byte, c string) *Block {
	return &Block{labels, p, OpaqueCode(c)}
}
func (x *Block) Labels() []Label  { return x.labels }
func (x *Block) Prose() []byte    { return x.prose }
func (x *Block) Code() OpaqueCode { return x.code }
